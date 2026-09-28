package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// monstrosity_test.go — #1700, CR 701.37 (ADR 0071 amendment
// 2026-09-28). The keyword action's rules first, through real
// activations on real cards, then one test per card.

const (
	polukranosOracle       = "ecbf5560-853a-43a9-bf70-ac8d403b0ce8"
	stormbreathOracle      = "7a770865-9383-45aa-883f-21ee649d1ea3"
	hundredHandedOneOracle = "954aaaa7-c3c1-4696-9ace-bcb5359b1709"
	emberSwallowerOracle   = "0cdc86c4-8c68-4ba6-8165-511f32156a4c"
	arborColossusOracle    = "bec585af-0c59-46a7-838f-d747c80a9e97"
	fleecemaneLionOracle   = "e3c8cdf7-a26a-45eb-9498-99b618cfba99"
	domesticatedHydraOrcl  = "b6d65fa3-26a9-44b2-a93d-a295bd080729"
	hydraBroodmasterOracle = "9e4fb4b1-9b28-4a49-b4ac-d5549c08594d"
)

func pushMonster(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Monster", OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// activateMonstrosity announces the card's one activated ability
// (every card here prints exactly one) with X = x.
func activateMonstrosity(t *testing.T, g *game.Game, who, id uuid.UUID, x int) {
	t.Helper()
	if err := g.ActivateCatalogAbility(who, id, 0, game.ActivateAbilityParams{XValue: x}); err != nil {
		t.Fatalf("activate monstrosity: %v", err)
	}
}

func monstrousNow(g *game.Game, id uuid.UUID) bool {
	var m bool
	g.ReadSnapshot(func() { m = g.IsMonstrous(id) })
	return m
}

// becameMonstrousEvents returns the Amount of every EventBecameMonstrous
// the game emitted for `id`.
func becameMonstrousEvents(g *game.Game, id uuid.UUID) []int {
	var out []int
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventBecameMonstrous && ev.CardID == id {
				out = append(out, ev.Amount)
			}
		}
	})
	return out
}

// activationsOnStackFrom counts the activated abilities from
// `source` waiting on the stack.
func activationsOnStackFrom(g *game.Game, source uuid.UUID) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemActivated && item.SourceCardID == source {
			n++
		}
	}
	return n
}

func giveHandCards(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

// --- the keyword action ------------------------------------------------

// Activating once puts the counters on, sets the designation, and
// fires the becomes-monstrous trigger with X on its event.
func TestMonstrosityPlacesCountersSetsTheFlagAndTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := pushMonster(g, me.ID, "Stormbreath Dragon", stormbreathOracle, 4, 4)
	giveHandCards(opp, 4)
	hand := len(opp.Hand.Cards)
	life := opp.Life

	activateMonstrosity(t, g, me.ID, dragon, 0)
	passPriorityAroundTable(t, g)

	if n := plusOneCounters(g, dragon); n != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", n)
	}
	if !monstrousNow(g, dragon) {
		t.Error("not monstrous after Monstrosity 3 resolved")
	}
	if got := becameMonstrousEvents(g, dragon); len(got) != 1 || got[0] != 3 {
		t.Errorf("EventBecameMonstrous amounts = %v, want [3]", got)
	}
	if opp.Life != life-hand {
		t.Errorf("opponent life %d → %d, want -%d (the cards in their hand)", life, opp.Life, hand)
	}
	if !e2Card(t, g, dragon).Monstrous {
		t.Error("Card.Monstrous not set")
	}
}

// CR 701.37a: a second activation is legal and does nothing — no
// counters, no event, no trigger.
func TestMonstrosityOnAMonstrousCreatureDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := pushMonster(g, me.ID, "Stormbreath Dragon", stormbreathOracle, 4, 4)
	giveHandCards(opp, 2)
	hand := len(opp.Hand.Cards)
	life := opp.Life

	for i := 0; i < 2; i++ {
		activateMonstrosity(t, g, me.ID, dragon, 0)
		passPriorityAroundTable(t, g)
	}

	if n := plusOneCounters(g, dragon); n != 3 {
		t.Errorf("+1/+1 counters = %d after two activations, want 3", n)
	}
	if got := becameMonstrousEvents(g, dragon); len(got) != 1 {
		t.Errorf("EventBecameMonstrous emitted %d times, want 1", len(got))
	}
	if opp.Life != life-hand {
		t.Errorf("opponent life %d → %d, want -%d once (the trigger fired twice?)", life, opp.Life, hand)
	}
}

