package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mentor of the Meek — Creature — Human Soldier {2}{W}, 2/2 (EDHREC
// rank 649):
//
//	"Whenever another creature you control with power 2 or less
//	 enters, you may pay {1}. If you do, draw a card."
//
// The token deck's card-draw engine. One trigger per small creature
// — "another", so the Mentor's own entry is excluded — and the "you
// may pay {1}. If you do" is the MayPay primitive: a prompt for the
// controller that draws on "Pay" (with the mana in pool or from
// untapped sources) and does nothing on "Don't pay". Power is read
// at trigger time (the entering creature's current power), which is
// when the printed condition is evaluated.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9f4f96b-6e54-4fe6-8df7-623e0fc72409",
		Name:         "Mentor of the Meek",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && c.CurrentPower() <= 2
			}, "Mentor of the Meek — pay {1} to draw a card", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}",
					Question: "Mentor of the Meek — pay {1} to draw a card?",
					OnPay: func(ctx *Context) error {
						return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
					},
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
