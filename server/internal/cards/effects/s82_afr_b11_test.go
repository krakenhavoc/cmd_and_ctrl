package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b11_test.go — S82 AFR batch 11: the dice cards. Every test runs a
// card under many pinned RNG keys with a fixed source id (the stream is
// seeded from both), reads the natural result off the announced die, and
// checks the printed row for that result; each table must have shown every
// row at least once, so no row goes untested.

const afrB11Keys = 300

var afrB11Source = uuid.MustParse("a0b11000-0000-4000-8000-000000000001")

// afrB11Row is the row of a d20 table with rows 1—a, a+1—b and 20.
func afrB11Row(roll, a, b int) int {
	switch {
	case roll <= a:
		return 0
	case roll <= b:
		return 1
	default:
		return 2
	}
}

// afrB11Rows runs fn for each key until `want` distinct rows have been
// checked; fn reports the row it just verified, or -1 when this key's roll
// is not one the test is about.
func afrB11Rows(t *testing.T, want int, fn func(t *testing.T, g *game.Game) int) {
	t.Helper()
	seen := map[int]bool{}
	for k := 0; k < afrB11Keys && len(seen) < want; k++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(byte(k)))
		if row := fn(t, g); row >= 0 {
			seen[row] = true
		}
	}
	if len(seen) < want {
		t.Fatalf("only %d of %d rows were reached in %d keys", len(seen), want, afrB11Keys)
	}
}

func afrB11Dice(g *game.Game) []game.Event { return randomEvents(g, game.EventRollDie) }

const chaosChannelerOracle = "ea9031de-6169-4b1c-9390-7fdaf876c7cd"

func TestChaosChannelerExilesOneTwoOrThreeAndMayPlayThem(t *testing.T) {
	afrB11Rows(t, 3, func(t *testing.T, g *game.Game) int {
		me := g.Seats[0]
		id := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Chaos Channeler", "Creature — Human Shaman", chaosChannelerOracle, false)
		lib := me.Library.Size()
		attackWith(t, g, g.Seats[1].ID, id)
		passPriorityAroundTable(t, g)
		dice := afrB11Dice(g)
		if len(dice) != 1 || dice[0].Sides != 20 {
			t.Fatalf("rolls: %+v", dice)
		}
		row := afrB11Row(dice[0].Amount, 9, 19)
		if got := g.Exile.Size(); got != row+1 {
			t.Fatalf("roll %d exiled %d cards, want %d", dice[0].Amount, got, row+1)
		}
		if me.Library.Size() != lib-(row+1) {
			t.Fatalf("library went %d -> %d", lib, me.Library.Size())
		}
		for _, c := range g.Exile.Cards {
			if !permissionLive(g, exiledPermission(g, c.InstanceID), me.ID) {
				t.Fatalf("%s is not playable this turn", c.Name)
			}
		}
		return row
	})
}

// --- helpers ----------------------------------------------------------

// afrB11Activate walks to the active seat's precombat main, floats mana and
// activates ability idx strictly, then lets the stack settle.
func afrB11Activate(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID, idx int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	advanceTo(t, g, game.StepPrecombatMain)
	if mana != "" {
		afrB11Float(t, g, me, mana)
	}
	params.Strict = true
	if err := g.ActivateCatalogAbility(me.ID, src, idx, params); err != nil {
		return err
	}
	passPriorityAroundTable(t, g)
	return nil
}

// afrB11Float floats mana: generic costs are paid with colourless, so "{4}"
// is written {C}{C}{C}{C} (AddManaForEffect takes colours only).
func afrB11Float(t *testing.T, g *game.Game, me *game.Player, mana string) {
	t.Helper()
	var out strings.Builder
	for i := 0; i < len(mana); i++ {
		if mana[i] == '{' && i+2 < len(mana) && mana[i+1] >= '0' && mana[i+1] <= '9' && mana[i+2] == '}' {
			out.WriteString(strings.Repeat("{C}", int(mana[i+1]-'0')))
			i += 2
			continue
		}
		out.WriteByte(mana[i])
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, out.String()); err != nil {
		t.Fatalf("AddManaForEffect(%s): %v", out.String(), err)
	}
}

// afrB11Roll is the one die announced since event index `from`.
func afrB11Roll(t *testing.T, g *game.Game, from, sides int) int {
	t.Helper()
	var got []game.Event
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventRollDie {
			got = append(got, ev)
		}
	}
	if len(got) != 1 || got[0].Sides != sides {
		t.Fatalf("rolls = %+v, want one d%d", got, sides)
	}
	return got[0].Amount
}

func afrB11Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	z := g.FindCardZoneForEffect(id)
	if z == nil {
		return ""
	}
	return z.Kind
}

func afrB11Put(g *game.Game, p *game.Player, z *game.Zone, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	z.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

// afrB11Confirm answers the open confirm prompt.
func afrB11Confirm(t *testing.T, g *game.Game, who uuid.UUID, accept bool) {
	t.Helper()
	c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, who)
	if c == nil {
		t.Fatalf("no confirm prompt for %s: %+v", who, g.PendingChoices)
	}
	if err := g.ResolveConfirm(c.ID, who, accept); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
}

func afrB11Settle(t *testing.T, g *game.Game) {
	t.Helper()
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
}

// --- Treasure Chest ---------------------------------------------------

const treasureChestOracle = "3df26058-c58d-4869-ae32-922d28c3aca2"

func TestTreasureChestEveryRow(t *testing.T) {
	afrB11Rows(t, 4, func(t *testing.T, g *game.Game) int {
		me := g.Seats[0]
		chest := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Treasure Chest", "Artifact", treasureChestOracle, false)
		life, hand, lib := me.Life, me.Hand.Size(), me.Library.Size()
		from := len(g.Events)
		if err := afrB11Activate(t, g, me, chest, 0, "{4}", game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if afrB11Zone(g, chest) != game.ZoneGraveyard {
			t.Fatalf("the Chest is a cost: it is in %q, want the graveyard", afrB11Zone(g, chest))
		}
		roll := afrB11Roll(t, g, from, 20)
		switch {
		case roll == 1:
			if me.Life != life-3 {
				t.Fatalf("roll 1: life %d -> %d, want 3 lost", life, me.Life)
			}
			return 0
		case roll <= 9:
			if n := randomTokenCount(g, "Treasure"); n != 5 {
				t.Fatalf("roll %d: %d Treasures, want 5", roll, n)
			}
			return 1
		case roll <= 19:
			if me.Life != life+3 || me.Hand.Size() != hand+3 || me.Library.Size() != lib-3 {
				t.Fatalf("roll %d: life %d->%d hand %d->%d library %d->%d, want +3 life, three cards drawn", roll, life, me.Life, hand, me.Hand.Size(), lib, me.Library.Size())
			}
			return 2
		}
		// 20: a search for a card; a non-artifact goes to the hand.
		c := pendingOfKind(g, game.PendingChoiceSearchLibrary)
		if c == nil {
			t.Fatalf("roll 20 asked no search: %+v", g.PendingChoices)
		}
		pick := me.Library.Cards[5].InstanceID
		if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{pick}); err != nil {
			t.Fatalf("search: %v", err)
		}
		passPriorityAroundTable(t, g)
		if afrB11Zone(g, pick) != game.ZoneHand || me.Hand.Size() != hand+1 || me.Library.Size() != lib-1 {
			t.Fatalf("a nonartifact found by 20 is in %q (hand %d->%d)", afrB11Zone(g, pick), hand, me.Hand.Size())
		}
		return 3
	})
}

