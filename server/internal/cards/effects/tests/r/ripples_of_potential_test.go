package testsr

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestRipplesOfPotentialPhasesOutChosenPermanentsThatGotACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Top100Opponent(g)
	a := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	b := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	notChosen := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	theirs := effectstest.SeedCreatureWithCounters(g, opp.ID, map[string]int{game.CounterMinusOne: 1})

	effectstest.CastRipples(t, g)
	effectstest.AnswerProliferate(t, g, a, b, theirs)

	c := effectstest.RipplesChoosePrompt(g)
	if c == nil {
		t.Fatal("no choose prompt after the proliferate")
	}
	if want := []uuid.UUID{a, b}; !slices.Equal(c.ChooseCards, want) {
		t.Fatalf("offered %v, want your two permanents that got a counter %v", c.ChooseCards, want)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !effectstest.PhasedOutInCatalogGame(g, a) {
		t.Error("the chosen permanent did not phase out")
	}
	for _, id := range []uuid.UUID{b, notChosen, theirs} {
		if effectstest.PhasedOutInCatalogGame(g, id) {
			t.Errorf("%s phased out without being chosen", id)
		}
	}
	if got := effectstest.CountersOn(g, b, game.CounterPlusOne); got != 2 {
		t.Errorf("the proliferate gave b %d counters, want 2", got)
	}
}

// TestRipplesOfPotentialSkipsAPermanentWhoseCounterWasReplacedAway: an
// opponent's Vorinclex halves the one counter you would put to none, so
// that permanent did not have a counter put on it "this way".
func TestRipplesOfPotentialSkipsAPermanentWhoseCounterWasReplacedAway(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Top100Opponent(g)
	effectstest.PushCatalogPermanent(g, opp.ID, "Vorinclex, Monstrous Raider", "Legendary Creature — Phyrexian Praetor", effectstest.VorinclexRaiderOracle, false)
	halved := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})

	effectstest.CastRipples(t, g)
	effectstest.AnswerProliferate(t, g, halved)
	if got := effectstest.CountersOn(g, halved, game.CounterPlusOne); got != 1 {
		t.Fatalf("Vorinclex: %d counters, want 1 (half of one is none)", got)
	}
	if c := effectstest.RipplesChoosePrompt(g); c != nil {
		t.Fatalf("offered %v though nothing got a counter", c.ChooseCards)
	}
}

// TestRipplesOfPotentialWaitsForACounterOrderingPrompt: a Doubling Season
// beside a Hardened Scales pauses the placement on a CR 616 ordering
// prompt, and the choice waits until the counter has landed.
func TestRipplesOfPotentialWaitsForACounterOrderingPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hs := effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	ds := effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	bear := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})

	effectstest.CastRipples(t, g)
	effectstest.AnswerProliferate(t, g, bear)
	if c := effectstest.RipplesChoosePrompt(g); c != nil {
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
	hsEff, dsEff := effectstest.ReplacementIDsForSources(t, g, order.ReplacementEffectIDs, hs, ds)
	if err := g.ResolveReplacementOrder(order.ID, me.ID, []game.ReplacementEffectID{hsEff, dsEff}); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 5 {
		t.Fatalf("counters = %d, want 5 (1 + (1+1)*2)", got)
	}
	c := effectstest.RipplesChoosePrompt(g)
	if c == nil || !slices.Equal(c.ChooseCards, []uuid.UUID{bear}) {
		t.Fatalf("after the counters landed: prompt %v, want the bear offered", c)
	}
}

// TestRipplesOfPotentialUnderTekuthalOffersBothPasses: "proliferate
// twice" is one action, so a permanent either pass gave a counter to
// had a counter put on it "this way".
func TestRipplesOfPotentialUnderTekuthalOffersBothPasses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushTekuthal(g, me.ID)
	a := effectstest.PushCounterCreature(g, me.ID, "A", game.CounterPlusOne, 1)
	b := effectstest.PushCounterCreature(g, me.ID, "B", game.CounterPlusOne, 1)

	effectstest.CastRipples(t, g)
	effectstest.AnswerProliferate(t, g, a)
	if c := effectstest.RipplesChoosePrompt(g); c != nil {
		t.Fatal("the choice opened between the two proliferates")
	}
	effectstest.AnswerProliferate(t, g, a, b)
	c := effectstest.RipplesChoosePrompt(g)
	if c == nil || !slices.Equal(c.ChooseCards, []uuid.UUID{a, b}) {
		t.Fatalf("offered %v, want both creatures, each once", c)
	}
	if got := effectstest.CountersOn(g, a, game.CounterPlusOne); got != 3 {
		t.Errorf("A has %d counters, want 3 (chosen twice)", got)
	}
}

func TestProliferateReportsPermanentsAndPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Top100Opponent(g)
	bear := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterByForEffect(me.ID, opp.ID, game.CounterPoison, 1); err != nil {
			t.Fatalf("poison: %v", err)
		}
	})
	effectstest.TestProliferatedHeard = nil
	g.WithWriteLock(func() {
		if err := g.ProliferateChoosingThenKeyedForEffect(me.ID, uuid.Nil, game.KeyedTail{Ref: effectstest.TestProliferatedTail}); err != nil {
			t.Fatalf("proliferate: %v", err)
		}
	})
	effectstest.AnswerProliferate(t, g, bear, opp.ID)
	if len(effectstest.TestProliferatedHeard) != 1 {
		t.Fatalf("the keyed tail ran %d times, want once", len(effectstest.TestProliferatedHeard))
	}
	got := effectstest.TestProliferatedHeard[0]
	if !slices.Equal(got.Permanents, []uuid.UUID{bear}) || !slices.Equal(got.Players, []uuid.UUID{opp.ID}) {
		t.Fatalf("report = %+v, want the bear and the poisoned opponent", got)
	}
}
