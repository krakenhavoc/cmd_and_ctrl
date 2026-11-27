package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twiddle — Instant {U} (Eighth Edition, S82 older sets batch 618, #5607):
//
//	"You may tap or untap target artifact, creature, or land."
//
// Rimewind Taskmage's tap-or-untap on a narrower target. The choice is made as
// it resolves, and either answer is always legal, so the player can leave the
// permanent as it is (the "may").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "773ad2ef-5acc-49ea-8d85-056330e87039",
		Name:         "Twiddle",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target artifact, creature, or land",
			effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B38TapOrUntapTarget("Twiddle")(ctx.Game, item)
		},
	})
}
