package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// random_retarget_cards_test.go — #1815: targets chosen by the engine
// at random. Chef's Kiss reselects a spell's targets (and its copy's)
// from the legal ones other than its controller and their permanents;
// Witch Hunt and Scab-Clan Giant draw a trigger's target as it goes on
// the stack. Nobody is asked, and an undo replays the draw.

const (
	chefsKissOracle      = "3d177c04-9aa3-44b8-bbe4-df5de23a4ce6"
	witchHuntOracle      = "e86bd38f-7804-449d-af29-21e96a56ab30"
	scabClanGiantOracle  = "1c2c2329-fdae-4894-a43f-41fd91cfd2a8"
	knightRampagerOracle = "2fa681b2-ca94-4f0c-83a4-98f00b8e1af8"
)

// Frenzied Rampage draws one opponent at the beginning of combat and
// binds the Knight to attack them this combat; every opponent is
// reachable and the controller never is.
func TestKnightRampagerMustAttackAnOpponentChosenAtRandom(t *testing.T) {
	seen := map[int]bool{}
	for key := byte(0); key < 48 && len(seen) < 3; key++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(key))
		me := g.Seats[0]
		knight := b12Push(g, me.ID, "Knight Rampager", "Artifact Creature — Knight", knightRampagerOracle, 6, 5)
		advanceTo(t, g, game.StepBeginCombat)
		passPriorityAroundTable(t, g)
		var must []uuid.UUID
		g.ReadSnapshot(func() {
			c, _ := g.LookupCardForEffect(knight)
			for _, r := range c.Effective().AttackRequirements {
				if r.MustAttack != uuid.Nil {
					must = append(must, r.MustAttack)
				}
			}
		})
		if len(must) != 1 || must[0] == me.ID {
			t.Fatalf("key %d: the Knight must attack %v; want exactly one opponent", key, must)
		}
		seen[seatOfPlayer(g, must[0])] = true
	}
	if len(seen) != 3 {
		t.Errorf("48 draws reached %d of the 3 opponents", len(seen))
	}
}

// The death trigger's target is an opponent drawn as it goes on the
// stack, and the Knight deals that opponent 4.
func TestKnightRampagerDiesAndHitsAnOpponentChosenAtRandom(t *testing.T) {
	seen := map[int]bool{}
	for key := byte(0); key < 48 && len(seen) < 3; key++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(key))
		me := g.Seats[0]
		knight := b12Push(g, me.ID, "Knight Rampager", "Artifact Creature — Knight", knightRampagerOracle, 6, 5)
		advanceToMain(t, g)
		life := map[int]int{}
		for i, p := range g.Seats {
			life[i] = p.Life
		}
		castCatalogSpell(t, g, "Doom Blade", "Instant", doomBladeOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: knight}})
		passPriorityAroundTable(t, g)
		if p := latestPickTarget(g, me.ID); p != nil {
			t.Fatalf("a target chosen at random asks nobody: %+v", p)
		}
		hit := -1
		for i, p := range g.Seats {
			if p.Life == life[i]-4 {
				if hit != -1 {
					t.Fatalf("key %d: two players took 4", key)
				}
				hit = i
			} else if p.Life != life[i] {
				t.Fatalf("key %d: seat %d life %d -> %d", key, i, life[i], p.Life)
			}
		}
		if hit <= 0 {
			t.Fatalf("key %d: the Knight's death hit seat %d; want one opponent", key, hit)
		}
		seen[hit] = true
	}
	if len(seen) != 3 {
		t.Errorf("48 draws reached %d of the 3 opponents", len(seen))
	}
}

// kissGame casts `spell` for seat 0 at `targets` and has seat 1 answer
// it with Chef's Kiss, under the RNG key `key`. Returns the game, the
// spell and the Kiss, with the Kiss on top of the stack.
func kissGame(t *testing.T, key byte, setup func(g *game.Game), name, typeLine, oracle string, targets func(g *game.Game) []game.TargetRef) (*game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	g.SetRNGKeyForTest(randomModesKey(key))
	if setup != nil {
		setup(g)
	}
	spell := castCatalogSpell(t, g, name, typeLine, oracle, targets(g))
	kiss := respondWith(t, g, g.Seats[1], game.Card{Name: "Chef's Kiss", TypeLine: "Instant", OracleID: chefsKissOracle},
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}}})
	return g, spell, kiss
}

