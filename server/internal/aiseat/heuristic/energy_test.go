package heuristic_test

import (
	"context"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// energy_test.go — ADR 0129 §7 (owner decision 5): energy is priced at
// a flat Weights.Energy per counter, charged when a move spends it
// (legal.MoveCost.Energy) and credited when a declared purpose gives it.

// Spending: the same activation is worth Weights.Energy less per
// counter it removes, so eight energy is not read as free.
func TestEnergySpentIsPriced(t *testing.T) {
	w := heuristic.DefaultWeights()
	if w.Energy <= 0 {
		t.Fatalf("Weights.Energy = %v, want a positive price", w.Energy)
	}
	src := protocol.CardView{
		InstanceID: cardID(1), Name: "Aethertorch Renegade", TypeLine: "Creature — Human Rogue",
		Owner: seatID(0).String(), Controller: seatID(0).String(), KnownByYou: true,
	}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(src, land(cardID(10), 0)))
	cheap, dear := "Renegade (two energy)", "Renegade (eight energy)"
	in := input(0, v,
		passMove(0),
		activateMove(t, 0, cardID(1), cheap, &legal.MoveCost{Energy: 2}),
		activateMove(t, 0, cardID(1), dear, &legal.MoveCost{Energy: 8}),
	)
	value := map[int]float64{}
	for _, c := range heuristic.New().Rank(context.Background(), in) {
		value[c.Index] = c.Value
	}
	if got, want := value[1]-value[2], 6*w.Energy; !nearly(got, want) {
		t.Errorf("six more energy cost %.3f, want %.3f", got, want)
	}
}

// Getting: a declared `energy` purpose adds Weights.Energy per counter
// to the spell's price.
func TestEnergyGainedIsPriced(t *testing.T) {
	pol := heuristic.New()
	price := func(p *protocol.PurposeView) float64 {
		t.Helper()
		c := sorcery(cardID(1), 0, "Glimmer of Genius", "{3}{U}", p)
		v := newView([]protocol.PlayerView{newSeat(0, withHand(c)), newSeat(1)},
			withBattlefield(manaLands(4, 0, 100)...), withTurn(3, 0, "precombat_main"))
		return rankValue(t, pol, input(0, v, passMove(0), castMove(t, 0, c.InstanceID, "Cast "+c.Name)), "Cast "+c.Name)
	}
	with := price(&protocol.PurposeView{Draws: 2, Energy: 2})
	without := price(&protocol.PurposeView{Draws: 2})
	if got, want := with-without, 2*heuristic.DefaultWeights().Energy; !nearly(got, want) {
		t.Errorf("two energy added %.3f, want %.3f", got, want)
	}
}

// The baseline is the heuristic before ADR 0129: energy is free.
func TestBaselinePricesEnergyAtNothing(t *testing.T) {
	if e := heuristic.BaselineConfig().Weights.Energy; e != 0 {
		t.Errorf("baseline Weights.Energy = %v, want 0", e)
	}
}
