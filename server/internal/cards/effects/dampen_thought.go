package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dampen Thought — Instant — Arcane {1}{U}:
//
//	"Target player mills four cards.
//	Splice onto Arcane {1}{U}"
func init() {
	Register(Spec{
		OracleID:     "fa0ef984-b4e0-4c48-978c-0648e49d53f5",
		Name:         "Dampen Thought",
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{1}{U}"),
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return TargetPlayerMills(4)(ctx.Game, item)
		},
	})
}
