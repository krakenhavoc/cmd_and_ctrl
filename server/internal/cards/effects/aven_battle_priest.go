package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aven Battle Priest — Creature — Bird Cleric {5}{W}, 3/3 (Magic Origins, #5549):
//
//	"Flying (This creature can't be blocked except by creatures with flying or reach.)
//	 When this creature enters, you gain 3 life."
//
// Angel of Mercy's shape: flying and an entry life gain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "696fb1fa-5535-4f73-8235-ad4416dcfdd0",
		Name:            "Aven Battle Priest",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Aven Battle Priest — you gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
}