// CR 701.37a's "if this creature isn't monstrous" is checked as the
// ability RESOLVES: a second activation in response to the first finds
// the creature monstrous by the time it resolves, and does nothing.
func TestMonstrosityInResponseToItselfResolvesToNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := pushMonster(g, me.ID, "Stormbreath Dragon", stormbreathOracle, 4, 4)
	giveHandCards(opp, 3)
	hand := len(opp.Hand.Cards)
	life := opp.Life

	activateMonstrosity(t, g, me.ID, dragon, 0)
	activateMonstrosity(t, g, me.ID, dragon, 0)
	if n := activationsOnStackFrom(g, dragon); n != 2 {
		t.Fatalf("activations on the stack = %d, want both", n)
	}
	passPriorityAroundTable(t, g)

	if n := plusOneCounters(g, dragon); n != 3 {
		t.Errorf("+1/+1 counters = %d, want 3 — the second resolution placed counters on a monstrous creature", n)
	}
	if got := becameMonstrousEvents(g, dragon); len(got) != 1 {
		t.Errorf("EventBecameMonstrous emitted %d times, want 1", len(got))
	}
	if opp.Life != life-hand {
		t.Errorf("opponent life %d → %d, want -%d once", life, opp.Life, hand)
	}
}

// The counters ride the CR 614 pipeline: Doubling Season doubles them.
// X on the trigger stays the announced X (Polukranos's and Hydra
// Broodmaster's rulings), so the tokens are 2/2 and there are two of
// them — times two again from Doubling Season's token half.
func TestMonstrosityDoublingSeasonDoublesCountersNotX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	brood := pushMonster(g, me.ID, "Hydra Broodmaster", hydraBroodmasterOracle, 7, 7)

	activateMonstrosity(t, g, me.ID, brood, 2)
	passPriorityAroundTable(t, g)

	if n := plusOneCounters(g, brood); n != 4 {
		t.Errorf("+1/+1 counters = %d, want 4 (2 doubled)", n)
	}
	if got := becameMonstrousEvents(g, brood); len(got) != 1 || got[0] != 2 {
		t.Errorf("EventBecameMonstrous amounts = %v, want [2] — the announced X, not the counters that landed", got)
	}
	hydras := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Hydra" && c.Controller == me.ID {
			hydras++
			if c.Power != 2 || c.Toughness != 2 {
				t.Errorf("Hydra token is %d/%d, want 2/2", c.Power, c.Toughness)
			}
		}
	}
	if hydras != 4 {
		t.Errorf("Hydra tokens = %d, want 4 (X=2, doubled by Doubling Season's token half)", hydras)
	}
}

// Two different counter replacements pause the placement on the CR 616
// ordering prompt. The creature is not monstrous — and nothing has
// triggered — until the counters land.
func TestMonstrosityWaitsForAPausedPlacement(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hs := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	dragon := pushMonster(g, me.ID, "Stormbreath Dragon", stormbreathOracle, 4, 4)
	giveHandCards(opp, 1)

	activateMonstrosity(t, g, me.ID, dragon, 0)
	passPriorityAroundTable(t, g)
	if monstrousNow(g, dragon) {
		t.Fatal("monstrous while the counters are still owed to the CR 616 prompt")
	}

	answerOrderBySource(t, g, ds, hs) // 3 → 6 → 7
	if n := plusOneCounters(g, dragon); n != 7 {
		t.Errorf("+1/+1 counters = %d, want 7 ((3*2)+1)", n)
	}
	if !monstrousNow(g, dragon) {
		t.Fatal("not monstrous after the placement settled")
	}
	if got := becameMonstrousEvents(g, dragon); len(got) != 1 || got[0] != 3 {
		t.Errorf("EventBecameMonstrous amounts = %v, want [3]", got)
	}
}

// CR 400.7 + CR 701.37b: a flickered monster is a new object, not
// monstrous, and can become monstrous again.
func TestMonstrousClearsOnFlicker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lion := pushMonster(g, me.ID, "Fleecemane Lion", fleecemaneLionOracle, 3, 3)
	activateMonstrosity(t, g, me.ID, lion, 0)
	passPriorityAroundTable(t, g)
	if !monstrousNow(g, lion) {
		t.Fatal("setup: not monstrous")
	}

	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()}
		if err := (Flicker{Target: lion}).Apply(ctxFor(g, item)); err != nil {
			t.Fatalf("Flicker: %v", err)
		}
	})
	back := findBattlefieldByName(g, "Fleecemane Lion")
	if back == uuid.Nil {
		t.Fatal("the Lion did not come back")
	}
	if monstrousNow(g, back) || e2Card(t, g, back).Monstrous {
		t.Error("the flickered Lion is still monstrous")
	}
	if hasString(effectiveAbilities(t, g, back), "hexproof") {
		t.Error("the flickered Lion kept hexproof")
	}

	activateMonstrosity(t, g, me.ID, back, 0)
	passPriorityAroundTable(t, g)
	if !monstrousNow(g, back) || plusOneCounters(g, back) != 1 {
		t.Error("the new object could not become monstrous")
	}
}

