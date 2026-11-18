package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jewel Thief (Streets of New Capenna, #5491):
//
//	Vigilance, trample
//	When this creature enters, create a Treasure token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "470f0bbf-a0c9-4216-96f0-d400dac17aef",
		Name:            "Jewel Thief",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Jewel Thief — create a Treasure token", Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