// resolveKiss passes priority until the Kiss has left the stack, and
// no further: the spell and its copy are still there to look at.
func resolveKiss(t *testing.T, g *game.Game, kiss uuid.UUID) {
	t.Helper()
	for i := 0; i < 16 && g.Stack.Contains(kiss); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if g.Stack.Contains(kiss) {
		t.Fatal("Chef's Kiss never resolved")
	}
}

// kissCopyOf is the copy Chef's Kiss made: the one copy on the stack.
func kissCopyOf(t *testing.T, g *game.Game) *game.StackItem {
	t.Helper()
	var out *game.StackItem
	for _, it := range g.StackMeta {
		if it != nil && it.IsCopy {
			if out != nil {
				t.Fatal("more than one copy on the stack")
			}
			out = it
		}
	}
	if out == nil {
		t.Fatal("Chef's Kiss made no copy")
	}
	return out
}

func soleTarget(t *testing.T, item *game.StackItem) game.TargetRef {
	t.Helper()
	var out []game.TargetRef
	for _, r := range item.Targets {
		if r.Kind == game.TargetCard || r.Kind == game.TargetPlayer {
			out = append(out, r)
		}
	}
	if len(out) != 1 {
		t.Fatalf("want one target, got %+v", item.Targets)
	}
	return out[0]
}

// seatOfPlayer is the seat index of player `id`, or -1.
func seatOfPlayer(g *game.Game, id uuid.UUID) int {
	for i, p := range g.Seats {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func boltAtSeat1(g *game.Game) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}
}

// A Lightning Bolt at the Kiss's controller: the Kiss steals it, copies
// it, and both are pointed at random at one of the three other players
// — never at the Kiss's controller, and every one of them reachable.
func TestChefsKissStealsCopiesAndReselectsAtRandom(t *testing.T) {
	seen := map[int]bool{}
	for key := byte(0); key < 48 && len(seen) < 3; key++ {
		g, bolt, kiss := kissGame(t, key, nil, "Lightning Bolt", "Instant", lightningBoltOracle, boltAtSeat1)
		you := g.Seats[1].ID
		resolveKiss(t, g, kiss)
		if p := latestChoiceOfKindFor(g, game.PendingChoiceRetarget, you); p != nil {
			t.Fatalf("a random reselection asks nobody: %+v", p)
		}
		if got := stackItemControllerOf(g, bolt); got != you {
			t.Fatalf("key %d: the Bolt's controller is %s, want the Kiss's %s", key, got, you)
		}
		cp := kissCopyOf(t, g)
		if cp.Controller != you {
			t.Errorf("key %d: the copy's controller is %s, want %s", key, cp.Controller, you)
		}
		for _, it := range []*game.StackItem{g.StackMeta[bolt], cp} {
			tgt := soleTarget(t, it)
			if tgt.Kind != game.TargetPlayer || tgt.ID == you {
				t.Fatalf("key %d: a new target is %+v; it can't be you", key, tgt)
			}
			seen[seatOfPlayer(g, tgt.ID)] = true
		}
	}
	if len(seen) != 3 {
		t.Errorf("48 draws reached %d of the 3 other players; the pick is uniform over them", len(seen))
	}
}

// "Reselect" draws over every legal target, the current one included
// (Grip of Chaos's "(Select from among all legal targets.)"): a Bolt at
// a third player can be reselected onto that same player.
func TestChefsKissCanRedrawTheCurrentTarget(t *testing.T) {
	redrawn := false
	for key := byte(0); key < 48 && !redrawn; key++ {
		g, bolt, kiss := kissGame(t, key, nil, "Lightning Bolt", "Instant", lightningBoltOracle, func(g *game.Game) []game.TargetRef {
			return []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[2].ID}}
		})
		resolveKiss(t, g, kiss)
		if soleTarget(t, g.StackMeta[bolt]).ID == g.Seats[2].ID {
			redrawn = true
		}
	}
	if !redrawn {
		t.Error("48 draws never landed on the Bolt's current target; a reselection draws over all legal targets")
	}
}

