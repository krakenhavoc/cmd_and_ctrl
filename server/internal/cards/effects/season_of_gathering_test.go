package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const seasonOfGatheringOracle = "3ea060d8-449c-4a5e-8883-b45bd7707b5b"

func TestSeasonOfGatheringIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(seasonOfGatheringOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Season of Gathering: %+v %v", spec, ok)
	}
}

func castSeasonOfGathering(t *testing.T, g *game.Game, modes []int) error {
	t.Helper()
	_, err := castPawprintSpell(t, g, "Season of Gathering", "Sorcery", "{4}{G}{G}", "{C}{C}{C}{C}{G}{G}", seasonOfGatheringOracle, modes, nil)
	return err
}

func TestSeasonOfGatheringCounterThenDrawReadsTheNewPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small := pushVanillaCreature(g, me.ID, "Small", 1, 1)
	big := pushVanillaCreature(g, me.ID, "Big", 3, 3)
	seedLibrary(me, "a", "b", "c", "d", "e")
	hand := me.Hand.Size()
	if err := castSeasonOfGathering(t, g, []int{0, 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The creature is chosen as the spell resolves.
	answerChooseCards(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, small)
	if counterOn(g, small, game.CounterPlusOne) != 1 || !game.HasKeyword(&c, "vigilance") || !game.HasKeyword(&c, "trample") {
		t.Errorf("counters=%d abilities=%v, want a counter, vigilance and trample", counterOn(g, small, game.CounterPlusOne), c.Effective().Abilities)
	}
	b, _ := battlefieldCard(g, big)
	if game.HasKeyword(&b, "vigilance") {
		t.Error("the other creature gained vigilance")
	}
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3 (the greatest power)", got)
	}
}

func TestSeasonOfGatheringDestroysAllOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	rock := pushPermanentForTest(g, me.ID, "Rock", "", "Artifact")
	theirRock := pushPermanentForTest(g, opp.ID, "Their Rock", "", "Artifact")
	ring := pushPermanentForTest(g, me.ID, "Ring", "", "Enchantment")
	if err := castSeasonOfGathering(t, g, []int{1}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(theirRock) {
		t.Error("an artifact survived")
	}
	if !g.Battlefield.Contains(ring) {
		t.Error("the enchantment was destroyed")
	}
}

func TestSeasonOfGatheringRefusesMoreThanFivePawprints(t *testing.T) {
	g := newCatalogGame(t)
	if err := castSeasonOfGathering(t, g, []int{2, 2}); err == nil {
		t.Error("six pawprints were accepted")
	}
}
