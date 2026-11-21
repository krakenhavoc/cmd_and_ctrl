package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyscanner — Artifact Creature — Thopter {3}, 1/1:
//
//	"Flying
//	 When this creature enters, draw a card."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "974f788a-039f-4310-a2fe-16b14a1e2d35",
		Name:            "Skyscanner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.SamiDrawOnETB("Skyscanner")},
	})
}
