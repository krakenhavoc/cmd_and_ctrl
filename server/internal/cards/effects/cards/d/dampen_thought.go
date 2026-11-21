package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dampen Thought — Instant — Arcane {1}{U}:
//
//	"Target player mills four cards.
//	Splice onto Arcane {1}{U}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa0ef984-b4e0-4c48-978c-0648e49d53f5",
		Name:         "Dampen Thought",
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoArcane("{1}{U}"),
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TargetPlayerMills(4)(ctx.Game, item)
		},
	})
}
