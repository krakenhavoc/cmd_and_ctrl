package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ripples_of_potential_test.go — #3930. "Proliferate, then choose any
// number of permanents you control that had a counter put on them this
// way. Those permanents phase out."

const (
	ripplesOfPotentialOracle = "45bcebd0-5e93-4433-b92f-b4c161d04bd6"
	vorinclexRaiderOracle    = "5a3fdf5a-bff8-4896-b288-3f43f9a72d9b"
)

// ripplesChoosePrompt is the open choose-cards prompt Ripples asks, or nil.
func ripplesChoosePrompt(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards {
			return c
		}
	}
	return nil
}

func castRipples(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Ripples of Potential", "Instant", ripplesOfPotentialOracle, nil)
	passPriorityAroundTable(t, g)
}

func TestRipplesOfPotentialPhasesOutChosenPermanentsThatGotACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := top100Opponent(g)
	a := seedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	b := seedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	notChosen := seedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	theirs := seedCreatureWithCounters(g, opp.ID, map[string]int{game.CounterMinusOne: 1})

	castRipples(t, g)
	answerProliferate(t, g, a, b, theirs)

	c := ripplesChoosePrompt(g)
	if c == nil {
		t.Fatal("no choose prompt after the proliferate")
	}
	if want := []uuid.UUID{a, b}; !slices.Equal(c.ChooseCards, want) {
		t.Fatalf("offered %v, want your two permanents that got a counter %v", c.ChooseCards, want)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !phasedOutInCatalogGame(g, a) {
		t.Error("the chosen permanent did not phase out")
	}
	for _, id := range []uuid.UUID{b, notChosen, theirs} {
		if phasedOutInCatalogGame(g, id) {
			t.Errorf("%s phased out without being chosen", id)
		}
	}
	if got := countersOn(g, b, game.CounterPlusOne); got != 2 {
		t.Errorf("the proliferate gave b %d counters, want 2", got)
	}
}

// TestRipplesOfPotentialSkipsAPermanentWhoseCounterWasReplacedAway: an
// opponent's Vorinclex halves the one counter you would put to none, so
// that permanent did not have a counter put on it "this way".
func TestRipplesOfPotentialSkipsAPermanentWhoseCounterWasReplacedAway(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := top100Opponent(g)
	pushCatalogPermanent(g, opp.ID, "Vorinclex, Monstrous Raider", "Legendary Creature — Phyrexian Praetor", vorinclexRaiderOracle, false)
	halved := seedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})

	castRipples(t, g)
	answerProliferate(t, g, halved)
	if got := countersOn(g, halved, game.CounterPlusOne); got != 1 {
		t.Fatalf("Vorinclex: %d counters, want 1 (half of one is none)", got)
	}
	if c := ripplesChoosePrompt(g); c != nil {
		t.Fatalf("offered %v though nothing got a counter", c.ChooseCards)
	}
}

// TestRipplesOfPotentialWaitsForACounterOrderingPrompt: a Doubling Season
// beside a Hardened Scales pauses the placement on a CR 616 ordering
// prompt, and the choice waits until the counter has landed.
func TestRipplesOfPotentialWaitsForACounterOrderingPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hs := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	bear := seedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})

	castRipples(t, g)
	answerProliferate(t, g, bear)
	if c := ripplesChoosePrompt(g); c != nil {
		t.Fatal("the choice opened while the counter was still owed")
	}
	var order *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceReplacementOrder {
			order = c
		}
	}
	if order == nil {
		t.Fatal("no CR 616 ordering prompt")
	}
	hsEff, dsEff := replacementIDsForSources(t, g, order.ReplacementEffectIDs, hs, ds)
	if err := g.ResolveReplacementOrder(order.ID, me.ID, []game.ReplacementEffectID{hsEff, dsEff}); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if got := countersOn(g, bear, game.CounterPlusOne); got != 5 {
		t.Fatalf("counters = %d, want 5 (1 + (1+1)*2)", got)
	}
	c := ripplesChoosePrompt(g)
	if c == nil || !slices.Equal(c.ChooseCards, []uuid.UUID{bear}) {
		t.Fatalf("after the counters landed: prompt %v, want the bear offered", c)
	}
}

// TestRipplesOfPotentialUnderTekuthalOffersBothPasses: "proliferate
// twice" is one action, so a permanent either pass gave a counter to
// had a counter put on it "this way".
func TestRipplesOfPotentialUnderTekuthalOffersBothPasses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushTekuthal(g, me.ID)
	a := pushCounterCreature(g, me.ID, "A", game.CounterPlusOne, 1)
	b := pushCounterCreature(g, me.ID, "B", game.CounterPlusOne, 1)

	castRipples(t, g)
	answerProliferate(t, g, a)
	if c := ripplesChoosePrompt(g); c != nil {
		t.Fatal("the choice opened between the two proliferates")
	}
	answerProliferate(t, g, a, b)
	c := ripplesChoosePrompt(g)
	if c == nil || !slices.Equal(c.ChooseCards, []uuid.UUID{a, b}) {
		t.Fatalf("offered %v, want both creatures, each once", c)
	}
	if got := countersOn(g, a, game.CounterPlusOne); got != 3 {
		t.Errorf("A has %d counters, want 3 (chosen twice)", got)
	}
}

// TestProliferateReportsPermanentsAndPlayers: the engine's report names
// players as well as permanents.
func TestProliferateReportsPermanentsAndPlayers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := top100Opponent(g)
	bear := seedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterByForEffect(me.ID, opp.ID, game.CounterPoison, 1); err != nil {
			t.Fatalf("poison: %v", err)
		}
	})
	var got game.Proliferated
	g.WithWriteLock(func() {
		if err := g.ProliferateChoosingThenForEffect(me.ID, uuid.Nil, func(_ *game.Game, p game.Proliferated) error {
			got = p
			return nil
		}); err != nil {
			t.Fatalf("proliferate: %v", err)
		}
	})
	answerProliferate(t, g, bear, opp.ID)
	if !slices.Equal(got.Permanents, []uuid.UUID{bear}) || !slices.Equal(got.Players, []uuid.UUID{opp.ID}) {
		t.Fatalf("report = %+v, want the bear and the poisoned opponent", got)
	}
}
