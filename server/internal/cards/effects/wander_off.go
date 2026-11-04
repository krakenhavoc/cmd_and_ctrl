package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wander Off — Instant {3}{B} (Secrets of Strixhaven, #4456):
//
//	"Exile target creature."
//
// Plain exile removal at instant speed. A target that is gone or illegal
// when the spell resolves makes it fizzle (CR 608.2b); a commander's owner
// may send it to the command zone instead of exile (CR 903.9a), which the
// shared exile move handles.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b14d3226-adc2-4bdd-9e69-fefdc946ac04",
		Name:         "Wander Off",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return exileTheFirstTargetIfLegal(ctx)
		},
	})
}
