package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cancel — Instant {1}{U}{U} (EDHREC rank 2058):
//
//	"Counter target spell."
//
// The three-mana Counterspell, and the reason every set has one.
// Counterspell's shape with a different cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d00fb28-ea6c-49a9-b4af-ffb38860a9a7",
		Name:         "Cancel",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