// CR 701.37b: monstrous is not a copiable value. A Clone of a
// monstrous Fleecemane Lion is an ordinary 3/3 with no counters, and
// its gated statics are off.
func TestACopyOfAMonstrousCreatureIsNotMonstrous(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	lion := pushMonster(g, active.ID, "Fleecemane Lion", fleecemaneLionOracle, 3, 3)
	activateMonstrosity(t, g, active.ID, lion, 0)
	passPriorityAroundTable(t, g)
	if !monstrousNow(g, lion) {
		t.Fatal("setup: not monstrous")
	}

	cloneID := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
	resolveWithCopyChoice(t, g, lion)

	c := e2Card(t, g, cloneID)
	if c.Name != "Fleecemane Lion" {
		t.Fatalf("clone name = %q", c.Name)
	}
	if c.Monstrous {
		t.Error("the copy is monstrous")
	}
	if c.Counters[game.CounterPlusOne] != 0 {
		t.Error("the copy took the counters")
	}
	if abs := effectiveAbilities(t, g, cloneID); hasString(abs, "hexproof") || hasString(abs, "indestructible") {
		t.Errorf("the copy has the monstrous-gated keywords: %v", abs)
	}
	if !hasString(effectiveAbilities(t, g, lion), "hexproof") {
		t.Error("the original lost hexproof")
	}
}

// Undo and the snapshot: the designation is carried, and an undo
// across the resolution takes it back.
func TestMonstrousSurvivesSnapshotAndUndo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lion := pushMonster(g, me.ID, "Fleecemane Lion", fleecemaneLionOracle, 3, 3)
	pre := g.Clone()

	activateMonstrosity(t, g, me.ID, lion, 0)
	passPriorityAroundTable(t, g)

	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if !monstrousNow(restored, lion) {
		t.Error("the snapshot round-trip dropped the monstrous designation")
	}
	if cl := g.Clone(); !monstrousNow(cl, lion) {
		t.Error("Clone dropped the monstrous designation")
	}

	g.WithWriteLock(func() { g.RestoreFrom(pre) })
	if monstrousNow(g, lion) {
		t.Error("still monstrous after undoing the activation")
	}
	if hasString(effectiveAbilities(t, g, lion), "hexproof") {
		t.Error("hexproof survived the undo")
	}
}

// --- the cards ---------------------------------------------------------

// Polukranos: X damage divided among opponents' creatures, and each
// target hits back for its power. A creature of mine is not offered.
func TestPolukranosDividesXAndTakesDamageBack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	polu := pushMonster(g, me.ID, "Polukranos, World Eater", polukranosOracle, 5, 5)
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 2, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 1, 30)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Wall", 1, 30)

	activateMonstrosity(t, g, me.ID, polu, 3)
	passPriorityAroundTable(t, g)

	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the becomes-monstrous trigger asked for no targets")
	}
	if hasID(p.PickTargetCards, mine) {
		t.Error("a creature you control is offered")
	}
	if hasID(p.PickTargetCards, polu) {
		t.Error("Polukranos is offered to its own trigger")
	}
	b17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 1, b: 2}, a, b)
	passPriorityAroundTable(t, g)

	if d1, d2 := e2Card(t, g, a).DamageMarked, e2Card(t, g, b).DamageMarked; d1 != 1 || d2 != 2 {
		t.Errorf("division landed %d/%d, want 1/2", d1, d2)
	}
	if d := e2Card(t, g, polu).DamageMarked; d != 3 {
		t.Errorf("Polukranos took %d, want 3 (powers 2 + 1)", d)
	}
	if n := plusOneCounters(g, polu); n != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", n)
	}
}

// Every target must be assigned at least 1, so X caps the target count.
func TestPolukranosCannotTargetMoreThanX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	polu := pushMonster(g, me.ID, "Polukranos, World Eater", polukranosOracle, 5, 5)
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)

	activateMonstrosity(t, g, me.ID, polu, 1)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	refs := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{a: 1, b: 1}); err == nil {
		t.Error("two targets accepted for X = 1")
	}
}

