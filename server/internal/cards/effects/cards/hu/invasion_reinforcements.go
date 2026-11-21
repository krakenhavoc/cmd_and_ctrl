package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invasion Reinforcements — Creature — Human Warrior Ally {1}{W}, 1/1:
//
//	"Flash
//	 When this creature enters, create a 1/1 white Ally creature
//	 token."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "185316ff-6fb7-4b5c-8a60-3238674562a6",
		Name:            "Invasion Reinforcements",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Invasion Reinforcements — create a 1/1 Ally", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Ally"), N: 1})),
		},
	})
}
