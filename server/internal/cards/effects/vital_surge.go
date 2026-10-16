package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vital Surge — Instant — Arcane {1}{G}:
//
//	"You gain 3 life.
//	Splice onto Arcane {1}{G}"
func init() {
	Register(Spec{
		OracleID:     "e19c0205-67e3-4e44-995b-751e18fdaafd",
		Name:         "Vital Surge",
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{1}{G}"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
		},
	})
}
