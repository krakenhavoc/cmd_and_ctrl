package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion Reinforcements — Creature — Human Warrior Ally {1}{W}, 1/1:
//
//	"Flash
//	 When this creature enters, create a 1/1 white Ally creature
//	 token."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "185316ff-6fb7-4b5c-8a60-3238674562a6",
		Name:            "Invasion Reinforcements",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Invasion Reinforcements — create a 1/1 Ally", Do(CreateToken{Template: TokenCard("1/1 white Ally"), N: 1})),
		},
	})
}
