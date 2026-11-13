package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foul Imp — Creature — Imp {B}{B}, 2/2 (Stronghold, #5627):
//
//	"Flying
//	 When this creature enters, you lose 2 life."
//
// Flying is a printed keyword. The enters trigger is Nightdrinker
// Moroii's life loss at 2 (CR 603.6a), on the stack and paid by the Imp's
// controller however it entered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "25c04be3-f2e0-4f45-a1ca-9e75ffffa9b5",
		Name:            "Foul Imp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Foul Imp — you lose 2 life", Do(GainLife{Amount: -2})),
		},
	})
}
