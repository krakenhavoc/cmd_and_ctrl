package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aspiring Aeronaut — Creature — Human Artificer {3}{U}, 1/2 (Magic Origins, #5549):
//
//	"Flying (This creature can't be blocked except by creatures with flying or reach.)
//	 When this creature enters, create a 1/1 colorless Thopter artifact creature token with flying."
//
// Flying, and Nimble Thopterist's entry token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b2e31c0f-8914-4a18-bc05-44efb7e0f18c",
		Name:            "Aspiring Aeronaut",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Aspiring Aeronaut — create a 1/1 colorless Thopter artifact creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
	})
}
