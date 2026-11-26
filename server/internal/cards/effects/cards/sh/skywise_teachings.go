package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skywise Teachings — Enchantment {3}{U} (Dragons of Tarkir):
//
//	"Whenever you cast a noncreature spell, you may pay {1}{U}. If you do, create
//	 a 2/2 blue Djinn Monk creature token with flying."
//
// A may-pay trigger (the question is asked as the trigger resolves, CR 603.5):
// the token is made only when the mana is really paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95a5e444-39f6-4f42-a83f-befa76350bc1",
		Name:         "Skywise Teachings",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Skywise Teachings — you may pay {1}{U}; if you do, create a 2/2 blue Djinn Monk with flying",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{U}",
						Question: "Skywise Teachings — pay {1}{U} to create a 2/2 blue Djinn Monk creature token with flying?",
						Then:     skywiseTeachingsPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var skywiseTeachingsPaid = effects.PayUnlessBranches("pay-unless/skywise-teachings-djinn-monk",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateToken{Template: effects.TokenCard("2/2 blue Djinn Monk with flying"), N: 1}.Apply(ctx)
	}, nil)
