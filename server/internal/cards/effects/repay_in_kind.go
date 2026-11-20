package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "24b8177f-5697-4465-9305-cb84307972ae",
		Name:         "Repay in Kind",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return darkholdLowestLifeBecomes(ctx.Game, item)
		},
	})
}
