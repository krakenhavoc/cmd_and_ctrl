package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jodi_big_spells_test.go — S58 PR 5, the Jodi deck's big spells and
// the Eldrazi that need no annihilator. Each card is driven through the
// real cast, trigger or loyalty path and its result read off the game.

const (
	beaconOfTomorrowsOracle   = "85909caa-d2ad-4487-b9b7-6ac60a14b833"
	bringerBlackDawnOracle    = "a9f98ad5-e8d4-4142-a19f-3368b85d1c95"
	bringerBlueDawnOracle     = "7d4676fa-4bfe-43b4-9241-c2a33214852a"
	confluxOracle             = "32ebb029-9f03-49c9-b9bb-cb1954e2a324"
	damiaOracle               = "eb7ae3e8-0489-4f3a-8343-7e7ec2fe3f01"
	emergentUltimatumOracle   = "0eecdfb3-3b05-4051-a660-060ff6df80ef"
	nexusOfFateOracle         = "6c1d22d4-f28e-4041-a9b6-1575e8929b61"
	nicolBolasOracle          = "cb33e07b-9599-4979-991f-df0c43ddac31"
	riseOfTheEldraziOracle    = "982f70af-077f-40df-b2fe-7c80ebcc7712"
	sheoldredWhisperingOracle = "9218b56d-aaec-482f-99e9-d95d227bfe25"
	sunbirdsInvocationOracle  = "10d482fd-e034-4187-b148-51fe15885360"
	uginsBindingOracle        = "ed622e71-f348-4e46-8fb9-05aae430983a"
	ulamogCeaselessOracle     = "0bfa4512-e35a-4c93-b324-80ec659f5a97"
	vorinclexVoiceOracle      = "dbf0ad03-ab31-49d2-89b1-05b45948a61f"
	zendikarResurgentOracle   = "a0af158b-f42a-4022-a438-e0a2e1fd6d9e"
)

// libraryHolds reports whether `id` is somewhere in `p`'s library.
func libraryHolds(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Library.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// libraryIndex is the card's position in the library (the top is the
// last element), or -1.
func libraryIndex(p *game.Player, id uuid.UUID) int {
	for i, c := range p.Library.Cards {
		if c.InstanceID == id {
			return i
		}
	}
	return -1
}

// --- extra turns ------------------------------------------------------

// Beacon of Tomorrows gives the TARGET player the turn, and the card
// is shuffled into its owner's library rather than left in the
// graveyard.
func TestBeaconOfTomorrowsGivesTheTargetAnExtraTurnAndShufflesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	opp := (me + 2) % len(g.Seats)
	spell := castCatalogSpell(t, g, "Beacon of Tomorrows", "Sorcery", beaconOfTomorrowsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[opp].ID}})
	passPriorityAroundTable(t, g)

	if g.Seats[me].Graveyard.Contains(spell) {
		t.Error("Beacon of Tomorrows must not be left in the graveyard")
	}
	if !libraryHolds(g.Seats[me], spell) {
		t.Error("Beacon of Tomorrows should have been shuffled into its owner's library")
	}
	endTurn(t, g)
	assertTurnOf(t, g, opp, true, "after Beacon of Tomorrows")
	endTurn(t, g)
	assertTurnOf(t, g, (me+1)%len(g.Seats), false, "rotation resumes after the caster")
}

// A Beacon that fizzles (its only target gone) does not resolve, so it
// takes no turn and is NOT shuffled away: it goes to the graveyard
// (CR 608.2b).
func TestBeaconOfTomorrowsThatFizzlesGoesToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	opp := (me + 1) % len(g.Seats)
	spell := castCatalogSpell(t, g, "Beacon of Tomorrows", "Sorcery", beaconOfTomorrowsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[opp].ID}})
	g.WithWriteLock(func() { g.Seats[opp].Eliminated = true })
	passPriorityAroundTable(t, g)

	if !g.Seats[me].Graveyard.Contains(spell) {
		t.Error("a fizzled Beacon goes to the graveyard")
	}
	if libraryHolds(g.Seats[me], spell) {
		t.Error("a fizzled Beacon must not be shuffled into the library")
	}
}

