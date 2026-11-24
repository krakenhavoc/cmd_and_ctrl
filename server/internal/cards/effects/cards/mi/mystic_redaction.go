package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mystic Redaction — Enchantment {2}{U} (Modern Horizons 2, #5501):
//
//	"At the beginning of your upkeep, scry 1.
//	 Whenever you discard a card, each opponent mills two cards."
//
// One discard is one trigger, whatever dropped it (cost or effect).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2e54e13-dad5-4d01-850f-a1735eea54ae",
		Name:         "Mystic Redaction",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Mystic Redaction — scry 1", effects.Do(effects.Scry{N: 1})),
			effects.On(game.EventDiscardCard, effects.YouDiscardedA(func(game.Card) bool { return true }),
				"Mystic Redaction — each opponent mills two cards",
				func(g *game.Game, item *game.StackItem) error { return effects.B12EachOpponentMills(g, item, 2) }),
		},
	})
}
