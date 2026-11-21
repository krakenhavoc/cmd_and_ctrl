package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valkyrior Skyrider — Creature — God Warrior Hero {4}{W}, 3/4 (Marvel Super
// Heroes Commander, #4455):
//
//	"Flying
//	 When this creature enters, you gain 4 life."
//
// Airdrop Aeronauts' enter-and-gain body without the revolt condition.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ea0b0c71-4a1b-49aa-b4cd-db6c5d26b95e",
		Name:            "Valkyrior Skyrider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Valkyrior Skyrider — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
