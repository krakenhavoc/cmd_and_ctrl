package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- extra turns ------------------------------------------------------

// Beacon of Tomorrows gives the TARGET player the turn, and the card
// is shuffled into its owner's library rather than left in the
// graveyard.
func TestBeaconOfTomorrowsGivesTheTargetAnExtraTurnAndShufflesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	opp := (me + 2) % len(g.Seats)
	spell := effectstest.CastCatalogSpell(t, g, "Beacon of Tomorrows", "Sorcery", effectstest.BeaconOfTomorrowsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[opp].ID}})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Seats[me].Graveyard.Contains(spell) {
		t.Error("Beacon of Tomorrows must not be left in the graveyard")
	}
	if !effectstest.LibraryHolds(g.Seats[me], spell) {
		t.Error("Beacon of Tomorrows should have been shuffled into its owner's library")
	}
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, opp, true, "after Beacon of Tomorrows")
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, (me+1)%len(g.Seats), false, "rotation resumes after the caster")
}

// A Beacon that fizzles (its only target gone) does not resolve, so it
// takes no turn and is NOT shuffled away: it goes to the graveyard
// (CR 608.2b).
func TestBeaconOfTomorrowsThatFizzlesGoesToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	opp := (me + 1) % len(g.Seats)
	spell := effectstest.CastCatalogSpell(t, g, "Beacon of Tomorrows", "Sorcery", effectstest.BeaconOfTomorrowsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[opp].ID}})
	g.WithWriteLock(func() { g.Seats[opp].Eliminated = true })
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Seats[me].Graveyard.Contains(spell) {
		t.Error("a fizzled Beacon goes to the graveyard")
	}
	if effectstest.LibraryHolds(g.Seats[me], spell) {
		t.Error("a fizzled Beacon must not be shuffled into the library")
	}
}

// Nexus of Fate takes the extra turn at instant speed and shuffles
// itself into the library as it resolves.
func TestNexusOfFateTakesAnExtraTurnAndShufflesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	spell := effectstest.CastCatalogSpell(t, g, "Nexus of Fate", "Instant", effectstest.NexusOfFateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Seats[me].Graveyard.Contains(spell) {
		t.Error("Nexus of Fate must not reach the graveyard")
	}
	if !effectstest.LibraryHolds(g.Seats[me], spell) {
		t.Error("Nexus of Fate should be shuffled into its owner's library")
	}
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me, true, "after Nexus of Fate")
}

// "From anywhere": a discarded Nexus of Fate is shuffled in, not
// binned, and so is one that is milled and one that is countered.
func TestNexusOfFateIsShuffledInFromAnywhere(t *testing.T) {
	t.Run("discard", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Nexus of Fate", TypeLine: "Instant",
			OracleID: effectstest.NexusOfFateOracle, Owner: me.ID, Controller: me.ID})
		effectstest.LengDiscard(t, g, me, id)
		if me.Graveyard.Contains(id) || !effectstest.LibraryHolds(me, id) {
			t.Errorf("discarded Nexus: graveyard=%v library=%v", me.Graveyard.Contains(id), effectstest.LibraryHolds(me, id))
		}
	})
	t.Run("mill", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Nexus of Fate", TypeLine: "Instant",
			OracleID: effectstest.NexusOfFateOracle, Owner: me.ID, Controller: me.ID})
		g.WithWriteLock(func() {
			if err := g.MillNForEffect(me.ID, 1); err != nil {
				t.Fatalf("MillNForEffect: %v", err)
			}
		})
		if me.Graveyard.Contains(id) || !effectstest.LibraryHolds(me, id) {
			t.Errorf("milled Nexus: graveyard=%v library=%v", me.Graveyard.Contains(id), effectstest.LibraryHolds(me, id))
		}
	})
	t.Run("countered", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		spell := effectstest.CastCatalogSpell(t, g, "Nexus of Fate", "Instant", effectstest.NexusOfFateOracle, nil)
		g.WithWriteLock(func() {
			if err := g.CounterTargetForEffect(spell); err != nil {
				t.Fatalf("CounterTargetForEffect: %v", err)
			}
		})
		if me.Graveyard.Contains(spell) || !effectstest.LibraryHolds(me, spell) {
			t.Errorf("countered Nexus: graveyard=%v library=%v", me.Graveyard.Contains(spell), effectstest.LibraryHolds(me, spell))
		}
	})
}

// Rise of the Eldrazi: destroy a permanent, a player draws four, an
// extra turn, and the spell is exiled. It cannot be countered.
func TestRiseOfTheEldraziDoesItAll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	opp := (me + 1) % len(g.Seats)
	victim := effectstest.PushCatalogPermanent(g, g.Seats[opp].ID, "Grizzly Bears", "Creature — Bear", "", false)
	before := g.Seats[opp].Hand.Size()
	spell := effectstest.CastCatalogSpell(t, g, "Rise of the Eldrazi", "Sorcery", effectstest.RiseOfTheEldraziOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: victim},
		{Kind: game.TargetPlayer, ID: g.Seats[opp].ID},
	})
	effectstest.PassPriorityAroundTable(t, g)

	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("the target permanent should be destroyed")
	}
	if got := g.Seats[opp].Hand.Size(); got != before+4 {
		t.Errorf("target player's hand = %d, want %d", got, before+4)
	}
	if g.Seats[me].Graveyard.Contains(spell) || !g.Exile.Contains(spell) {
		t.Errorf("Rise of the Eldrazi should be exiled (graveyard %v, exile %v)",
			g.Seats[me].Graveyard.Contains(spell), g.Exile.Contains(spell))
	}
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me, true, "after Rise of the Eldrazi")
}

func TestRiseOfTheEldraziCantBeCountered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	opp := (me + 1) % len(g.Seats)
	victim := effectstest.PushCatalogPermanent(g, g.Seats[opp].ID, "Grizzly Bears", "Creature — Bear", "", false)
	spell := effectstest.CastCatalogSpell(t, g, "Rise of the Eldrazi", "Sorcery", effectstest.RiseOfTheEldraziOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: victim},
		{Kind: game.TargetPlayer, ID: g.Seats[opp].ID},
	})
	g.WithWriteLock(func() { _ = g.CounterTargetForEffect(spell) })
	if g.Stack.Size() != 1 {
		t.Fatalf("Rise of the Eldrazi was countered: stack size %d", g.Stack.Size())
	}
}
