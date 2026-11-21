package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Searchlight Companion — Artifact Creature — Drone {3}, 1/1 (Kamigawa:
// Neon Dynasty, #5493):
//
//	"Flying
//	 When this creature enters, create a 1/1 colorless Spirit creature
//	 token."
//
// An enters trigger that makes the printed 1/1 colorless Spirit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "426569c7-d868-4af1-bc83-5b5200532d17",
		Name:            "Searchlight Companion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Searchlight Companion — create a 1/1 colorless Spirit creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Spirit"), N: 1})),
		},
	})
}
