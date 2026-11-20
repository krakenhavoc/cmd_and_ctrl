package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fate Forgotten — Instant {2}{W} (Dragons of Tarkir,
// #5551):
//
//	"Exile target artifact or enchantment."
//
// Altar's Light's body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7f591bfb-a241-4980-9cb3-fb5af3355f6b",
		Name:         "Fate Forgotten",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ExileFirstTarget(ctx.Game, item)
		},
	})
}
