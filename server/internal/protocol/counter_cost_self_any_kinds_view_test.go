package protocol

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_cost_self_any_kinds_view_test.go — the wire half of #3987,
// "Remove two counters from Ghost-Spider". The client reads it as the
// among picker with the source as its one option: self AND among, no
// printed kind, one row per kind the source holds (down to a floor of
// one, since a single shield counter is a legal part of the payment).
func TestCounterCostViewSelfAnyKindsListsTheSourcesKinds(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	src := seatCounterCostSource(g, owner, &game.CounterRemovalCost{N: 2, Among: true})
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == src {
			g.Battlefield.Cards[i].Counters = map[string]int{
				game.CounterPlusOne: 2, game.CounterShield: 1,
			}
		}
	}
	other := seatCreature(g, owner, "Bear", 2, false, false)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == other {
			g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 5}
		}
	}
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, src).ActivatedAbilities[0]
	if ab.CounterCostN != 2 || !ab.CounterCostAmong || !ab.CounterCostSelf || ab.CounterCostKind != "" {
		t.Errorf("shape: n=%d among=%v self=%v kind=%q, want 2 / true / true / empty",
			ab.CounterCostN, ab.CounterCostAmong, ab.CounterCostSelf, ab.CounterCostKind)
	}
	if len(ab.CounterCostOptions) != 1 || ab.CounterCostOptions[0].CardID != src.String() {
		t.Fatalf("counter_cost_options = %+v, want the source alone", ab.CounterCostOptions)
	}
	k := ab.CounterCostOptions[0].Kinds
	if len(k) != 2 || k[0].Kind != game.CounterPlusOne || k[1].Kind != game.CounterShield || k[1].Count != 1 {
		t.Errorf("kinds = %+v, want +1/+1 (2) then shield (1)", k)
	}
}
