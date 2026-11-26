package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyrrhic Revival — {3}{W/B}{W/B}{W/B} Sorcery:
//
//	"Each player returns each creature card from their graveyard to the
//	 battlefield with an additional -1/-1 counter on it."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e8daf155-1fc9-4a0e-a7b5-69a50bc23576",
		Name:         "Pyrrhic Revival",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB522EachPlayerReturnsEachCreatureWithMinusCounter(ctx.Game, item)
		},
	})
}
