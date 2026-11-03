package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imperious Inkmage — Creature — Orc Warlock {1}{W}{B}, 3/3 (Secrets of
// Strixhaven, #4456):
//
//	"Vigilance
//	 When this creature enters, surveil 2."
//
// Vigilance is a printed keyword; the enters trigger is Surveil (A.I.M.
// Synthoids' shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fcd31653-4964-46a6-820c-f17e8cbaa1fa",
		Name:            "Imperious Inkmage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Imperious Inkmage — surveil 2", Do(Surveil{N: 2})),
		},
	})
}
