package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Satyr Nyx-Smith — Creature — Satyr Shaman {2}{R}, 2/1 (Born of the Gods,
// #5559):
//
//	"Haste
//	 Inspired — Whenever this creature becomes untapped, you may pay
//	 {2}{R}. If you do, create a 3/1 red Elemental enchantment creature
//	 token with haste."
//
// Inspired is an ability word: the trigger is Key to the City's
// (EventUntapCard on itself), so it fires in the untap step and on
// any other untap.
// The payment is Aphelia's MayPay, its paid branch registered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3ad9d55d-02f7-412e-8756-6a026f8e68d7",
		Name:            "Satyr Nyx-Smith",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventUntapCard, effects.Self,
				"Satyr Nyx-Smith — you may pay {2}{R}; if you do, create a 3/1 red Elemental enchantment creature token with haste (inspired)",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}{R}",
						Question: "Satyr Nyx-Smith — pay {2}{R} to create a 3/1 red Elemental enchantment creature token with haste?",
						Then:     satyrNyxSmithPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// satyrNyxSmithPaid is the paid branch: one token. An on-disk key.
var satyrNyxSmithPaid = effects.PayUnlessBranches("satyr-nyx-smith/paid", func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
	return effects.CreateToken{Template: effects.TokenCard("3/1 red Elemental enchantment with haste"), N: 1}.Apply(ctx)
}, nil)
