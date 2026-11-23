package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quicksilver Fisher — Creature — Phyrexian Drake {3}{U}{U}, 4/3:
//
//	"Flying
//	 When this creature enters, draw a card, then discard a card."
//
// Bellowing Crier's loot: the discard prompt opens after the draw, so the
// drawn card is a legal discard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b3e3d037-992b-4718-913d-7ab509d2047e",
		Name:            "Quicksilver Fisher",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Draws: 1, Discards: 1},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Quicksilver Fisher — draw a card, then discard a card", effects.B35LootOne),
		},
	})
}
