package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sun-Collared Raptor — Creature — Dinosaur {1}{R}, 1/2 (Rivals of
// Ixalan, #5531):
//
//	"Trample
//	 {2}{R}: This creature gets +3/+0 until end of turn."
//
// Trample is printed data (CR 702.19b); the pump is Avarax's
// thisGetsUntilEndOfTurn at +3/+0, and it can be activated any number of
// times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9debaed0-b605-4778-9977-216d194ecdeb",
		Name:            "Sun-Collared Raptor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}: This creature gets +3/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}{R}"),
			Effect:  thisGetsUntilEndOfTurn(3, 0, "Sun-Collared Raptor — +3/+0 until end of turn"),
		}},
	})
}
