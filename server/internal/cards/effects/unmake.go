package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unmake — Instant {W/B}{W/B}{W/B} (Eventide, #5590):
//
//	"Exile target creature."
//
// Wander Off's exile at instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c551ba63-2a74-439d-8996-d566e2efbd5f",
		Name:         "Unmake",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return exileTheFirstTargetIfLegal(ctx)
		},
	})
}