// "Or a permanent you control": a Doom Blade at the Kiss controller's
// creature is moved to another player's creature, never to one of the
// Kiss controller's, and its own creature lives.
func TestChefsKissNeverPicksAPermanentYouControl(t *testing.T) {
	seen := map[string]bool{}
	for key := byte(0); key < 48 && len(seen) < 2; key++ {
		var mine, mine2, other0, other2 uuid.UUID
		setup := func(g *game.Game) {
			other0 = b12Creature(g, g.Seats[0].ID, "Bear Zero", "Creature — Bear", 2, 2)
			mine = b12Creature(g, g.Seats[1].ID, "Bear One", "Creature — Bear", 2, 2)
			mine2 = b12Creature(g, g.Seats[1].ID, "Bear One B", "Creature — Bear", 2, 2)
			other2 = b12Creature(g, g.Seats[2].ID, "Bear Two", "Creature — Bear", 2, 2)
		}
		g, blade, kiss := kissGame(t, key, setup, "Doom Blade", "Instant", doomBladeOracle, func(*game.Game) []game.TargetRef {
			return []game.TargetRef{{Kind: game.TargetCard, ID: mine}}
		})
		resolveKiss(t, g, kiss)
		for _, it := range []*game.StackItem{g.StackMeta[blade], kissCopyOf(t, g)} {
			tgt := soleTarget(t, it)
			if tgt.ID != other0 && tgt.ID != other2 {
				t.Fatalf("key %d: a new target is %s; only Bear Zero and Bear Two qualify", key, tgt.ID)
			}
			seen[b12Card(t, g, tgt.ID).Name] = true
		}
		passPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(mine2) {
			t.Errorf("key %d: a creature the Kiss's controller controls was destroyed", key)
		}
	}
	if len(seen) != 2 {
		t.Errorf("48 draws reached %d of the 2 creatures that qualify", len(seen))
	}
}

// With no legal target but you and your permanents, the targets stay,
// even illegal ones (the rulings): Blossoming Defense's "target
// creature you control" is the Kiss controller's now, every creature
// it could name is excluded, and the spell and the copy keep the
// caster's creature, which is no longer a legal target, so both do
// nothing.
func TestChefsKissKeepsTargetsWhenNothingElseCanBeSelected(t *testing.T) {
	var bear0 uuid.UUID
	setup := func(g *game.Game) {
		bear0 = b12Creature(g, g.Seats[0].ID, "Bear Zero", "Creature — Bear", 2, 2)
		b12Creature(g, g.Seats[1].ID, "Bear One", "Creature — Bear", 2, 2)
	}
	g, def, kiss := kissGame(t, 0, setup, "Blossoming Defense", "Instant", blossomingDefenseOracle, func(*game.Game) []game.TargetRef {
		return []game.TargetRef{{Kind: game.TargetCard, ID: bear0}}
	})
	resolveKiss(t, g, kiss)
	for _, it := range []*game.StackItem{g.StackMeta[def], kissCopyOf(t, g)} {
		if tgt := soleTarget(t, it); tgt.ID != bear0 {
			t.Fatalf("the target moved to %s with nothing that could be selected", tgt.ID)
		}
	}
	kept := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventOptionChosen && ev.CardID == kiss && strings.Contains(ev.Label, "keeps its targets") {
			kept = true
		}
	}
	if !kept {
		t.Error("the log does not say the spell kept its targets")
	}
	passPriorityAroundTable(t, g)
	if p := b12Card(t, g, bear0).CurrentPower(); p != 2 {
		t.Errorf("Bear Zero has power %d; both spells target a creature their controller doesn't control and do nothing", p)
	}
}

// The draw is the keyed stream: an undo of the Kiss's resolution and a
// second resolution points the spell and the copy where the first did.
func TestChefsKissReplaysTheDrawAfterUndo(t *testing.T) {
	for key := byte(0); key < 6; key++ {
		g, bolt, kiss := kissGame(t, key, nil, "Lightning Bolt", "Instant", lightningBoltOracle, boltAtSeat1)
		pre := g.Clone()
		resolve := func() (game.TargetRef, game.TargetRef) {
			resolveKiss(t, g, kiss)
			return soleTarget(t, g.StackMeta[bolt]), soleTarget(t, kissCopyOf(t, g))
		}
		a1, b1 := resolve()
		g.WithWriteLock(func() { g.RestoreFrom(pre) })
		a2, b2 := resolve()
		if a1.ID != a2.ID || b1.ID != b2.ID {
			t.Errorf("key %d: undo then redo drew %s/%s then %s/%s", key, a1.ID, b1.ID, a2.ID, b2.ID)
		}
	}
}

