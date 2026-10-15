package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eater of Days — Artifact Creature — Leviathan {4}, 9/8:
//
//	"Flying, trample
//	 When this creature enters, you skip your next two turns."
//
// The first card to skip more than one turn (CR 614.10a, ADR 0188
// Decision 1, #3565): N is two, and each is a separate debt the seam
// spends on one turn. The trigger is an ordinary one on the stack, so it
// can be answered, and the skip is written even if Eater left in
// response (the trigger does not depend on it staying).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "205f6384-7203-42ed-bc0f-8803cf1f4514",
		Name:            "Eater of Days",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Purpose:         game.Purpose{SkipsTurns: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Eater of Days — you skip your next two turns", Do(SkipYourNextTurns{N: 2})),
		},
	})
}
