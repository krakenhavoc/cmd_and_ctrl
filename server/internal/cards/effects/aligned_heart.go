package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aligned Heart — Enchantment {2}{W} (Tarkir: Dragonstorm Commander):
//
//	"Flurry — Whenever you cast your second spell each turn, put a rally
//	 counter on this enchantment. Then create a 1/1 white Monk creature token
//	 with prowess for each rally counter on it. (Whenever you cast a
//	 noncreature spell, the token gets +1/+1 until end of turn.)"
//
// The trigger is Cori-Steel Cutter's flurry. The counter goes on first, and the
// Monk count is read from the enchantment as the tokens are made, so the first
// flurry makes one Monk, the second two, and so on; each step settles before the
// next (a counter replacement's prompt is waited out, CR 608.2c). An Aligned Heart
// that left the battlefield puts no counter and makes no Monk: the count of "rally
// counters on it" is 0 (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ae61b2ef-2e60-499a-8708-e37b9bd5620e",
		Name:         "Aligned Heart",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Aligned Heart — flurry: put a rally counter on it, then create a 1/1 white Monk with prowess for each rally counter on it",
				Do(
					CounterOnThis{Kind: "rally", N: 1},
					Instruction(func(ctx *Context) error {
						if !onBattlefield(ctx.Game, ctx.Source()) {
							return nil
						}
						n := counterCountOn(ctx.Game, ctx.Source(), "rally")
						if n <= 0 {
							return nil
						}
						return CreateToken{Template: TokenCard("1/1 white Monk with prowess"), N: n}.Apply(ctx)
					}),
				)),
		},
	})
}
