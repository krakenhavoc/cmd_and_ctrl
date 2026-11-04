package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #4689: an alternative cost that BEHOLDS its cards (`behold` on the
// offer, Kindle the Inner Flame's three Elementals) spends none of them, so
// the bot charges nothing for which ones it names: beholding a 6/6 costs
// what beholding a 1/1 does. The same offer without `behold` (an exile)
// still charges the dearer card more.
func TestABeholdAlternativeCostChargesNothingForItsCards(t *testing.T) {
	for _, tc := range []struct {
		name    string
		beholds bool
	}{
		{"behold", true},
		{"exile", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sp := spell(cardID(1), 0, "Behold Spell", "{3}{R}")
			sp.AlternativeCosts = []protocol.AlternativeCostView{{
				Key:        "alt",
				Behold:     tc.beholds,
				PayOptions: &protocol.LegalTargetsView{Min: 1, Max: 1, Cards: []string{cardID(2), cardID(3)}},
			}}
			big := creature(cardID(2), 0, "Big", 6, 6)
			small := creature(cardID(3), 0, "Small", 1, 1)
			cast := func(pay string) legal.Move {
				m := castMove(t, 0, sp.InstanceID, "Cast for its alternative cost naming "+pay)
				m.Params = mustJSON(t, map[string]any{"instance_id": sp.InstanceID, "from_zone": "hand",
					"alternative_cost": "alt", "alt_cost_ids": []string{pay}})
				return m
			}
			v := newView([]protocol.PlayerView{newSeat(0, withHand(sp)), newSeat(1)},
				withBattlefield(big, small), withTurn(9, 1, "precombat_main"))
			in := input(0, v, passMove(0), cast(big.InstanceID), cast(small.InstanceID))
			pol := heuristic.New()
			gap := rankValue(t, pol, in, in.Moves[2].Label) - rankValue(t, pol, in, in.Moves[1].Label)
			if tc.beholds && !nearly(gap, 0) {
				t.Errorf("beholding the 6/6 costs %.3f more than beholding the 1/1, want the same (nothing)", gap)
			}
			if !tc.beholds && gap <= 0 {
				t.Errorf("exiling the 6/6 is priced %.3f against the 1/1, want it dearer", gap)
			}
		})
	}
}
