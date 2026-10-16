package heuristic_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// tap_any_number_test.go — #3961: Burn at the Stake taps any number of
// creatures and deals three times that many. The enumerator offers one
// cast per count; the purpose entry declares damage_per_tapped, so the
// policy reads the count each move names and weighs the damage against
// the creatures it taps: the smallest count that kills the target wins,
// a smaller one kills nothing, and a bigger one taps blockers for nothing.
// Tapping none deals nothing at all and is declined.

func TestBurnAtTheStakeTapsJustEnoughToKill(t *testing.T) {
	burn := spell(cardID(1), 0, "Burn at the Stake", "{2}{R}{R}{R}")
	burn.AdditionalCost = &protocol.AdditionalCostView{
		TapOthersLabel:   "any number of untapped creatures you control",
		TapOthersOptions: &protocol.LegalTargetsView{Min: 0, Max: 3, Cards: []string{cardID(2), cardID(3), cardID(4)}},
	}
	burn.Purpose = targetEntries(protocol.TargetPurposeView{Slot: 0, DamagePerTapped: 3})
	mine := []protocol.CardView{
		creature(cardID(2), 0, "Bear A", 2, 2),
		creature(cardID(3), 0, "Bear B", 2, 2),
		creature(cardID(4), 0, "Bear C", 2, 2),
	}
	ogre := creature(cardID(5), 1, "Ogre", 5, 5)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(burn)), newSeat(1)},
		withBattlefield(append(mine, ogre)...), withTurn(3, 0, "precombat_main"))

	moves := []legal.Move{passMove(0)}
	for n := 0; n <= 3; n++ {
		taps := []string{}
		for i := 0; i < n; i++ {
			taps = append(taps, mine[i].InstanceID)
		}
		m := castMove(t, 0, burn.InstanceID, fmt.Sprintf("Burn at the Stake tapping %d", n))
		m.Params = mustJSON(t, map[string]any{
			"instance_id": burn.InstanceID, "from_zone": "hand", "teamwork_ids": taps,
			"targets": []map[string]string{cardTarget(ogre.InstanceID)},
		})
		moves = append(moves, m)
	}
	in := input(0, v, moves...)
	best, bestV := -1, 0.0
	for _, c := range heuristic.New().Rank(context.Background(), in) {
		if c.Index == 1 && c.Value >= 0 {
			t.Errorf("tapping none (no damage) is valued %.2f, want below passing", c.Value)
		}
		if best < 0 || c.Value > bestV {
			best, bestV = c.Index, c.Value
		}
	}
	if got := in.Moves[best].Label; got != "Burn at the Stake tapping 2" {
		t.Errorf("best move = %q, want tapping 2 (six damage for the 5/5)", got)
	}
}
