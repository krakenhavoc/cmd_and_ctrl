package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overflowing Insight — Sorcery {4}{U}{U}{U} (Foundations Jumpstart):
//
//	"Target player draws seven cards."
//
// Ancestral Vision's shape at seven: the chosen player draws, not the
// caster. A target that is gone by resolution draws nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ff03a063-9634-4746-a208-9dbcde5c6bdb",
		Name:         "Overflowing Insight",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		Purpose:      ForTargets(game.TargetPurpose{Slot: 0, Draws: 7}),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			return DrawCards{Player: item.Targets[0].ID, N: 7}.Apply(ctx)
		},
	})
}
