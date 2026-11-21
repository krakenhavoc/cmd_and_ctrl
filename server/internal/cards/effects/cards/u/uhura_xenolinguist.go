package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uhura, Xenolinguist — {4}{U} Legendary Creature — Human Officer 3/4
// (Star Trek Commander, #4974):
//
//	"Whenever a non-Human creature enters, you may pay {1}. If you do,
//	 draw a card."
//
// Any player's creature, token or card (b33AnyCreatureEntered), that is
// not a Human as it enters, through the layered view (a changeling is a
// Human). "You may pay {1}. If you do" is Mentor of the Meek's MayPay:
// a prompt for Uhura's controller that draws on "Pay" and does nothing on
// "Don't pay".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5d5c4a17-fa04-4889-819e-408fc2ff112e",
		Name:         "Uhura, Xenolinguist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.TrekB46NonHumanCreatureEntered, "Uhura, Xenolinguist — pay {1} to draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Uhura, Xenolinguist — pay {1} to draw a card?",
						Then:     effects.MayPayDrawACard,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
