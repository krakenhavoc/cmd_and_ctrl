package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rabid Bloodsucker — Creature — Vampire {4}{B}, 3/2 (Magic Origins, #5549):
//
//	"Flying (This creature can't be blocked except by creatures with flying or reach.)
//	 When this creature enters, each player loses 2 life."
//
// Flying, and Rathi Fiend's entry: every live player, its controller
// included, loses 2 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e6b0c47b-9758-4ee3-b831-0af0650660fe",
		Name:            "Rabid Bloodsucker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rabid Bloodsucker — each player loses 2 life", nemEachPlayerLoses(2)),
		},
	})
}
