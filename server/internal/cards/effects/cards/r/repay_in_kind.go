package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Repay in Kind — Sorcery {5}{B}{B} (Rise of the Eldrazi, #5581):
//
//	"Each player's life total becomes the lowest life total among all
//	 players."
//
// The Darkhold's second ability as a spell: the lowest life total among the
// players still in the game is found once, then each of them gains or loses
// what it takes to end there (CR 119.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "24b8177f-5697-4465-9305-cb84307972ae",
		Name:         "Repay in Kind",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DarkholdLowestLifeBecomes(ctx.Game, item)
		},
	})
}
