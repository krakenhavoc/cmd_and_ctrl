package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b16_helpers.go — shared helpers for Midnight Hunt Commander batch 16.

// s82MidTotalPowerRuleMax is the largest total-power ceiling that has its
// own registered set rule; a ceiling above it that some set of cards can
// still exceed is not representable, and no graveyard holds that much.
const S82MidTotalPowerRuleMax = 100

// s82MidTotalPowerKey is the on-disk key of the pick "total power N or
// less" for one Moorland Rescuer ceiling. Never renamed, never reused.
func S82MidTotalPowerKey(n int) string {
	return fmt.Sprintf("choose-cards/moorland-rescuer-total-power-%d", n)
}

// s82MidTotalPowerAtMost is the set rule: the picked cards' powers sum to
// `limit` or less.
func S82MidTotalPowerAtMost(limit int) game.SetRuleFunc {
	return func(cards []game.Card) bool {
		total := 0
		for _, c := range cards {
			total += c.PowerForComparison()
		}
		return total <= limit
	}
}
