package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Willow-Wind — Creature — Elemental {4}{U}, 3/4 (The Hobbit Eternal, HOC):
//
//	"Flying
//	 When this creature enters, scry 2."
//
// Faerie Seer's text on a bigger body (#4454).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ee3fd92c-48e0-4651-aeda-7eb2c75acbb0",
		Name:            "Willow-Wind",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Willow-Wind — scry 2", Do(Scry{N: 2})),
		},
	})
}