// CR 115.9c: "targets only a single permanent or player" counts the
// DIFFERENT objects and players chosen, then checks that one's current
// state (CR 115.9b).
func TestTargetsOnlyASinglePermanentOrPlayer(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
	bear2 := b12Creature(g, g.Seats[1].ID, "Bear Two", "Creature — Bear", 2, 2)
	gone := uuid.New()
	g.WithWriteLock(func() {
		g.Seats[0].Graveyard.PushTop(game.Card{InstanceID: gone, Name: "Dead Bear", TypeLine: "Creature — Bear",
			Owner: g.Seats[0].ID, Controller: g.Seats[0].ID})
	})
	card := func(id uuid.UUID, slot int) game.TargetRef {
		return game.TargetRef{Kind: game.TargetCard, ID: id, Slot: slot}
	}
	player := func(p *game.Player) game.TargetRef { return game.TargetRef{Kind: game.TargetPlayer, ID: p.ID} }
	cases := []struct {
		name    string
		targets []game.TargetRef
		want    bool
	}{
		{"one player", []game.TargetRef{player(g.Seats[2])}, true},
		{"one permanent", []game.TargetRef{card(bear, 0)}, true},
		{"one permanent twice", []game.TargetRef{card(bear, 0), card(bear, 1)}, true},
		{"two permanents", []game.TargetRef{card(bear, 0), card(bear2, 1)}, false},
		{"a card in a graveyard", []game.TargetRef{card(gone, 0)}, false},
		{"no target", nil, false},
	}
	pred := TargetsOnlyASinglePermanentOrPlayer()
	g.WithWriteLock(func() {
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*game.StackItem{}
		}
	})
	for _, tc := range cases {
		id := uuid.New()
		c := game.Card{InstanceID: id, Name: "Probe", TypeLine: "Instant", Owner: g.Seats[3].ID, Controller: g.Seats[3].ID}
		var got bool
		g.WithWriteLock(func() {
			g.Stack.PushTop(c)
			g.StackMeta[id] = &game.StackItem{ID: id, Kind: game.StackItemSpell, Controller: g.Seats[3].ID,
				SourceCardID: id, Targets: tc.targets}
			got = pred(g, g.Seats[1].ID, c)
			delete(g.StackMeta, id)
			_, _ = g.Stack.Remove(id)
		})
		if got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
	g.Seats[2].Eliminated = true
	id := uuid.New()
	c := game.Card{InstanceID: id, Name: "Probe", TypeLine: "Instant", Owner: g.Seats[3].ID, Controller: g.Seats[3].ID}
	g.WithWriteLock(func() {
		g.Stack.PushTop(c)
		g.StackMeta[id] = &game.StackItem{ID: id, Kind: game.StackItemSpell, Controller: g.Seats[3].ID,
			SourceCardID: id, Targets: []game.TargetRef{player(g.Seats[2])}}
		if pred(g, g.Seats[1].ID, c) {
			t.Error("a player who has left the game is ignored (CR 115.9b)")
		}
	})
}

// Witch Hunt's end-step trigger draws its target among its controller's
// opponents as it goes on the stack, and that opponent gets it.
func TestWitchHuntGoesToAnOpponentChosenAtRandom(t *testing.T) {
	seen := map[int]bool{}
	for key := byte(0); key < 48 && len(seen) < 3; key++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(key))
		me := g.Seats[0]
		wh := b12Push(g, me.ID, "Witch Hunt", "Enchantment", witchHuntOracle, 0, 0)
		advanceToEndStepOf(t, g, 0)
		if p := latestPickTarget(g, me.ID); p != nil {
			t.Fatalf("a target chosen at random asks nobody: %+v", p)
		}
		item := triggerOnStack(g, wh)
		if item == nil {
			t.Fatal("Witch Hunt's end-step trigger is not on the stack")
		}
		tgt := soleTarget(t, item)
		if tgt.Kind != game.TargetPlayer || tgt.ID == me.ID {
			t.Fatalf("key %d: target %+v; it must be an opponent", key, tgt)
		}
		seen[seatOfPlayer(g, tgt.ID)] = true
		passPriorityAroundTable(t, g)
		if got := b12Card(t, g, wh).Controller; got != tgt.ID {
			t.Errorf("key %d: Witch Hunt is controlled by %s, want the drawn opponent %s", key, got, tgt.ID)
		}
	}
	if len(seen) != 3 {
		t.Errorf("48 draws reached %d of the 3 opponents", len(seen))
	}
}

