package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Treetop Freedom Fighters — Creature — Human Rebel Ally {2}{R}, 2/1
// (Avatar: The Last Airbender, #4459):
//
//	"Haste
//	 When this creature enters, create a 1/1 white Ally creature
//	 token."
//
// The token is the set's printed 1/1 white Ally (Invasion
// Reinforcements' row).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "922176b8-7174-4248-92fb-b196ceb802c7",
		Name:            "Treetop Freedom Fighters",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Treetop Freedom Fighters — create a 1/1 Ally", Do(CreateToken{Template: TokenCard("1/1 white Ally"), N: 1})),
		},
	})
}
