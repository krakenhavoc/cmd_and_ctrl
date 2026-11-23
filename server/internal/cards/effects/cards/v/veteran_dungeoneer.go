package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Veteran Dungeoneer — Creature — Human Warrior {3}{W}, 3/4:
//
//	"When this creature enters, venture into the dungeon."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ce7b966-2a10-4250-8c0d-525e73efe2f4",
		Name:         "Veteran Dungeoneer",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.WhenThisEntersVenture("Veteran Dungeoneer")},
	})
}
