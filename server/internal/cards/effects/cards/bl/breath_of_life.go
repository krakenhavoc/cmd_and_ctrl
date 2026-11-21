package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Breath of Life — Sorcery {3}{W} (Portal, #5630):
//
//	"Return target creature card from your graveyard to the battlefield."
//
// Zombify's text, word for word.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30d9e200-b944-43ff-89b8-a550a788ae03",
		Name:         "Breath of Life",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreatureInYourGraveyard(),
		Purpose:      effects.ForTargets(effects.Reanimates(0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			_, _ = effects.ReanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}
