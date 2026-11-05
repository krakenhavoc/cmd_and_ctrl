package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glider Kids — Creature — Human Pilot Ally {2}{W}, 2/3:
//
//	"Flying
//	 When this creature enters, scry 1."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ff5ce28b-25be-4136-a7b2-8f25c08a6455",
		Name:            "Glider Kids",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Scry: 1},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Glider Kids — scry 1", Do(Scry{N: 1})),
		},
	})
}
