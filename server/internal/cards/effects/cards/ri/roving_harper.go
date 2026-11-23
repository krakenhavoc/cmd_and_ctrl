package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roving Harper — Creature — Elf Scout {2}{W}, 2/2 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"When this creature enters, draw a card."
//
// Baleful Strix's cantrip trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6474f6b3-4e8e-4056-a01c-2f5e5e59c8de",
		Name:         "Roving Harper",
		Purpose:      game.Purpose{Draws: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Roving Harper — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
