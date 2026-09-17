package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exsanguinate — Sorcery for {X}{B}{B}:
//
//	"Each opponent loses X life. You gain life equal to the life
//	lost this way."
//
// S20 sub-PR 3: an untargeted X spell. "Life lost this way" is the sum
// actually lost, which is Gray Merchant's and Kokusho's sentence at a
// different N — so it is their shared body
// (b21DrainEachOpponentAndGainTheTotal), not a fourth copy of the loop.
//
// #793 moved that body onto the life continuation. This card used to
// read each opponent's life total back on the line after changing it,
// which is right until the loss pauses on a CR 616 ordering prompt: the
// read-back then happened before the prompt was answered, saw a life
// total that had not moved, and gained nothing.
func init() {
	Register(Spec{
		OracleID:     "8164b1e8-3350-465e-8a17-75f57d326344",
		Name:         "Exsanguinate",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b21DrainEachOpponentAndGainTheTotal(ctx.Game, item, ctx.X())
		},
	})
}