// X = 0: monstrous, but the trigger targets nothing and nothing hits back.
func TestPolukranosAtXZeroTargetsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	polu := pushMonster(g, me.ID, "Polukranos, World Eater", polukranosOracle, 5, 5)
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 4, 30)

	activateMonstrosity(t, g, me.ID, polu, 0)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("X = 0 still asked for targets")
	}
	if !monstrousNow(g, polu) {
		t.Error("X = 0 did not make Polukranos monstrous")
	}
	if e2Card(t, g, a).DamageMarked != 0 || e2Card(t, g, polu).DamageMarked != 0 {
		t.Error("damage was dealt at X = 0")
	}
}

func TestStormbreathDragonKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dragon := pushMonster(g, me.ID, "Stormbreath Dragon", stormbreathOracle, 4, 4)
	abs := effectiveAbilities(t, g, dragon)
	for _, kw := range []string{"flying", "haste", "protection from white"} {
		if !hasString(abs, kw) {
			t.Errorf("missing %q: %v", kw, abs)
		}
	}
}

// Reach arrives with the designation; before it, the Giant has only
// vigilance.
func TestHundredHandedOneGainsReachWhenMonstrous(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	giant := pushMonster(g, me.ID, "Hundred-Handed One", hundredHandedOneOracle, 3, 5)
	if abs := effectiveAbilities(t, g, giant); hasString(abs, "reach") || !hasString(abs, "vigilance") {
		t.Fatalf("before monstrosity: %v, want vigilance and no reach", abs)
	}
	activateMonstrosity(t, g, me.ID, giant, 0)
	passPriorityAroundTable(t, g)
	if !hasString(effectiveAbilities(t, g, giant), "reach") {
		t.Error("no reach once monstrous")
	}
	if c := e2Card(t, g, giant); c.CurrentPower() != 6 || c.CurrentToughness() != 8 {
		p, tg := c.CurrentPower(), c.CurrentToughness()
		t.Errorf("P/T = %d/%d, want 6/8", p, tg)
	}
}

// Each player — you too — sacrifices three lands; a nonland is never
// offered, and a player with two lands loses two.
func TestEmberSwallowerEachPlayerSacrificesThreeLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	swallower := pushMonster(g, me.ID, "Ember Swallower", emberSwallowerOracle, 4, 5)
	for i := 0; i < 4; i++ {
		pushLand(g, me.ID, "My Island")
	}
	pushLand(g, opp.ID, "Their Island")
	pushLand(g, opp.ID, "Their Island")
	theirBear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	activateMonstrosity(t, g, me.ID, swallower, 0)
	passPriorityAroundTable(t, g)

	for i := 0; i < 16; i++ {
		c := sacrificeChoiceFor(g, me.ID)
		if c == nil {
			c = sacrificeChoiceFor(g, opp.ID)
		}
		if c == nil {
			break
		}
		if hasID(c.SacrificeOptions, swallower) || hasID(c.SacrificeOptions, theirBear) {
			t.Fatal("a nonland permanent is offered to a land edict")
		}
		if err := g.ResolveSacrificeChoice(c.ID, c.Chooser, c.SacrificeOptions[0]); err != nil {
			t.Fatalf("ResolveSacrificeChoice: %v", err)
		}
	}
	passPriorityAroundTable(t, g)

	if n := countBattlefieldNamed(g, me.ID, "My Island"); n != 1 {
		t.Errorf("my lands = %d, want 1 (4 - 3)", n)
	}
	if n := countBattlefieldNamed(g, opp.ID, "Their Island"); n != 0 {
		t.Errorf("their lands = %d, want 0", n)
	}
	if !g.Battlefield.Contains(theirBear) || !g.Battlefield.Contains(swallower) {
		t.Error("a nonland permanent was sacrificed")
	}
}

// Destroys an opponent's flier; neither your own flier nor their
// ground creature is a legal target.
func TestArborColossusDestroysAnOpposingFlier(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	colossus := pushMonster(g, me.ID, "Arbor Colossus", arborColossusOracle, 6, 6)
	flier := func(owner uuid.UUID, name string) uuid.UUID {
		return pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bird", Keywords: []string{"flying"},
			Power: 1, Toughness: 1, Owner: owner, Controller: owner,
		})
	}
	theirFlier := flier(opp.ID, "Their Bird")
	myFlier := flier(me.ID, "My Bird")
	theirGround := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	activateMonstrosity(t, g, me.ID, colossus, 0)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if hasID(p.PickTargetCards, myFlier) || hasID(p.PickTargetCards, theirGround) {
		t.Errorf("illegal targets offered: %v", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, theirFlier)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirFlier) {
		t.Error("the flier survived")
	}
	if !hasString(effectiveAbilities(t, g, colossus), "reach") {
		t.Error("Arbor Colossus has no reach")
	}
}

