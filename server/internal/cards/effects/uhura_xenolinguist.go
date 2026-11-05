package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5d5c4a17-fa04-4889-819e-408fc2ff112e",
		Name:         "Uhura, Xenolinguist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, trekB46NonHumanCreatureEntered, "Uhura, Xenolinguist — pay {1} to draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Uhura, Xenolinguist — pay {1} to draw a card?",
						Then:     mayPayDrawACard,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
