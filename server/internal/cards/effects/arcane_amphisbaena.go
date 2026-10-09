package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arcane Amphisbaena — Creature — Snake {1}{G}, 1/1 (Reality
// Fracture):
//
//	"Deathtouch
//	 When this creature enters, empower Jace 2."
//
// ADR 0139: an enters trigger on the keyword action.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5c151e3b-fcf2-4adb-a60b-9eff0c360705",
		Name:            "Arcane Amphisbaena",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Arcane Amphisbaena — empower Jace 2", Do(EmpowerJace{N: 2})),
		},
	})
}
