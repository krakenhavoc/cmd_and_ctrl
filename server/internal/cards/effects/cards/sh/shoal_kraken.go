package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shoal Kraken — Creature — Kraken {4}{U}, 3/5:
//
//	"Constellation — Whenever an enchantment you control enters, you may draw a
//	 card. If you do, discard a card."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be700da5-123e-461e-8192-88f9d86925b9",
		Name:         "Shoal Kraken",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.On(game.EventETB, effects.GapB103EnchantmentEnteredUnderYourControl,
					"Shoal Kraken — you may draw a card; if you do, discard a card (constellation)",
					func(g *game.Game, item *game.StackItem) error { return effects.B16DrawThenDiscard(g, item, 1) }),
				"Shoal Kraken — draw a card, then discard a card?"),
		},
	})
}
