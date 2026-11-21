package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Manamorphose — Instant {1}{R/G} (Special Guests):
//
//	"Add two mana in any combination of colors.
//	 Draw a card."
//
// A spell, not a mana ability (CR 605.1): it uses the stack. The two
// mana are each of a colour its controller names (AnyCombinationOfColors).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "89c83a6d-f6c8-4984-a888-0db62dfb93b1",
		Name:         "Manamorphose",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.AddMana{Produced: effects.AnyCombinationOfColors(2)}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
