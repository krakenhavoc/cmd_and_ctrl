package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inheritance — Enchantment {W} (Alliances, #5634):
//
//	"Whenever a creature dies, you may pay {3}. If you do, draw a card."
//
// Any creature's death, yours or not (CR 700.4, read as it last existed), asks
// the controller for {3}; paying draws a card (a payment they cannot fund is
// a decline).
//
// No simplification.
func init() {
	const name = "Inheritance"
	effects.Register(effects.Spec{
		OracleID:     "742cb752-dd20-4b1f-8d96-7e9a8f864e75",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.DiedCreature(ev, g)
				return ok
			}, name+" — you may pay {3}; if you do, draw a card", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{3}",
					Question: name + " — pay {3} to draw a card?",
					Then:     effects.MayPayDrawACard,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
