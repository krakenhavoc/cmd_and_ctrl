package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lunar Mystic — {2}{U}{U} Creature — Human Wizard 2/2 (Commander 2020):
//
//	"Whenever you cast an instant spell, you may pay {1}. If you do, draw a
//	 card."
//
// Mentor of the Meek's "you may pay {1}" with its draw (MayPay, whose "yes"
// pays as the trigger resolves), off the cast of an instant spell.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9652f79c-3060-4e9c-9623-0ebe4ab4094c",
		Name:         "Lunar Mystic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Instant(), "Lunar Mystic — you may pay {1} to draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Lunar Mystic — pay {1} to draw a card?",
						Then:     effects.MayPayDrawACard,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
