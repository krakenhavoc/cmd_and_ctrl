package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// God-Favored General — Creature — Human Soldier {1}{W}, 1/1 (Born of the Gods,
// #5559):
//
//	"Inspired — Whenever this creature becomes untapped, you may pay
//	 {2}{W}. If you do, create two 1/1 white Soldier enchantment creature
//	 tokens."
//
// Inspired is an ability word: the trigger is Key to the City's
// (EventUntapCard on itself), so it fires in the untap step and on
// any other untap.
// The payment is Aphelia's MayPay, its paid branch registered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e861ddd5-64f7-4c59-ab96-0c69ef74febe",
		Name:         "God-Favored General",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventUntapCard, Self,
				"God-Favored General — you may pay {2}{W}; if you do, create two 1/1 white Soldier enchantment creature tokens (inspired)",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}{W}",
						Question: "God-Favored General — pay {2}{W} to create two 1/1 white Soldier enchantment creature tokens?",
						Then:     godFavoredGeneralPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// godFavoredGeneralPaid is the paid branch: two tokens. An on-disk key.
var godFavoredGeneralPaid = PayUnlessBranches("god-favored-general/paid", func(ctx *Context, _ game.PayUnlessAnswer) error {
	return CreateToken{Template: TokenCard("1/1 white Soldier enchantment"), N: 2}.Apply(ctx)
}, nil)
