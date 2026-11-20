package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glare of Heresy — Sorcery {1}{W} (Theros, #5561):
//
//	"Exile target white permanent."
//
// Celestial Purge's body for a white permanent of any type.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6e127e6e-b292-4398-b3fa-8b520a7d26a3",
		Name:         "Glare of Heresy",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target white permanent", OfColor("W")),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ExileFirstTarget(ctx.Game, item)
		},
	})
}