// A found artifact is a "you may" put onto the battlefield: yes puts it
// there, no leaves it in the (shuffled) library; a nonartifact never asks.
func TestTreasureChestFoundArtifactMayEnter(t *testing.T) {
	for _, accept := range []bool{true, false} {
		reached := false
		for k := 0; k < afrB11Keys && !reached; k++ {
			g := newCatalogGame(t)
			g.SetRNGKeyForTest(randomModesKey(byte(k)))
			me := g.Seats[0]
			chest := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Treasure Chest", "Artifact", treasureChestOracle, false)
			gem := afrB11Put(g, me, me.Library, "Sol Ring Stand-in", "Artifact", "{1}")
			from := len(g.Events)
			if err := afrB11Activate(t, g, me, chest, 0, "{4}", game.ActivateAbilityParams{}); err != nil {
				t.Fatal(err)
			}
			if afrB11Roll(t, g, from, 20) != 20 {
				continue
			}
			reached = true
			c := pendingOfKind(g, game.PendingChoiceSearchLibrary)
			if c == nil {
				t.Fatal("no search prompt")
			}
			if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{gem}); err != nil {
				t.Fatal(err)
			}
			pc := latestChooseCardsFor(g, me.ID)
			if pc == nil {
				t.Fatalf("an artifact card found was not offered the battlefield: %+v", g.PendingChoices)
			}
			picks := []uuid.UUID{}
			if accept {
				picks = []uuid.UUID{gem}
			}
			if err := g.ResolveChooseCards(pc.ID, me.ID, picks); err != nil {
				t.Fatal(err)
			}
			passPriorityAroundTable(t, g)
			if accept && afrB11Zone(g, gem) != game.ZoneBattlefield {
				t.Fatalf("accepted, the artifact is in %q", afrB11Zone(g, gem))
			}
			if !accept && afrB11Zone(g, gem) != game.ZoneLibrary {
				t.Fatalf("declined, the artifact is in %q, want the library", afrB11Zone(g, gem))
			}
			if me.Hand.Size() > 8 {
				t.Fatal("the artifact was put into the hand")
			}
		}
		if !reached {
			t.Fatalf("no key rolled a 20 (accept=%v)", accept)
		}
	}
}

func TestTreasureChestNeedsFourManaAndIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chest := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Treasure Chest", "Artifact", treasureChestOracle, false)
	if err := afrB11Activate(t, g, me, chest, 0, "{3}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with three mana")
	}
	if afrB11Zone(g, chest) != game.ZoneBattlefield || len(afrB11Dice(g)) != 0 {
		t.Fatal("a refused activation sacrificed the Chest or rolled")
	}
}

// --- Loathsome Troll --------------------------------------------------

const loathsomeTrollOracle = "d360ce89-d80d-4be6-be8c-7e7758cd5840"

func afrB11TrollInGraveyard(g *game.Game, me *game.Player) uuid.UUID {
	me.Graveyard.PushTop(game.Card{
		InstanceID: afrB11Source, Name: "Loathsome Troll", TypeLine: "Creature — Troll", OracleID: loathsomeTrollOracle,
		ManaCost: "{3}{G}{G}", Power: 6, Toughness: 2, Owner: me.ID, Controller: me.ID,
		KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	return afrB11Source
}

func TestLoathsomeTrollRollsFromTheGraveyard(t *testing.T) {
	afrB11Rows(t, 3, func(t *testing.T, g *game.Game) int {
		me := g.Seats[0]
		troll := afrB11TrollInGraveyard(g, me)
		from := len(g.Events)
		if err := afrB11Activate(t, g, me, troll, 0, "{3}{G}", game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate from the graveyard: %v", err)
		}
		roll := afrB11Roll(t, g, from, 20)
		switch {
		case roll <= 9:
			top, err := me.Library.Top()
			if err != nil || top.InstanceID != troll || afrB11Zone(g, troll) != game.ZoneLibrary {
				t.Fatalf("roll %d: the Troll is in %q, want on top of the library", roll, afrB11Zone(g, troll))
			}
			return 0
		case roll <= 19:
			if afrB11Zone(g, troll) != game.ZoneHand {
				t.Fatalf("roll %d: the Troll is in %q, want the hand", roll, afrB11Zone(g, troll))
			}
			return 1
		}
		c := findBattlefieldCardForTest(g, troll)
		if c == nil || !c.Tapped {
			t.Fatalf("roll 20: the Troll is in %q (card %+v), want the battlefield tapped", afrB11Zone(g, troll), c)
		}
		return 2
	})
}

// Only from the graveyard: on the battlefield or in hand there is no such
// ability to activate; and without {3}{G} it is refused.
func TestLoathsomeTrollIsRefusedOutsideTheGraveyardOrUnpaid(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bf := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Loathsome Troll", "Creature — Troll", loathsomeTrollOracle, false)
	if err := afrB11Activate(t, g, me, bf, 0, "{3}{G}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated from the battlefield")
	}
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	troll := afrB11TrollInGraveyard(g2, me2)
	if err := afrB11Activate(t, g2, me2, troll, 0, "{3}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without {G}")
	}
	if afrB11Zone(g2, troll) != game.ZoneGraveyard || len(afrB11Dice(g2)) != 0 {
		t.Fatal("a refused activation moved the Troll or rolled")
	}
}

