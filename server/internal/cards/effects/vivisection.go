package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vivisection — Sorcery {3}{U}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Draw three cards."
//
// Altar's Reap's machinery at three cards: the sacrifice is an
// additional cost paid on cast (CR 601.2b, 118.8), the draw is the
// whole effect.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "46f78c42-85d1-4cd4-9fc3-d5fdec04a4c8",
		Name:           "Vivisection",
		Purpose:        game.Purpose{Draws: 3, Answers: game.AnswerSacOutlet},
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
	})
}
