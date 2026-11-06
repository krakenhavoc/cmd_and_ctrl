package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prideful Parent — Creature — Cat {2}{W}, 2/2 (Foundations, #4465):
//
//	"Vigilance
//	 When this creature enters, create a 1/1 white Cat creature token."
//
// Kyoshi Warriors' enter trigger with the Cat token; vigilance rides
// PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a90d1d35-6d67-43b3-8256-b0318b3a4b07",
		Name:            "Prideful Parent",
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Prideful Parent — create a 1/1 white Cat", Do(CreateToken{Template: b36WhiteCatToken(), N: 1})),
		},
	})
}
