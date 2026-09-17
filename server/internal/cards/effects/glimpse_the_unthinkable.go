package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glimpse the Unthinkable — "Target player mills ten cards." Fixed
// count (no "until you hit a land" surveillance clause to model).
// A target with fewer than ten cards mills the rest and stays in the
// game (CR 701.17b); they lose only when they next draw from the
// empty library (CR 704.5b).
func init() {
	Register(Spec{
		OracleID:     "552f0163-a19d-4671-888f-044fc0354875",
		Name:         "Glimpse the Unthinkable",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return MillCards{Player: item.Targets[0].ID, N: 10}.Apply(ctx)
		},
	})
}
