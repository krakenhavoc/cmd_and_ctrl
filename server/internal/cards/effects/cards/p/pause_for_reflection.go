package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pause for Reflection — Instant {2}{G} (Guilds of Ravnica, #5525):
//
//	"Convoke
//	 Prevent all combat damage that would be dealt this turn."
//
// Convoke is the shared tap cost (CR 702.51a); the prevention is
// Darkness's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7f266b5-7258-4eaa-b447-66460ec38505",
		Name:         "Pause for Reflection",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		OnResolve:    effects.ResolveSteps(effects.PreventAllCombatDamageThisTurn{Label: "Pause for Reflection: prevent combat damage"}),
	})
}