// The Troll that left the graveyard in response is not "this card": nothing
// moves (CR 400.7).
func TestLoathsomeTrollGoneFromTheGraveyardIsLeftAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	troll := afrB11TrollInGraveyard(g, me)
	advanceTo(t, g, game.StepPrecombatMain)
	afrB11Float(t, g, me, "{3}{G}")
	if err := g.ActivateCatalogAbility(me.ID, troll, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(troll) })
	passPriorityAroundTable(t, g)
	if afrB11Zone(g, troll) != game.ZoneExile {
		t.Fatalf("the exiled Troll moved to %q", afrB11Zone(g, troll))
	}
}

// --- Power of Persuasion ---------------------------------------------

const powerOfPersuasionOracle = "5c615c99-77b7-4bea-9286-1e8d257153fb"

func afrB11CastPower(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID, strict bool) error {
	t.Helper()
	me.Hand.PushTop(game.Card{
		InstanceID: afrB11Source, Name: "Power of Persuasion", TypeLine: "Sorcery", ManaCost: "{2}{U}",
		OracleID: powerOfPersuasionOracle, Owner: me.ID, Controller: me.ID,
	})
	advanceTo(t, g, game.StepPrecombatMain)
	afrB11Float(t, g, me, "{2}{U}")
	return g.CastSpell(me.ID, afrB11Source, game.CastSpellParams{Strict: strict, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
}

func TestPowerOfPersuasionEveryRow(t *testing.T) {
	afrB11Rows(t, 4, func(t *testing.T, g *game.Game) int {
		me, opp := g.Seats[0], g.Seats[1]
		bear := pushCatalogPermanentWithID(g, uuid.New(), opp.ID, "Bear", "Creature — Bear", "", false)
		from := len(g.Events)
		if err := afrB11CastPower(t, g, me, bear, true); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		roll := afrB11Roll(t, g, from, 20)
		switch {
		case roll <= 9:
			if afrB11Zone(g, bear) != game.ZoneHand || !opp.Hand.Contains(bear) {
				t.Fatalf("roll %d: the creature is in %q, want its owner's hand", roll, afrB11Zone(g, bear))
			}
			return 0
		case roll <= 19:
			c := pendingOfKind(g, game.PendingChoicePutInLibrary)
			if c == nil || c.Chooser != opp.ID {
				t.Fatalf("roll %d: its owner was not asked for the lane: %+v", roll, g.PendingChoices)
			}
			top := roll%2 == 0
			bottom, topOrder := []uuid.UUID{bear}, []uuid.UUID(nil)
			if top {
				bottom, topOrder = nil, []uuid.UUID{bear}
			}
			if err := g.ResolvePutInLibrary(c.ID, opp.ID, bottom, topOrder); err != nil {
				t.Fatalf("choosing the lane: %v", err)
			}
			passPriorityAroundTable(t, g)
			if top {
				if c, err := opp.Library.Top(); err != nil || c.InstanceID != bear {
					t.Fatalf("roll %d: not on top of its owner's library", roll)
				}
				return 1
			}
			if c, err := opp.Library.Bottom(); err != nil || c.InstanceID != bear {
				t.Fatalf("roll %d: not on the bottom of its owner's library", roll)
			}
			return 2
		}
		c := findBattlefieldCardForTest(g, bear)
		if c == nil || c.Controller != me.ID || c.Owner != opp.ID {
			t.Fatalf("roll 20: card %+v, want it controlled by the caster and owned by the opponent", c)
		}
		return 3
	})
}

// Row 3 lasts through the end of the caster's NEXT turn and then ends.
func TestPowerOfPersuasionControlLastsUntilTheEndOfYourNextTurn(t *testing.T) {
	for k := 0; k < afrB11Keys; k++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(byte(k)))
		me, opp := g.Seats[0], g.Seats[1]
		bear := pushCatalogPermanentWithID(g, uuid.New(), opp.ID, "Bear", "Creature — Bear", "", false)
		from := len(g.Events)
		if err := afrB11CastPower(t, g, me, bear, true); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		if afrB11Roll(t, g, from, 20) != 20 {
			continue
		}
		controller := func() uuid.UUID {
			if c := findBattlefieldCardForTest(g, bear); c != nil {
				return c.Controller
			}
			return uuid.Nil
		}
		startTurn := g.Turn.Seq
		// Through the rest of this turn and the opponents' turns: still mine.
		for i := 0; i < 400 && g.Turn.Seq == startTurn; i++ {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
		if controller() != me.ID {
			t.Fatalf("control went back at the end of the turn it was gained (turn %d)", g.Turn.Seq)
		}
		for i := 0; i < 800 && !(g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepEnd); i++ {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
			if g.Turn.ActiveSeat != 0 && controller() != me.ID {
				t.Fatalf("control ended during another player's turn")
			}
		}
		if g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepEnd {
			t.Fatal("never reached the caster's next end step")
		}
		if controller() != me.ID {
			t.Fatal("control ended before the end of the caster's next turn")
		}
		for i := 0; i < 60 && g.Turn.ActiveSeat == 0; i++ {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
		if controller() != opp.ID {
			t.Fatalf("control did not revert after the caster's next turn: %v", controller())
		}
		return
	}
	t.Fatal("no key rolled a 20")
}

func TestPowerOfPersuasionTargetsOnlyAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := pushCatalogPermanentWithID(g, uuid.New(), me.ID, "Own Bear", "Creature — Bear", "", false)
	if err := afrB11CastPower(t, g, me, mine, true); err == nil {
		t.Fatal("targeted a creature the caster controls")
	}
	if len(afrB11Dice(g)) != 0 {
		t.Fatal("a refused cast rolled")
	}
}

// With the creature gone as the spell resolves, nothing is rolled (CR 608.2b).
func TestPowerOfPersuasionFizzlesWithoutARoll(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushCatalogPermanentWithID(g, uuid.New(), opp.ID, "Bear", "Creature — Bear", "", false)
	if err := afrB11CastPower(t, g, me, bear, true); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(bear) })
	passPriorityAroundTable(t, g)
	if len(afrB11Dice(g)) != 0 {
		t.Fatal("rolled for a spell with no legal target")
	}
}

// --- Underdark Rift ---------------------------------------------------

const underdarkRiftOracle = "199c7604-6b4b-4cec-b6c2-b6b2918b11c8"

