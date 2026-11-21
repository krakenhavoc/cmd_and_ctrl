package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gleaming Bastion — Land:
//
//	"{T}: Add {C}.
//	 {T}: Add {W} or {U}. Activate only if this land entered this turn or
//	 if you control a basic land."
//
// The twin of Dark Fortress (mshB3ConditionalDualLand).
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7785ffd4-f169-475d-9558-ce4877b3378a",
		Name:         "Gleaming Bastion",
		Completeness: effects.CompletenessFull,
		ManaAbilities: effects.MshB3ConditionalDualLand("{W|U}",
			"Add {W} or {U}. Activate only if this land entered this turn or if you control a basic land."),
	})
}
