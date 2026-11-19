package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Experimental Aviator — Creature — Human Artificer {3}{U}{U}, 0/3
// (Foundations Jumpstart):
//
//	"Flying
//	 When this creature enters, create two 1/1 colorless Thopter artifact
//	 creature tokens with flying."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "47f221a4-b51a-46b9-a4fc-46bc8ec7ddd3",
		Name:            "Experimental Aviator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Experimental Aviator — create two Thopters",
				Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 2})),
		},
	})
}
