package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valorous Steed — Creature — Unicorn {4}{W}, 3/3:
//
//	"Vigilance
//	 When this creature enters, create a 2/2 white Knight creature token
//	 with vigilance."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4b5f30a6-b0fb-4dcb-85d5-1db1bba73737",
		Name:            "Valorous Steed",
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Valorous Steed — create a 2/2 white Knight token with vigilance",
				Do(CreateToken{Template: TokenCard("2/2 white Knight with vigilance"), N: 1})),
		},
	})
}