func TestUnderdarkRiftPutsTheTargetBeneathTheTopXCards(t *testing.T) {
	seen := map[int]bool{}
	for k := 0; k < afrB11Keys && len(seen) < 4; k++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(byte(k)))
		me, opp := g.Seats[0], g.Seats[1]
		rift := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Underdark Rift", "Land", underdarkRiftOracle, false)
		bear := pushCatalogPermanentWithID(g, uuid.New(), opp.ID, "Bear", "Creature — Bear", "", false)
		from := len(g.Events)
		err := afrB11Activate(t, g, me, rift, 0, "{5}", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
		if err != nil {
			t.Fatalf("activate: %v", err)
		}
		if afrB11Zone(g, rift) != game.ZoneExile {
			t.Fatalf("the Rift is a cost: it is in %q, want exile", afrB11Zone(g, rift))
		}
		x := afrB11Roll(t, g, from, 10)
		idx := -1
		for i, c := range opp.Library.Cards {
			if c.InstanceID == bear {
				idx = len(opp.Library.Cards) - 1 - i // the top is the last element
			}
		}
		if idx != x {
			t.Fatalf("roll %d: the creature is %d cards down its owner's library, want just beneath the top %d", x, idx, x)
		}
		seen[x] = true
	}
	if len(seen) < 4 {
		t.Fatalf("only %d different rolls reached", len(seen))
	}
}

// A short library takes the card on the bottom, and the ability is a
// sorcery: refused in combat; an enchantment or land is no target.
func TestUnderdarkRiftShortLibrarySorceryTimingAndTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rift := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Underdark Rift", "Land", underdarkRiftOracle, false)
	bear := pushCatalogPermanentWithID(g, uuid.New(), opp.ID, "Bear", "Creature — Bear", "", false)
	land := pushCatalogPermanentWithID(g, uuid.New(), opp.ID, "Forest", "Basic Land — Forest", "", false)
	ench := pushCatalogPermanentWithID(g, uuid.New(), opp.ID, "Pacifism", "Enchantment — Aura", "", false)
	if err := afrB11Activate(t, g, me, rift, 0, "{5}", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err == nil {
		t.Fatal("targeted a land")
	}
	if err := afrB11Activate(t, g, me, rift, 0, "{5}", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ench}}}); err == nil {
		t.Fatal("targeted an enchantment")
	}
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, rift, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	if afrB11Zone(g, rift) != game.ZoneBattlefield {
		t.Fatal("a refused activation exiled the Rift")
	}

	// A library of two cards, whatever X is, takes it at the bottom or
	// beneath the top X: never deeper than the library allows.
	opp.Library.Cards = opp.Library.Cards[:2]
	advanceTo(t, g, game.StepPostcombatMain)
	afrB11Float(t, g, me, "{5}")
	if err := g.ActivateCatalogAbility(me.ID, rift, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if afrB11Zone(g, bear) != game.ZoneLibrary {
		t.Fatalf("the creature is in %q", afrB11Zone(g, bear))
	}
	if n := opp.Library.Size(); n != 3 {
		t.Fatalf("library of %d, want 3", n)
	}
	b, _ := opp.Library.Bottom()
	x := afrB11Roll(t, g, 0, 10)
	if x >= 2 && b.InstanceID != bear {
		t.Fatalf("roll %d over a two-card library: not on the bottom", x)
	}
}

// --- Bag of Tricks ----------------------------------------------------

const bagOfTricksOracle = "f14ce7d7-6cfa-4f88-a381-d6e912c18b5d"

func afrB11Cost(n int) string { return "{" + string(rune('0'+n)) + "}" }

