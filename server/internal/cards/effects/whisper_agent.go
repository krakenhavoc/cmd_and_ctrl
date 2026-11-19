package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whisper Agent — Creature — Human Rogue {1}{U/B}{U/B}, 3/2 (Ravnica: Clue Edition):
//
//	"Flash
//	 When this creature enters, surveil 1."
//
// A.I.M. Synthoids's ETB trigger: the surveil goes on the stack, so an
// opponent can respond to it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2ed76a99-fe5c-4d68-b572-a30848055c8a",
		Name:            "Whisper Agent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Whisper Agent — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