// With no opposing flier the trigger never asks (CR 603.3d) and the
// Giant is still monstrous.
func TestArborColossusWithNoFlierJustBecomesMonstrous(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	colossus := pushMonster(g, me.ID, "Arbor Colossus", arborColossusOracle, 6, 6)
	pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	activateMonstrosity(t, g, me.ID, colossus, 0)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Error("asked for a target with no flier on the table")
	}
	if !monstrousNow(g, colossus) {
		t.Error("not monstrous")
	}
}

// Hexproof and indestructible only once monstrous.
func TestFleecemaneLionIsProtectedOnlyWhenMonstrous(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lion := pushMonster(g, me.ID, "Fleecemane Lion", fleecemaneLionOracle, 3, 3)
	if abs := effectiveAbilities(t, g, lion); hasString(abs, "hexproof") || hasString(abs, "indestructible") {
		t.Fatalf("protected before monstrosity: %v", abs)
	}
	activateMonstrosity(t, g, me.ID, lion, 0)
	passPriorityAroundTable(t, g)
	abs := effectiveAbilities(t, g, lion)
	if !hasString(abs, "hexproof") || !hasString(abs, "indestructible") {
		t.Errorf("monstrous Lion abilities = %v, want hexproof and indestructible", abs)
	}
	if p := e2Card(t, g, lion).CurrentPower(); p != 4 {
		t.Errorf("power = %d, want 4", p)
	}
}

// Monstrosity X puts X counters on and switches trample on.
func TestDomesticatedHydraMonstrosityX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hydra := pushMonster(g, me.ID, "Domesticated Hydra", domesticatedHydraOrcl, 3, 3)
	if hasString(effectiveAbilities(t, g, hydra), "trample") {
		t.Fatal("trample before monstrosity")
	}
	activateMonstrosity(t, g, me.ID, hydra, 4)
	passPriorityAroundTable(t, g)
	if n := plusOneCounters(g, hydra); n != 4 {
		t.Errorf("+1/+1 counters = %d, want 4", n)
	}
	if !hasString(effectiveAbilities(t, g, hydra), "trample") {
		t.Error("no trample once monstrous")
	}
}

// X = 0 places no counters — so no counter event invalidates the layer
// pass — and the gated trample must still switch on: the designation's
// own event is what announces the change.
func TestDomesticatedHydraAtXZeroStillGainsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hydra := pushMonster(g, me.ID, "Domesticated Hydra", domesticatedHydraOrcl, 3, 3)
	if hasString(effectiveAbilities(t, g, hydra), "trample") {
		t.Fatal("trample before monstrosity")
	}
	activateMonstrosity(t, g, me.ID, hydra, 0)
	passPriorityAroundTable(t, g)
	if plusOneCounters(g, hydra) != 0 {
		t.Fatal("X = 0 placed counters")
	}
	if !hasString(effectiveAbilities(t, g, hydra), "trample") {
		t.Error("a Hydra made monstrous at X = 0 has no trample")
	}
}

// X X/X Hydras; at X = 0, none — and the Broodmaster is spent.
func TestHydraBroodmasterMakesXHydras(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	brood := pushMonster(g, me.ID, "Hydra Broodmaster", hydraBroodmasterOracle, 7, 7)
	activateMonstrosity(t, g, me.ID, brood, 3)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Hydra"); n != 3 {
		t.Fatalf("Hydra tokens = %d, want 3", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Hydra" && (c.Power != 3 || c.Toughness != 3) {
			t.Errorf("token is %d/%d, want 3/3", c.Power, c.Toughness)
		}
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	brood2 := pushMonster(g2, me2.ID, "Hydra Broodmaster", hydraBroodmasterOracle, 7, 7)
	activateMonstrosity(t, g2, me2.ID, brood2, 0)
	passPriorityAroundTable(t, g2)
	if n := countBattlefieldNamed(g2, me2.ID, "Hydra"); n != 0 {
		t.Errorf("X = 0 made %d tokens", n)
	}
	if !monstrousNow(g2, brood2) {
		t.Error("X = 0 did not make it monstrous")
	}
}
