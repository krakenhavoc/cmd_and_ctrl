package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mind's Eye — Artifact {5} (EDHREC rank 2294):
//
//	"Whenever an opponent draws a card, you may pay {1}. If you do,
//	 draw a card."
//
// The colourless Rhystic Study. The trigger is the "whenever an
// opponent draws" shape (the drawer is the event's Actor; one event
// per card, so an opponent's draw-three asks three times, as
// printed), and the payoff is a MayPay prompt for {1} with the draw
// riding the "yes" — Leaf-Crowned Visionary's shape at a generic
// cost. A "pay" the controller cannot fund degrades to a decline.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "63fd2a57-7a47-4e07-947c-f4e9da7ee538",
		Name:         "Mind's Eye",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverAnOpponentDraws("Mind's Eye — pay {1} to draw a card", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}",
					Question: "Mind's Eye — pay {1} to draw a card?",
					OnPay: func(ctx *Context) error {
						return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
					},
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
