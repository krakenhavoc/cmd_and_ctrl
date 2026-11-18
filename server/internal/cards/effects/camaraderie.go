package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Camaraderie — Sorcery {4}{G}{W}:
//
//	"You gain X life and draw X cards, where X is the number of
//	 creatures you control. Creatures you control get +1/+1 until end
//	 of turn."
//
// X is counted once as the spell resolves and is the same for the life
// and the cards (CR 608.2c); the pump then reaches the creatures you
// control at that point (CR 611.2c), after any that the draw put into
// play, which is none. A layer 7c change that ends at cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9aab386c-d48c-4611-b757-aa69b26cc1b1",
		Name:         "Camaraderie",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsCreature() })
			if err := (GainLife{Amount: x}).Apply(ctx); err != nil {
				return err
			}
			if err := (DrawCards{N: x}).Apply(ctx); err != nil {
				return err
			}
			return creaturesYouControlGetUntilEOT(1, 1, "Camaraderie — +1/+1")(item, ctx)
		},
	})
}