// afrB11TrickLibrary stacks, top first: a land, a sorcery, then creatures of
// mana value 5 3 7 1 8 2 6 4 each with an instant between, then filler.
func afrB11TrickLibrary(me *game.Player) map[int]uuid.UUID {
	me.Library.Cards = nil
	byMV := map[int]uuid.UUID{}
	order := []int{5, 3, 7, 1, 8, 2, 6, 4}
	var cards []game.Card
	cards = append(cards, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"})
	for _, mv := range order {
		id := uuid.New()
		byMV[mv] = id
		cards = append(cards,
			game.Card{InstanceID: id, Name: "Creature " + afrB11Cost(mv), TypeLine: "Creature — Bear", ManaCost: afrB11Cost(mv), Power: 1, Toughness: 1},
			game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"})
	}
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		if c.InstanceID == uuid.Nil {
			c.InstanceID = uuid.New()
		}
		c.Owner, c.Controller = me.ID, me.ID
		me.Library.PushTop(c)
	}
	return byMV
}

func TestBagOfTricksPutsTheFirstCreatureWithThatManaValueAndBottomsTheRest(t *testing.T) {
	afrB11Rows(t, 4, func(t *testing.T, g *game.Game) int {
		me := g.Seats[0]
		bag := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Bag of Tricks", "Artifact", bagOfTricksOracle, false)
		byMV := afrB11TrickLibrary(me)
		size := me.Library.Size()
		var before []uuid.UUID // top first
		for i := len(me.Library.Cards) - 1; i >= 0; i-- {
			before = append(before, me.Library.Cards[i].InstanceID)
		}
		from := len(g.Events)
		if err := afrB11Activate(t, g, me, bag, 0, "{4}{G}", game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if c := findBattlefieldCardForTest(g, bag); c == nil || !c.Tapped {
			t.Fatal("the Bag was not tapped as a cost")
		}
		x := afrB11Roll(t, g, from, 8)
		hit := byMV[x]
		if afrB11Zone(g, hit) != game.ZoneBattlefield {
			t.Fatalf("roll %d: the creature of that mana value is in %q, want the battlefield", x, afrB11Zone(g, hit))
		}
		// Everything else is still in the library; the revealed cards went to
		// the bottom, so the first unrevealed card is on top.
		if me.Library.Size() != size-1 {
			t.Fatalf("library %d -> %d, want one card gone", size, me.Library.Size())
		}
		for mv, id := range byMV {
			if mv != x && afrB11Zone(g, id) != game.ZoneLibrary {
				t.Fatalf("creature of mana value %d is in %q", mv, afrB11Zone(g, id))
			}
		}
		// The cards revealed before the hit are on the bottom (in any order) and
		// the first unrevealed card is now on top.
		n := 0
		for i, id := range before {
			if id == hit {
				n = i + 1
			}
		}
		top, _ := me.Library.Top()
		if n < len(before) && top.InstanceID != before[n] {
			t.Fatalf("roll %d: the top of the library is not the first unrevealed card", x)
		}
		bottom := map[uuid.UUID]bool{}
		for i := 0; i < n-1; i++ {
			bottom[me.Library.Cards[i].InstanceID] = true
		}
		for _, id := range before[:n-1] {
			if !bottom[id] {
				t.Fatalf("roll %d: a revealed card is not among the bottom %d", x, n-1)
			}
		}
		return x - 1
	})
}

// No creature card with that mana value: the library is revealed, nothing
// enters, and every revealed card goes to the bottom.
func TestBagOfTricksWithNoMatchRevealsItAllAndPutsNothingOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bag := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Bag of Tricks", "Artifact", bagOfTricksOracle, false)
	me.Library.Cards = nil
	var ids []uuid.UUID
	for i := 0; i < 5; i++ {
		ids = append(ids, afrB11Put(g, me, me.Library, "Shock", "Instant", "{R}"))
	}
	before := len(g.Battlefield.Cards)
	if err := afrB11Activate(t, g, me, bag, 0, "{4}{G}", game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if len(g.Battlefield.Cards) != before || me.Library.Size() != 5 {
		t.Fatalf("battlefield %d -> %d, library %d: nothing should enter and nothing leave the library", before, len(g.Battlefield.Cards), me.Library.Size())
	}
	for _, id := range ids {
		if afrB11Zone(g, id) != game.ZoneLibrary {
			t.Fatalf("a revealed card is in %q", afrB11Zone(g, id))
		}
	}
}

// Without {4}{G} or an untapped Bag it is refused and rolls nothing.
func TestBagOfTricksIsRefusedWithoutFiveManaAndNeedsTheTap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bag := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Bag of Tricks", "Artifact", bagOfTricksOracle, false)
	if err := afrB11Activate(t, g, me, bag, 0, "{3}{G}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with four mana")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bag {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := afrB11Activate(t, g, me, bag, 0, "{4}{G}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated a tapped Bag")
	}
	if len(afrB11Dice(g)) != 0 {
		t.Fatal("a refused activation rolled")
	}
}

// --- Ebony Fly --------------------------------------------------------

const ebonyFlyOracle = "2a2f2d6c-3937-4e4b-9e4a-3480e0c787d0"

func TestEbonyFlyEntersTappedAndTapsForColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.PushTop(game.Card{InstanceID: afrB11Source, Name: "Ebony Fly", TypeLine: "Artifact", ManaCost: "{2}", OracleID: ebonyFlyOracle, Owner: me.ID, Controller: me.ID})
	advanceTo(t, g, game.StepPrecombatMain)
	afrB11Float(t, g, me, "{2}")
	if err := g.CastSpell(me.ID, afrB11Source, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardForTest(g, afrB11Source)
	if c == nil || !c.Tapped {
		t.Fatalf("the Fly did not enter tapped: %+v", c)
	}
	if err := g.ActivateManaAbility(me.ID, afrB11Source, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("tapped for mana while tapped")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == afrB11Source {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	before := me.ManaPool.Total()
	if err := g.ActivateManaAbility(me.ID, afrB11Source, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if me.ManaPool.Total() != before+1 {
		t.Fatalf("pool %d -> %d, want {C}", before, me.ManaPool.Total())
	}
}

func TestEbonyFlyBecomesAnXByXFlyingInsectWhenYouSayYes(t *testing.T) {
	afrB11Rows(t, 4, func(t *testing.T, g *game.Game) int {
		me := g.Seats[0]
		fly := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Ebony Fly", "Artifact", ebonyFlyOracle, false)
		from := len(g.Events)
		if err := afrB11Activate(t, g, me, fly, 0, "{4}", game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		x := afrB11Roll(t, g, from, 6)
		if findBattlefieldCardForTest(g, fly).IsCreature() {
			t.Fatal("it became a creature before the player said so")
		}
		afrB11Confirm(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
		c := findBattlefieldCardForTest(g, fly)
		if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Insect") {
			t.Fatalf("roll %d: type line %q, want an Insect artifact creature", x, c.TypeLine)
		}
		if p := effectivePower(t, g, fly); p != x {
			t.Fatalf("roll %d: power %d", x, p)
		}
		if !hasEffectiveKeyword(t, g, fly, "flying") {
			t.Fatalf("roll %d: no flying", x)
		}
		return x - 1
	})
}

func TestEbonyFlyDecliningLeavesItAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fly := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Ebony Fly", "Artifact", ebonyFlyOracle, false)
	if err := afrB11Activate(t, g, me, fly, 0, "{4}", game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	afrB11Confirm(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, fly).IsCreature() {
		t.Fatal("declined, it still became a creature")
	}
	if len(afrB11Dice(g)) != 1 {
		t.Fatal("the d6 was not rolled")
	}
}

func TestEbonyFlyIsOnlyACreatureUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fly := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Ebony Fly", "Artifact", ebonyFlyOracle, false)
	if err := afrB11Activate(t, g, me, fly, 0, "{4}", game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	afrB11Confirm(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !findBattlefieldCardForTest(g, fly).IsCreature() {
		t.Fatal("not a creature")
	}
	start := g.Turn.Seq
	for i := 0; i < 40 && g.Turn.Seq == start; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if findBattlefieldCardForTest(g, fly).IsCreature() {
		t.Fatal("still a creature in the next turn")
	}
}

// Whenever it attacks, another target attacking creature gains flying.
func TestEbonyFlyAttackGivesAnotherAttackerFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fly := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Ebony Fly", "Artifact", ebonyFlyOracle, false)
	bear := pushCatalogPermanentWithID(g, uuid.New(), me.ID, "Bear", "Creature — Bear", "", false)
	if err := afrB11Activate(t, g, me, fly, 0, "{4}", game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	afrB11Confirm(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{fly, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	if c := pendingOfKind(g, game.PendingChoicePickTarget); c != nil {
		answerPickTarget(t, g, bear)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Fatal("the other attacker did not gain flying")
	}
}

// --- Sword of Hours ---------------------------------------------------

const swordOfHoursOracle = "abaddcce-9039-4b4e-baf8-fdab304b5c52"

func afrB11Equipped(t *testing.T, g *game.Game, me *game.Player, counters int) (sword, creature uuid.UUID) {
	t.Helper()
	sword = pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Sword of Hours", "Artifact — Equipment", swordOfHoursOracle, false)
	creature = pushCatalogPermanentWithID(g, uuid.New(), me.ID, "Bear", "Creature — Bear", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == creature && counters > 0 {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: counters}
			}
		}
	})
	if err := afrB11Activate(t, g, me, sword, 0, "{2}", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}}}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	if c := findBattlefieldCardForTest(g, sword); c == nil || !c.IsAttachedTo(creature) {
		t.Fatal("the Sword did not attach")
	}
	return sword, creature
}

func TestSwordOfHoursPutsACounterOnTheEquippedCreatureWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, creature := afrB11Equipped(t, g, me, 0)
	other := pushCatalogPermanentWithID(g, uuid.New(), me.ID, "Other Bear", "Creature — Bear", "", false)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{creature, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if n := counterOnBattlefieldCard(g, creature, game.CounterPlusOne); n != 1 {
		t.Fatalf("equipped attacker has %d +1/+1 counters, want 1", n)
	}
	if n := counterOnBattlefieldCard(g, other, game.CounterPlusOne); n != 0 {
		t.Fatalf("an unequipped attacker got %d counters", n)
	}
}

func swordDamage(g *game.Game, creature uuid.UUID, amounts ...int) {
	g.WithWriteLock(func() {
		for i, n := range amounts {
			g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: g.Seats[0].ID, Source: creature, Target: g.Seats[1+i].ID, Amount: n, Combat: true})
		}
	})
}

// Damage 6: a roll of 7 or more doubles, 12 always does (12 > 6 anyway);
// with damage 12, only the 12 does. Three rows reached under real keys.
func TestSwordOfHoursDoublesOnAGreaterRollOrATwelve(t *testing.T) {
	for _, dealt := range []int{6, 12} {
		seen := map[int]bool{}
		for k := 0; k < afrB11Keys && len(seen) < 2+boolInt(dealt == 6); k++ {
			g := newCatalogGame(t)
			g.SetRNGKeyForTest(randomModesKey(byte(k)))
			me := g.Seats[0]
			_, creature := afrB11Equipped(t, g, me, 2)
			from := len(g.Events)
			swordDamage(g, creature, dealt)
			afrB11Settle(t, g)
			roll := afrB11Roll(t, g, from, 12)
			n := counterOnBattlefieldCard(g, creature, game.CounterPlusOne)
			want := 2
			if roll > dealt || roll == 12 {
				want = 4
			}
			if n != want {
				t.Fatalf("damage %d, roll %d: %d counters, want %d", dealt, roll, n, want)
			}
			seen[n*100+boolInt(roll == 12)] = true
		}
		if len(seen) < 2 {
			t.Fatalf("damage %d: only %d outcomes reached", dealt, len(seen))
		}
	}
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Combat damage dealt to several recipients at once is one trigger, one roll,
// and "the damage dealt" is the total.
func TestSwordOfHoursIsOneRollAgainstTheTotalDamage(t *testing.T) {
	reached := false
	for k := 0; k < afrB11Keys && !reached; k++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(byte(k)))
		me := g.Seats[0]
		_, creature := afrB11Equipped(t, g, me, 1)
		from := len(g.Events)
		swordDamage(g, creature, 3, 3, 3) // total 9
		afrB11Settle(t, g)
		roll := afrB11Roll(t, g, from, 12)
		n := counterOnBattlefieldCard(g, creature, game.CounterPlusOne)
		want := 1
		if roll > 9 {
			want = 2
		}
		if n != want {
			t.Fatalf("roll %d against 9 damage: %d counters, want %d", roll, n, want)
		}
		if roll >= 4 && roll <= 9 {
			reached = true // would have doubled against one 3-damage share
		}
	}
	if !reached {
		t.Fatal("no key rolled between the share and the total")
	}
}

