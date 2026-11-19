package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Celestial Purge — Instant {1}{W} (Conflux, #5587):
//
//	"Exile target black or red permanent."
//
// The colours are words of the spell's text, read through OfColor (CR
// 612.2), and judged again at resolution (CR 608.2b), so a permanent
// that stopped being black or red in response stays.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ec1f6188-2516-46ac-8a03-7b7285b23a62",
		Name:         "Celestial Purge",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target black or red permanent", Or(OfColor("B"), OfColor("R"))),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ExileFirstTarget(ctx.Game, item)
		},
	})
}
