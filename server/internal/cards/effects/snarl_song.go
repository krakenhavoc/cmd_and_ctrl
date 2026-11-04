package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snarl Song — Sorcery {5}{G} (Secrets of Strixhaven, #4456):
//
//	"Converge — Create two 0/0 green and blue Fractal creature tokens. Put
//	 X +1/+1 counters on each of them and you gain X life, where X is the
//	 number of colors of mana spent to cast this spell."
//
// Converge is an ability word (CR 207.2c), so the counting is the card's:
// WantsDistinctColors pays the generic part of {5}{G} with colours not yet
// spent, and ReadsOwnPayment makes a waived cast ask what was paid
// (Painful Truths). X is read from the stack item as the spell resolves.
// The Fractals are created first and the counters PUT on each as printed
// (socB35FractalsWithCountersSteps); the life gain is the last step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "3b6fb13b-f126-4697-90a1-64b14ffcecd6",
		Name:                "Snarl Song",
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		Completeness:        CompletenessFull,
		OnResolve: ResolveSteps(append(
			socB35FractalsWithCountersSteps("snarl-song/fractals", 2, func(ctx *Context) int { return ctx.ColorsSpentCount() }),
			Instruction(func(ctx *Context) error {
				return GainLife{Amount: ctx.ColorsSpentCount()}.Apply(ctx)
			}))...),
	})
}
