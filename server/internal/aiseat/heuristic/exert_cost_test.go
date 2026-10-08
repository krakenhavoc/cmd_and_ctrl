package heuristic_test

import (
	"context"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// exert_cost_test.go — ADR 0130 §4 and §9: an activation whose cost
// exerts its source pays for the source's next untap, under PriceExert,
// and nothing changes for heuristic-baseline.

// exertCostValues ranks the same activation twice, once plain and once
// with cost.exert, and returns both values.
func exertCostValues(t *testing.T, p *heuristic.Policy, v protocol.GameView, src string) (plain, exert float64) {
	t.Helper()
	const label = "Steward: {T}, Exert this creature: Create a 1/1 Warrior"
	in := input(0, v, passMove(0),
		activateMove(t, 0, src, label, nil),
		activateMove(t, 0, src, label+" ", &legal.MoveCost{Exert: true}),
	)
	for _, c := range p.Rank(context.Background(), in) {
		switch c.Index {
		case 1:
			plain = c.Value
		case 2:
			exert = c.Value
		}
	}
	return plain, exert
}

func TestExertCostIsChargedForTheNextUntap(t *testing.T) {
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(30)), newSeat(1, withLife(30))},
		withBattlefield(
			creature(cardID(10), 0, "Steward", 3, 3),
			creature(cardID(20), 1, "Raider", 2, 2),
		),
		withTurn(9, 0, "precombat_main"),
	)
	plain, exert := exertCostValues(t, heuristic.New(), v, cardID(10))
	if exert >= plain {
		t.Errorf("exert %.2f, plain %.2f: the exert should cost the next untap", exert, plain)
	}
	plain, exert = exertCostValues(t, heuristic.NewWithConfig(heuristic.BaselineConfig()), v, cardID(10))
	if exert != plain {
		t.Errorf("baseline: exert %.2f, plain %.2f; heuristic-baseline does not price an exert", exert, plain)
	}
}

// A source that won't untap anyway pays nothing more for the exert.
func TestExertCostIsFreeWhenItWontUntapAnyway(t *testing.T) {
	steward := creature(cardID(10), 0, "Steward", 3, 3)
	steward.NoUntap = &protocol.NoUntapView{Static: true}
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(30)), newSeat(1, withLife(30))},
		withBattlefield(steward, creature(cardID(20), 1, "Raider", 2, 2)),
		withTurn(9, 0, "precombat_main"),
	)
	plain, exert := exertCostValues(t, heuristic.New(), v, cardID(10))
	if exert != plain {
		t.Errorf("exert %.2f, plain %.2f; it won't untap anyway", exert, plain)
	}
}
