package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Promise of Power — Sorcery {2}{B}{B}{B}:
//
//	"Choose one —
//	 • You draw five cards and you lose 5 life.
//	 • Create an X/X black Demon creature token with flying, where X is
//	   the number of cards in your hand.
//	 Entwine {4} (Choose both if you pay the entwine cost.)"
//
// X is counted as the second bullet runs and fixed then: the token is a
// plain X/X, not a characteristic-defining ability. Entwined (CR
// 702.42, #3866) the bullets run in the order written (CR 702.42b), so
// the five cards are drawn before X is counted, and the Demon is that
// much bigger. A draw a replacement pauses holds the Demon until it is
// answered (#2789). With an empty hand the Demon is a 0/0 and dies to
// the state-based check, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "fde459b2-f45f-40d2-b328-58f65440f494",
		Name:          "Promise of Power",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Entwine("{4}")},
		Modes: Entwining(ChooseOne(
			// The life loss waits for a prompt a draw asks (a dredge offer; CR 608.2c, #4155).
			ModeSteps("You draw five cards and you lose 5 life.", nil,
				DrawCards{N: 5},
				Instruction(func(ctx *Context) error {
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -5)
				})),
			ModeDoing("Create an X/X black Demon creature token with flying, where X is the number of cards in your hand.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					x := 0
					if p := ctx.Game.PlayerByIDForEffect(item.Controller); p != nil {
						x = len(p.Hand.Cards)
					}
					return CreateToken{Controller: item.Controller, Template: promiseOfPowerDemon(x), N: 1}.Apply(ctx)
				}),
		)),
	})
}

// promiseOfPowerDemon is the printed X/X black Demon with flying.
func promiseOfPowerDemon(x int) game.Card {
	return game.Card{
		Name:      "Demon",
		TypeLine:  "Token Creature — Demon",
		Power:     x,
		Toughness: x,
		Colors:    []string{"B"},
		Keywords:  []string{"flying"},
	}
}
