package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0135 §1 (#2030): an alternative cost that TAPS its permanents
// (tap_options on the offer) spends none of them, so the bot charges each
// what tapping it costs, not what it is worth. Outside the bot's own
// precombat main a 6/6 and a 1/1 cost the same to tap; to sacrifice, the
// 6/6 is dearer.
func TestATapAlternativeCostPricesTappingNotLosing(t *testing.T) {
	for _, tc := range []struct {
		name string
		taps bool
	}{
		{"tap", true},
		{"sacrifice", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sp := spell(cardID(1), 0, "Tap Spell", "{3}{W}")
			offer := protocol.AlternativeCostView{Key: "alt"}
			if tc.taps {
				offer.TapOptions = &protocol.LegalTargetsView{Min: 1, Max: 1, Cards: []string{cardID(2), cardID(3)}}
			} else {
				offer.SacrificeOptions = &protocol.LegalTargetsView{Min: 1, Max: 1, Cards: []string{cardID(2), cardID(3)}}
			}
			sp.AlternativeCosts = []protocol.AlternativeCostView{offer}
			big := creature(cardID(2), 0, "Big", 6, 6)
			small := creature(cardID(3), 0, "Small", 1, 1)
			cast := func(pay string) legal.Move {
				m := castMove(t, 0, sp.InstanceID, "Cast for its alternative cost paying "+pay)
				m.Params = mustJSON(t, map[string]any{"instance_id": sp.InstanceID, "from_zone": "hand",
					"alternative_cost": "alt", "alt_cost_ids": []string{pay}})
				return m
			}
			v := newView([]protocol.PlayerView{newSeat(0, withHand(sp)), newSeat(1)},
				withBattlefield(big, small), withTurn(9, 1, "precombat_main"))
			in := input(0, v, passMove(0), cast(big.InstanceID), cast(small.InstanceID))
			pol := heuristic.New()
			gap := rankValue(t, pol, in, in.Moves[2].Label) - rankValue(t, pol, in, in.Moves[1].Label)
			if tc.taps && !nearly(gap, 0) {
				t.Errorf("tapping the 6/6 costs %.3f more than tapping the 1/1, want the same price", gap)
			}
			if !tc.taps && gap <= 0 {
				t.Errorf("sacrificing the 6/6 is priced %.3f against the 1/1, want it dearer", gap)
			}
			// The enumerator's order reads the same price.
			order := pol.CostFuelPrice(in)
			for _, id := range []string{big.InstanceID, small.InstanceID} {
				uid := uuid.MustParse(id)
				tapPrice := order(legal.TargetCandidate{ID: uid, Tap: true})
				if spend := order(legal.TargetCandidate{ID: uid}); tapPrice >= spend {
					t.Errorf("%s: tapping is priced %.3f, spending it %.3f; tapping must be cheaper", id, tapPrice, spend)
				}
			}
		})
	}
}