// The creature that left the battlefield has no counters to double.
func TestSwordOfHoursDoublesNothingOnACreatureThatLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	_, creature := afrB11Equipped(t, g, me, 2)
	swordDamage(g, creature, 1)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(creature) })
	afrB11Settle(t, g)
	if afrB11Zone(g, creature) == game.ZoneBattlefield {
		t.Fatal("still on the battlefield")
	}
}

// --- Bag of Devouring -------------------------------------------------

const bagOfDevouringOracle = "4443345d-a292-4394-97dd-0e6f4d9aedf7"

func afrB11Untap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
}

// afrB11SacrificeToTheBag pays the draw ability with `victim` and lets the
// exile trigger resolve.
func afrB11SacrificeToTheBag(t *testing.T, g *game.Game, me *game.Player, bag, victim uuid.UUID) error {
	t.Helper()
	afrB11Untap(g, bag)
	err := afrB11Activate(t, g, me, bag, 0, "{2}", game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{victim}})
	if err == nil {
		afrB11Settle(t, g)
	}
	return err
}

func TestBagOfDevouringExilesWhatYouSacrificeAndDrawsACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bag := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Bag of Devouring", "Artifact", bagOfDevouringOracle, false)
	creature := pushCatalogPermanentWithID(g, uuid.New(), me.ID, "Bear", "Creature — Bear", "", false)
	trinket := pushCatalogPermanentWithID(g, uuid.New(), me.ID, "Trinket", "Artifact", "", false)
	hand := me.Hand.Size()
	if err := afrB11SacrificeToTheBag(t, g, me, bag, creature); err != nil {
		t.Fatalf("sacrifice a creature: %v", err)
	}
	if !inExile(g, creature) {
		t.Fatalf("the sacrificed creature is in %q, want exile", afrB11Zone(g, creature))
	}
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d -> %d, want a card drawn", hand, me.Hand.Size())
	}
	if err := afrB11SacrificeToTheBag(t, g, me, bag, trinket); err != nil {
		t.Fatalf("sacrifice an artifact: %v", err)
	}
	if !inExile(g, trinket) {
		t.Fatalf("the sacrificed artifact is in %q, want exile", afrB11Zone(g, trinket))
	}
	linked := 0
	g.ReadSnapshot(func() {
		bc := findBattlefieldCardForTest(g, bag)
		linked = len(g.ExiledWithForEffect(game.ObjectRefOf(*bc)))
	})
	if linked != 2 {
		t.Fatalf("%d cards linked to the Bag, want 2", linked)
	}
}

func TestBagOfDevouringIgnoresTokensAndOtherPermanentsAndAnOpponentsSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bag := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Bag of Devouring", "Artifact", bagOfDevouringOracle, false)
	token := pushCatalogPermanentWithID(g, uuid.New(), me.ID, "Bear Token", "Token Creature — Bear", "", false)
	if err := afrB11SacrificeToTheBag(t, g, me, bag, token); err != nil {
		t.Fatal(err)
	}
	if inExile(g, token) {
		t.Fatal("a token was exiled")
	}
	// An opponent's Bag-less sacrifice (a creature they sacrifice) is not mine.
	theirs := pushCatalogPermanentWithID(g, uuid.New(), opp.ID, "Their Bear", "Creature — Bear", "", false)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirs) })
	afrB11Settle(t, g)
	if inExile(g, theirs) || afrB11Zone(g, theirs) != game.ZoneGraveyard {
		t.Fatalf("an opponent's sacrificed creature is in %q, want their graveyard", afrB11Zone(g, theirs))
	}
}

