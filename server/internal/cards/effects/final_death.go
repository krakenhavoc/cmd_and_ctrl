package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Final Death — Instant {4}{B} (Theros Beyond Death, #5514):
//
//	"Exile target creature."
//
// Wander Off's text: the target is exiled if it is still legal as the
// spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1be3eea4-eaff-469c-9689-abadac1e848c",
		Name:         "Final Death",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return exileTheFirstTargetIfLegal(ctx)
		},
	})
}