// Nexus of Fate takes the extra turn at instant speed and shuffles
// itself into the library as it resolves.
func TestNexusOfFateTakesAnExtraTurnAndShufflesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	spell := castCatalogSpell(t, g, "Nexus of Fate", "Instant", nexusOfFateOracle, nil)
	passPriorityAroundTable(t, g)

	if g.Seats[me].Graveyard.Contains(spell) {
		t.Error("Nexus of Fate must not reach the graveyard")
	}
	if !libraryHolds(g.Seats[me], spell) {
		t.Error("Nexus of Fate should be shuffled into its owner's library")
	}
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "after Nexus of Fate")
}

// "From anywhere": a discarded Nexus of Fate is shuffled in, not
// binned, and so is one that is milled and one that is countered.
func TestNexusOfFateIsShuffledInFromAnywhere(t *testing.T) {
	t.Run("discard", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Nexus of Fate", TypeLine: "Instant",
			OracleID: nexusOfFateOracle, Owner: me.ID, Controller: me.ID})
		lengDiscard(t, g, me, id)
		if me.Graveyard.Contains(id) || !libraryHolds(me, id) {
			t.Errorf("discarded Nexus: graveyard=%v library=%v", me.Graveyard.Contains(id), libraryHolds(me, id))
		}
	})
	t.Run("mill", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Nexus of Fate", TypeLine: "Instant",
			OracleID: nexusOfFateOracle, Owner: me.ID, Controller: me.ID})
		g.WithWriteLock(func() {
			if err := g.MillNForEffect(me.ID, 1); err != nil {
				t.Fatalf("MillNForEffect: %v", err)
			}
		})
		if me.Graveyard.Contains(id) || !libraryHolds(me, id) {
			t.Errorf("milled Nexus: graveyard=%v library=%v", me.Graveyard.Contains(id), libraryHolds(me, id))
		}
	})
	t.Run("countered", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		spell := castCatalogSpell(t, g, "Nexus of Fate", "Instant", nexusOfFateOracle, nil)
		g.WithWriteLock(func() {
			if err := g.CounterTargetForEffect(spell); err != nil {
				t.Fatalf("CounterTargetForEffect: %v", err)
			}
		})
		if me.Graveyard.Contains(spell) || !libraryHolds(me, spell) {
			t.Errorf("countered Nexus: graveyard=%v library=%v", me.Graveyard.Contains(spell), libraryHolds(me, spell))
		}
	})
}

// Rise of the Eldrazi: destroy a permanent, a player draws four, an
// extra turn, and the spell is exiled. It cannot be countered.
func TestRiseOfTheEldraziDoesItAll(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	opp := (me + 1) % len(g.Seats)
	victim := pushCatalogPermanent(g, g.Seats[opp].ID, "Grizzly Bears", "Creature — Bear", "", false)
	before := g.Seats[opp].Hand.Size()
	spell := castCatalogSpell(t, g, "Rise of the Eldrazi", "Sorcery", riseOfTheEldraziOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: victim},
		{Kind: game.TargetPlayer, ID: g.Seats[opp].ID},
	})
	passPriorityAroundTable(t, g)

	if _, ok := battlefieldCard(g, victim); ok {
		t.Error("the target permanent should be destroyed")
	}
	if got := g.Seats[opp].Hand.Size(); got != before+4 {
		t.Errorf("target player's hand = %d, want %d", got, before+4)
	}
	if g.Seats[me].Graveyard.Contains(spell) || !g.Exile.Contains(spell) {
		t.Errorf("Rise of the Eldrazi should be exiled (graveyard %v, exile %v)",
			g.Seats[me].Graveyard.Contains(spell), g.Exile.Contains(spell))
	}
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "after Rise of the Eldrazi")
}

func TestRiseOfTheEldraziCantBeCountered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	opp := (me + 1) % len(g.Seats)
	victim := pushCatalogPermanent(g, g.Seats[opp].ID, "Grizzly Bears", "Creature — Bear", "", false)
	spell := castCatalogSpell(t, g, "Rise of the Eldrazi", "Sorcery", riseOfTheEldraziOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: victim},
		{Kind: game.TargetPlayer, ID: g.Seats[opp].ID},
	})
	g.WithWriteLock(func() { _ = g.CounterTargetForEffect(spell) })
	if g.Stack.Size() != 1 {
		t.Fatalf("Rise of the Eldrazi was countered: stack size %d", g.Stack.Size())
	}
}
