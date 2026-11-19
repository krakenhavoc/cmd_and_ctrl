package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hurska Sweet-Tooth — Legendary Creature — Bear {2}{G}, 3/3 (Foundations
// Jumpstart):
//
//	"Whenever Hurska attacks, create a Food token.
//	 Whenever you gain life, you may pay {G/W}. When you do, target creature
//	 gets +X/+X until end of turn, where X is the amount of life you gained."
//
// The attack trigger is a plain Food. The life trigger is MayPay then a
// reflexive trigger (CR 603.12, Go-Shintai's shape): the payment is made as
// the trigger resolves, the target is chosen as the reflexive trigger goes
// on the stack, and X is the life gained, fixed from the triggering event
// when the payment is made and carried as Params.Amount.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d49b6a1-4ae2-4443-a6b3-9ef806cb6fc4",
		Name:         "Hurska Sweet-Tooth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Hurska Sweet-Tooth — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
			WheneverYouGainLife("Hurska Sweet-Tooth — you may pay {G/W}; target creature gets +X/+X",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{G/W}",
						Question: "Hurska Sweet-Tooth — pay {G/W}?",
						Then:     s82BnsB51HurskaPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
