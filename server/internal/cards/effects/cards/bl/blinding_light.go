package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blinding Light — Sorcery {2}{W} (Invasion, S82 older sets batch 662, #5616):
//
//	"Tap all nonwhite creatures."
//
// Every creature on the battlefield, whoever controls it, whose color after
// the layers does not include white (a colorless creature is nonwhite).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b315dc3-c330-4b30-b6ad-4da12ccf6ca3",
		Name:         "Blinding Light",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.OldB662TapAllNonwhiteCreatures(ctx.Game, item)
		},
	})
}
