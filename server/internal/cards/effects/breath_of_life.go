package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Breath of Life — Sorcery {3}{W} (Portal, #5630):
//
//	"Return target creature card from your graveyard to the battlefield."
//
// Zombify's text, word for word.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30d9e200-b944-43ff-89b8-a550a788ae03",
		Name:         "Breath of Life",
		Completeness: CompletenessFull,
		Targets:      targetCreatureInYourGraveyard(),
		Purpose:      ForTargets(Reanimates(0)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			_, _ = reanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}
