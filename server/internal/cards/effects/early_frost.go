package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Early Frost — Instant {1}{U} (Fifth Dawn, #5604):
//
//	"Tap up to three target lands."
//
// Choking Tethers' tap of each still-legal target, over up to three lands.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f9cdbc2-d2f2-4390-8506-350fabf3f89a",
		Name:         "Early Frost",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("up to three target lands", Land()).WithCount(0, 3),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return tapEachLegalTarget(ctx.Game, item)
		},
	})
}
