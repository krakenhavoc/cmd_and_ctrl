package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serra's Blessing — Enchantment {1}{W} (Weatherlight, #5629):
//
//	"Creatures you control have vigilance. (Attacking doesn't cause them
//	 to tap.)"
//
// A layer 6 keyword grant to every creature its controller controls
// (CR 702.20a), the same grant as Samut's "Creatures you control have
// haste."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49cdd05c-feeb-4c24-9d88-069f5d2e08c3",
		Name:         "Serra's Blessing",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{YoursOnly: true}, "vigilance"),
		},
	})
}
