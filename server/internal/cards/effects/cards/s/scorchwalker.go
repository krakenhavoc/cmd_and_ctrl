package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scorchwalker — Creature — Elemental {3}{R}, 5/1 (Gatecrash, #5565):
//
//	"Bloodrush — {1}{R}{R}, Discard this card: Target attacking creature gets
//	 +5/+1 until end of turn."
//
// Bloodrush is Pyrewild Shaman's shape (gtcBloodrush): an activated
// ability that works from the hand, with discarding this card as part of
// the cost. A target that stopped attacking or left is illegal (CR
// 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "566d6fcd-dd50-4135-a912-ec923fc433df",
		Name:         "Scorchwalker",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.GtcBloodrush("Scorchwalker", "{1}{R}{R}", 5, 1, "")},
	})
}
