package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ceta Sanctuary — Enchantment {2}{U} (Apocalypse, S82 OLD, #5614):
//
//	"At the beginning of your upkeep, if you control a red or green
//	 permanent, draw a card, then discard a card. If you control a red
//	 permanent and a green permanent, instead draw two cards, then discard a
//	 card."
//
// One trigger with an intervening if (CR 603.4); the "instead" is read as it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45c4e67c-e452-41c5-8baa-050819a1321f",
		Name:         "Ceta Sanctuary",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Discards: 1},
		Triggered: []game.TriggeredAbility{
			effects.OldB629UpkeepIfYouControlEither("R", "G", "you control a red or green permanent",
				"Ceta Sanctuary — draw a card, then discard a card (two cards if you control a red permanent and a green permanent)",
				func(g *game.Game, item *game.StackItem, both bool) error {
					if both {
						return effects.B16DrawThenDiscard(g, item, 2)
					}
					return effects.B16DrawThenDiscard(g, item, 1)
				}),
		},
	})
}
