package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0135 §3, owner decision 6: an awaken cast is priced as the spell's
// own purpose plus a hasty N/N creature, at AwakenLandShare of what the
// board prices one, so the bot pays the awaken cost when it can. A bigger
// N is worth more, and the baseline (AwakenLandShare 0) prices awaken at
// nothing.
func TestAwakenIsPricedAsAHastyBodyBesideTheSpell(t *testing.T) {
	build := func(n int) (heuristicInput func(cfg heuristic.Config) (hard, awaken float64)) {
		return func(cfg heuristic.Config) (float64, float64) {
			sp := spell(cardID(1), 0, "Coastal Discovery", "{3}{U}")
			sp.Purpose = &protocol.PurposeView{Draws: 2}
			sp.AlternativeCosts = []protocol.AlternativeCostView{{
				Key: "awaken", Label: "Awaken", ManaCost: "{5}{U}",
				Purpose: &protocol.PurposeView{AwakenLand: n},
			}}
			land := land(cardID(2), 0)
			hard := castMove(t, 0, sp.InstanceID, "Cast Coastal Discovery")
			hard.Params = mustJSON(t, map[string]any{"instance_id": sp.InstanceID, "from_zone": "hand"})
			aw := castMove(t, 0, sp.InstanceID, "Cast Coastal Discovery for its awaken cost")
			aw.Params = mustJSON(t, map[string]any{"instance_id": sp.InstanceID, "from_zone": "hand",
				"alternative_cost": "awaken", "targets": []map[string]string{{"kind": "card", "id": land.InstanceID}}})
			v := newView([]protocol.PlayerView{newSeat(0, withHand(sp)), newSeat(1)},
				withBattlefield(land), withTurn(9, 0, "precombat_main"))
			in := input(0, v, passMove(0), hard, aw)
			pol := heuristic.NewWithConfig(cfg)
			return rankValue(t, pol, in, hard.Label), rankValue(t, pol, in, aw.Label)
		}
	}
	hard, four := build(4)(heuristic.DefaultConfig())
	if four <= hard {
		t.Errorf("awaken 4 priced %.3f, the hard cast %.3f; want awaken above", four, hard)
	}
	_, two := build(2)(heuristic.DefaultConfig())
	if two >= four || two <= hard {
		t.Errorf("awaken 2 priced %.3f; want it between the hard cast %.3f and awaken 4 %.3f", two, hard, four)
	}
	bh, ba := build(4)(heuristic.BaselineConfig())
	if !nearly(bh, ba) {
		t.Errorf("baseline prices awaken at %.3f against %.3f for the hard cast; want the same", ba, bh)
	}
}
