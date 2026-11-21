package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pheres-Band Raiders — Creature — Centaur Warrior {5}{G}, 5/5 (Born of the Gods,
// #5559):
//
//	"Inspired — Whenever this creature becomes untapped, you may pay
//	 {2}{G}. If you do, create a 3/3 green Centaur enchantment creature
//	 token."
//
// Inspired is an ability word: the trigger is Key to the City's
// (EventUntapCard on itself), so it fires in the untap step and on
// any other untap.
// The payment is Aphelia's MayPay, its paid branch registered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fb54c18-f69b-4712-af36-14d8a6d4df2b",
		Name:         "Pheres-Band Raiders",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventUntapCard, effects.Self,
				"Pheres-Band Raiders — you may pay {2}{G}; if you do, create a 3/3 green Centaur enchantment creature token (inspired)",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}{G}",
						Question: "Pheres-Band Raiders — pay {2}{G} to create a 3/3 green Centaur enchantment creature token?",
						Then:     pheresBandRaidersPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// pheresBandRaidersPaid is the paid branch: one token. An on-disk key.
var pheresBandRaidersPaid = effects.PayUnlessBranches("pheres-band-raiders/paid", func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
	return effects.CreateToken{Template: effects.TokenCard("3/3 green Centaur enchantment"), N: 1}.Apply(ctx)
}, nil)
