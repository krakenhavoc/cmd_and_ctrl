package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burning Curiosity — {2}{R} Sorcery:
//
//	"As an additional cost to cast this spell, you may blight 1. (You
//	 may put a -1/-1 counter on a creature you control.)
//	 Exile the top two cards of your library. If this spell's
//	 additional cost was paid, exile the top three cards instead. Until
//	 the end of your next turn, you may play those cards."
//
// #1703: OptionalBlight(1), paid at CR 601.2h onto a creature the
// caster names. The impulse exile is ExileTopNUntilYourNextTurn with a
// count read off ctx.BlightPaid() at resolution. No simplification.
func init() {
	Register(Spec{
		OracleID:      "7f420633-901d-47f9-ae0f-0f5b0ea8359c",
		Name:          "Burning Curiosity",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{OptionalBlight(1)},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := 2
			if ctx.BlightPaid() {
				n = 3
			}
			return ExileTopNUntilYourNextTurn(ctx, n)
		},
	})
}
