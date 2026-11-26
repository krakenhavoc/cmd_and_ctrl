package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reknit — Instant {1}{G/W} (Shadowmoor, S82 older sets):
//
//	"Regenerate target permanent."
//
// Welding Jar's and Asceticism's regeneration body (RegenerateTheTargetPermanent)
// as a spell. The shield lasts until end of turn; a target that has left the
// battlefield as the spell resolves does nothing (CR 701.19b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "963e36de-b803-417f-8166-4d328d358203",
		Name:         "Reknit",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.RegenerateTheTargetPermanent(ctx.Game, item)
		},
	})
}