func TestBagOfDevouringSacrificingItselfExilesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bag := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Bag of Devouring", "Artifact", bagOfDevouringOracle, false)
	if err := afrB11Activate(t, g, me, bag, 1, "{3}", game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	afrB11Settle(t, g)
	if afrB11Zone(g, bag) != game.ZoneGraveyard {
		t.Fatalf("the Bag is in %q: another, not this one, is exiled", afrB11Zone(g, bag))
	}
}

func TestBagOfDevouringReturnsUpToXOfTheExiledCardsToTheirOwnersHands(t *testing.T) {
	afrB11Rows(t, 3, func(t *testing.T, g *game.Game) int {
		me := g.Seats[0]
		bag := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Bag of Devouring", "Artifact", bagOfDevouringOracle, false)
		var victims []uuid.UUID
		for i := 0; i < 3; i++ {
			v := pushCatalogPermanentWithID(g, uuid.New(), me.ID, "Bear", "Creature — Bear", "", false)
			victims = append(victims, v)
			if err := afrB11SacrificeToTheBag(t, g, me, bag, v); err != nil {
				t.Fatalf("sacrifice %d: %v", i, err)
			}
		}
		afrB11Untap(g, bag)
		from := len(g.Events)
		if err := afrB11Activate(t, g, me, bag, 1, "{3}", game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate the roll: %v", err)
		}
		x := afrB11Roll(t, g, from, 10)
		c := latestChooseCardsFor(g, me.ID)
		if c == nil {
			t.Fatalf("roll %d: not asked which to return: %+v", x, g.PendingChoices)
		}
		if c.ChooseMin != 0 || c.ChooseMax != min(x, 3) || len(c.ChooseCards) != 3 {
			t.Fatalf("roll %d: prompt %d..%d over %d cards, want up to %d of 3", x, c.ChooseMin, c.ChooseMax, len(c.ChooseCards), min(x, 3))
		}
		// Pick as many as allowed (at most all three).
		n := x
		if n > 3 {
			n = 3
		}
		if err := g.ResolveChooseCards(c.ID, me.ID, victims[:n]); err != nil {
			t.Fatalf("picking %d: %v", n, err)
		}
		passPriorityAroundTable(t, g)
		for i, v := range victims {
			want := game.ZoneExile
			if i < n {
				want = game.ZoneHand
			}
			if afrB11Zone(g, v) != want {
				t.Fatalf("roll %d: victim %d is in %q, want %q", x, i, afrB11Zone(g, v), want)
			}
		}
		// More than the roll is refused.
		return min(x, 4) - 1
	})
}

// --- Delina, Wild Mage ------------------------------------------------

const delinaOracle = "8d57dadb-c689-4691-b331-aba15daf46ff"

func afrB11DelinaAttack(t *testing.T, g *game.Game, me, opp *game.Player, delina, target uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(delina, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	answerPickTarget(t, g, target)
	passPriorityAroundTable(t, g)
}

func afrB11Tokens(g *game.Game, me *game.Player, name string) []*game.Card {
	var out []*game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == me.ID && c.IsToken() && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

func TestDelinaCopiesTheChosenCreatureTappedAndAttackingWithAnExileClause(t *testing.T) {
	reached := map[bool]bool{}
	for k := 0; k < afrB11Keys && len(reached) < 2; k++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(byte(k)))
		me, opp := g.Seats[0], g.Seats[1]
		delina := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Delina, Wild Mage", "Legendary Creature — Elf Shaman", delinaOracle, false)
		legend := pushCatalogPermanentWithID(g, uuid.New(), me.ID, "Bear Chief", "Legendary Creature — Bear", "", false)
		from := len(g.Events)
		afrB11DelinaAttack(t, g, me, opp, delina, legend)
		answerAttackSplit(t, g, me, 0)
		roll := afrB11Roll(t, g, from, 20)
		toks := afrB11Tokens(g, me, "Bear Chief")
		if len(toks) != 1 {
			t.Fatalf("roll %d: %d tokens, want one", roll, len(toks))
		}
		tok := toks[0]
		if !tok.Tapped || tok.AttackingTarget != opp.ID {
			t.Fatalf("token tapped=%v attacking=%s, want tapped and attacking %s", tok.Tapped, tok.AttackingTarget, opp.ID)
		}
		if tok.HasSupertype("Legendary") || strings.Contains(tok.TypeLine, "Legendary") {
			t.Fatalf("the token is legendary: %q", tok.TypeLine)
		}
		if len(tok.GrantedAbilities) == 0 {
			t.Fatal("the token does not have the exile clause")
		}
		if roll >= 15 {
			afrB11Confirm(t, g, me.ID, false)
			passPriorityAroundTable(t, g)
			if len(afrB11Tokens(g, me, "Bear Chief")) != 1 || len(afrB11Dice(g)) != 1 {
				t.Fatal("declining to roll again still rolled or made another token")
			}
		} else if c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID); c != nil {
			t.Fatalf("roll %d: offered a re-roll", roll)
		}
		reached[roll >= 15] = true
		// At end of combat the token is exiled.
		tokID := tok.InstanceID
		advanceTo(t, g, game.StepEndCombat)
		passPriorityAroundTable(t, g)
		if afrB11Zone(g, tokID) == game.ZoneBattlefield {
			t.Fatal("the token is still on the battlefield after end of combat")
		}
	}
	if len(reached) < 2 {
		t.Fatal("did not reach both 1—14 and 15—20")
	}
}

// 15—20 asks "roll again": yes rolls another d20 with the same table for the
// same creature, repeatedly.
func TestDelinaRollsAgainWhenYouSayYes(t *testing.T) {
	for k := 0; k < afrB11Keys; k++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(byte(k)))
		me, opp := g.Seats[0], g.Seats[1]
		delina := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Delina, Wild Mage", "Legendary Creature — Elf Shaman", delinaOracle, false)
		bear := pushCatalogPermanentWithID(g, uuid.New(), me.ID, "Bear", "Creature — Bear", "", false)
		afrB11DelinaAttack(t, g, me, opp, delina, bear)
		answerAttackSplit(t, g, me, 0)
		first := afrB11Dice(g)
		if len(first) != 1 || first[0].Amount < 15 {
			continue
		}
		afrB11Confirm(t, g, me.ID, true)
		answerAttackSplit(t, g, me, 0)
		passPriorityAroundTable(t, g)
		rolls := afrB11Dice(g)
		if len(rolls) < 2 {
			t.Fatalf("said yes and rolled %d times", len(rolls))
		}
		for _, r := range rolls {
			if r.Sides != 20 {
				t.Fatalf("rolled a d%d", r.Sides)
			}
		}
		if n := len(afrB11Tokens(g, me, "Bear")); n < 2 {
			t.Fatalf("%d tokens after rolling again, want at least 2", n)
		}
		return
	}
	t.Fatal("no key rolled 15 or more first")
}

