package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urza's Guilt — Sorcery {2}{U}{B} (Planeshift, S82 OLD #5615):
//
//	"Each player draws two cards, then discards three cards, then loses 4
//	 life."
//
// Every player draws two, then chooses their own three discards, then each
// loses 4 life (a loss, not damage), in that order (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0aeeeb6-490c-4638-a4ad-498e4d85413e",
		Name:         "Urza's Guilt",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue, Draws: 2, Discards: 3},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.OldB637UrzasGuilt(ctx.Game, item)
		},
	})
}