func TestWitchHuntDamagesItsControllerAndStopsLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	b12Push(g, owner.ID, "Witch Hunt", "Enchantment", witchHuntOracle, 0, 0)
	advanceToUpkeepOfSeat(t, g, 1)
	life := owner.Life
	passPriorityAroundTable(t, g)
	if owner.Life != life-4 {
		t.Errorf("life %d after the upkeep trigger, want %d", owner.Life, life-4)
	}
	other := g.Seats[3]
	before := other.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, other.ID, 5) })
	if other.Life != before {
		t.Errorf("a player gained life under Witch Hunt: %d -> %d", before, other.Life)
	}
}

// Scab-Clan Giant fights a creature an opponent controls drawn at
// random, never one of its controller's.
func TestScabClanGiantFightsAnOpposingCreatureChosenAtRandom(t *testing.T) {
	seen := map[string]bool{}
	for key := byte(0); key < 48 && len(seen) < 2; key++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(key))
		me := g.Seats[0]
		elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 6)
		bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 6)
		ogre := b12Creature(g, g.Seats[2].ID, "Ogre", "Creature — Ogre", 3, 6)
		giant := castCreatureCard(t, g, "Scab-Clan Giant", "Creature — Giant Warrior", scabClanGiantOracle, 4, 5)
		if p := latestPickTarget(g, me.ID); p != nil {
			t.Fatalf("a target chosen at random asks nobody: %+v", p)
		}
		if d := b12Card(t, g, elf).DamageMarked; d != 0 {
			t.Fatalf("key %d: the Giant fought its controller's own creature", key)
		}
		hitBear, hitOgre := b12Card(t, g, bear).DamageMarked == 4, b12Card(t, g, ogre).DamageMarked == 4
		took := b12Card(t, g, giant).DamageMarked
		switch {
		case hitBear && !hitOgre && took == 2:
			seen["bear"] = true
		case hitOgre && !hitBear && took == 3:
			seen["ogre"] = true
		default:
			t.Fatalf("key %d: one fight with one opposing creature; bear hit %v, ogre hit %v, giant took %d", key, hitBear, hitOgre, took)
		}
	}
	if len(seen) != 2 {
		t.Errorf("48 draws reached %d of the 2 opposing creatures", len(seen))
	}
}

func TestScabClanGiantWithNoOpposingCreatureDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 6)
	giant := castCreatureCard(t, g, "Scab-Clan Giant", "Creature — Giant Warrior", scabClanGiantOracle, 4, 5)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatalf("no opposing creature, yet a prompt: %+v", p)
	}
	if b12Card(t, g, elf).DamageMarked != 0 || b12Card(t, g, giant).DamageMarked != 0 {
		t.Error("the Giant fought with no creature an opponent controls")
	}
}

// Only a trigger's or an activated ability's target is drawn by the
// engine, only one, and on an activated ability only as its lone clause.
func TestRegisterRefusesATargetChosenAtRandomOffATrigger(t *testing.T) {
	expectRegisterPanic(t, "only a triggered or an activated ability's target", func() {
		Register(Spec{OracleID: "random-target-spell", Name: "random-target-spell",
			Targets: ChosenAtRandom(TargetAny())})
	})
	expectRegisterPanic(t, "only a lone random clause", func() {
		Register(Spec{OracleID: "random-target-activated-two", Name: "random-target-activated-two",
			Activated: []ActivatedAbility{{Label: "{T}: Probe.", Cost: TapCost(),
				Targets: Clauses(ChosenAtRandom(TargetAny()), TargetCreature("target creature")),
				Effect:  func(*game.Game, *game.StackItem) error { return nil }}}})
	})
	expectRegisterPanic(t, "only a triggered or an activated ability's target", func() {
		Register(Spec{OracleID: "random-target-activated-mode", Name: "random-target-activated-mode",
			Activated: []ActivatedAbility{{Label: "{T}: Choose one.", Cost: TapCost(),
				Modes:  ChooseOne(Mode("Deal 1 damage to any target chosen at random.", ChosenAtRandom(TargetAny())), Mode("Draw a card.")),
				Effect: func(*game.Game, *game.StackItem) error { return nil }}}})
	})
	expectRegisterPanic(t, "not exactly one undivided target", func() {
		Register(Spec{OracleID: "random-target-two", Name: "random-target-two",
			Triggered: []game.TriggeredAbility{{
				Watches: []game.EventKind{game.EventETB}, AppliesTo: Self, Key: "probe",
				Targets: ChosenAtRandom(TargetCreature("two target creatures chosen at random").WithCount(2, 2)),
				Effect:  func(*game.Game, *game.StackItem) error { return nil },
			}}})
	})
}
