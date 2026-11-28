package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defensive Formation — Enchantment {W}:
//
//	"Rather than the attacking player, you assign the combat damage of each
//	 creature attacking you. You can divide that creature's combat damage as you
//	 choose among any of the creatures blocking it."
//
// The engine's CR 510.1c division, handed to the defending player (#6192, ADR
// 0224 amendment 2026-11-25) for every creature attacking its controller that is
// blocked by two or more creatures.
func init() {
	effects.Register(effects.Spec{
		OracleID:                 "9386f4d3-aa72-4849-8921-83740762dc0e",
		Name:                     "Defensive Formation",
		Completeness:             effects.CompletenessFull,
		YouAssignAttackersDamage: true,
	})
}
