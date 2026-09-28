package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valor Made Real — Instant {W}:
//
//	"Target creature can block any number of creatures this turn."
//
// BlockCapacityUntilEOT with AnyNumber (#1715).
func init() {
	Register(Spec{
		OracleID:     "5bd1348c-51cc-4426-b995-8bb6d4a5bd2e",
		Name:         "Valor Made Real",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			return BlockCapacityUntilEOT{
				Target:    ts[0].ID,
				AnyNumber: true,
				Label:     "Valor Made Real — can block any number of creatures",
			}.Apply(ctx)
		},
	})
}