// Copying Delina herself: the token is not legendary (so the legend rule
// leaves both), and it has no attack trigger of its own firing on entry.
func TestDelinaCanCopyHerselfAndTheCopyIsNotLegendary(t *testing.T) {
	g := newCatalogGame(t)
	g.SetRNGKeyForTest(randomModesKey(3))
	me, opp := g.Seats[0], g.Seats[1]
	delina := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Delina, Wild Mage", "Legendary Creature — Elf Shaman", delinaOracle, false)
	afrB11DelinaAttack(t, g, me, opp, delina, delina)
	answerAttackSplit(t, g, me, 0)
	toks := afrB11Tokens(g, me, "Delina, Wild Mage")
	if len(toks) < 1 {
		t.Fatalf("no copy of Delina: %+v", g.PendingChoices)
	}
	if strings.Contains(toks[0].TypeLine, "Legendary") {
		t.Fatalf("the copy is legendary: %q", toks[0].TypeLine)
	}
	if afrB11Zone(g, delina) != game.ZoneBattlefield {
		t.Fatal("the legend rule removed Delina")
	}
}

// --- Wizard's Spellbook -----------------------------------------------

const wizardsSpellbookOracle = "740fe41a-16bb-4c97-a03c-f678f15a0a0c"

func afrB11Spellbook(t *testing.T, g *game.Game, me *game.Player, bolt uuid.UUID) error {
	t.Helper()
	return afrB11Activate(t, g, me, afrB11Source, 0, "", game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}})
}

func TestWizardsSpellbookEveryRow(t *testing.T) {
	afrB11Rows(t, 3, func(t *testing.T, g *game.Game) int {
		me, opp := g.Seats[0], g.Seats[1]
		pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Wizard's Spellbook", "Artifact — Book", wizardsSpellbookOracle, false)
		bolt := gyCard(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
		from := len(g.Events)
		if err := afrB11Spellbook(t, g, me, bolt); err != nil {
			t.Fatalf("activate: %v", err)
		}
		roll := afrB11Roll(t, g, from, 20)
		if !inExile(g, bolt) {
			t.Fatalf("roll %d: the target is in %q, want exile", roll, afrB11Zone(g, bolt))
		}
		copies := 0
		var cost string
		for _, c := range cardCopies(g) {
			copies++
			g.ReadSnapshot(func() {
				if perm := g.CastPermissionOnCardForEffect(c, game.ZoneExile); perm != nil && perm.Player == me.ID {
					cost = perm.Cost
				}
			})
		}
		if copies != 1 {
			t.Fatalf("roll %d: %d copies, want 1", roll, copies)
		}
		switch {
		case roll <= 9:
			if cost != "" {
				t.Fatalf("roll %d: copy costs %q, want its mana cost", roll, cost)
			}
			return 0
		case roll <= 19:
			if cost != "{1}" {
				t.Fatalf("roll %d: copy costs %q, want {1}", roll, cost)
			}
			return 1
		}
		if cost != "{0}" {
			t.Fatalf("roll 20: copy costs %q, want free", cost)
		}
		return 2
	})
}

// Row 20 copies EVERY card exiled with the Spellbook, an earlier
// activation's included, and offers each a free cast.
func TestWizardsSpellbookTwentyCopiesEveryCardExiledWithIt(t *testing.T) {
	for k := 0; k < 4*afrB11Keys; k++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(byte(k)))
		me, opp := g.Seats[0], g.Seats[1]
		book := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Wizard's Spellbook", "Artifact — Book", wizardsSpellbookOracle, false)
		first := gyCard(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
		second := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)
		from := len(g.Events)
		if err := afrB11Spellbook(t, g, me, first); err != nil {
			t.Fatal(err)
		}
		if afrB11Roll(t, g, from, 20) == 20 {
			continue
		}
		// The first copy is not cast: it ceases to exist.
		declineResolutionCasts(t, g, me.ID)
		passPriorityAroundTable(t, g)
		if n := len(cardCopies(g)); n != 0 {
			t.Fatalf("%d copies left after declining the cast", n)
		}
		afrB11Untap(g, book)
		from = len(g.Events)
		if err := afrB11Spellbook(t, g, me, second); err != nil {
			t.Fatal(err)
		}
		if afrB11Roll(t, g, from, 20) != 20 {
			continue
		}
		if !inExile(g, first) || !inExile(g, second) {
			t.Fatal("both originals stay exiled")
		}
		copies := cardCopies(g)
		if len(copies) != 2 {
			t.Fatalf("%d copies, want one for each card exiled with the Spellbook", len(copies))
		}
		names := map[string]bool{}
		for _, c := range copies {
			names[c.Name] = true
			g.ReadSnapshot(func() {
				perm := g.CastPermissionOnCardForEffect(c, game.ZoneExile)
				if perm == nil || perm.Player != me.ID || perm.Cost != "{0}" {
					t.Fatalf("copy of %s: permission %+v, want a free cast for the controller", c.Name, perm)
				}
			})
		}
		if !names["Lightning Bolt"] || !names["Divination"] {
			t.Fatalf("copies are %v", names)
		}
		return
	}
	t.Fatal("no key gave a non-20 and then a 20")
}

// Only an instant or sorcery card, and only at sorcery speed, tapping it.
func TestWizardsSpellbookTargetsOnlyInstantsAndSorceriesAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	book := pushCatalogPermanentWithID(g, afrB11Source, me.ID, "Wizard's Spellbook", "Artifact — Book", wizardsSpellbookOracle, false)
	bear := gyCard(opp, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")
	if err := afrB11Spellbook(t, g, me, bear); err == nil {
		t.Fatal("targeted a creature card")
	}
	bolt := gyCard(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	if c := findBattlefieldCardForTest(g, book); c == nil || c.Tapped {
		t.Fatal("a refused activation tapped the Spellbook")
	}
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	book2 := pushCatalogPermanentWithID(g2, afrB11Source, me2.ID, "Wizard's Spellbook", "Artifact — Book", wizardsSpellbookOracle, false)
	bolt2 := gyCard(opp2, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	if err := afrB11Spellbook(t, g2, me2, bolt2); err != nil {
		t.Fatalf("a legal activation: %v", err)
	}
	if c := findBattlefieldCardForTest(g2, book2); c == nil || !c.Tapped {
		t.Fatal("the Spellbook was not tapped")
	}
}
