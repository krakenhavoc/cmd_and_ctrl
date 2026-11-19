package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Manamorphose — Instant {1}{R/G} (Special Guests):
//
//	"Add two mana in any combination of colors.
//	 Draw a card."
//
// A spell, not a mana ability (CR 605.1): it uses the stack. The two
// mana are each of a colour its controller names (AnyCombinationOfColors).
func init() {
	Register(Spec{
		OracleID:     "89c83a6d-f6c8-4984-a888-0db62dfb93b1",
		Name:         "Manamorphose",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (AddMana{Produced: AnyCombinationOfColors(2)}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
