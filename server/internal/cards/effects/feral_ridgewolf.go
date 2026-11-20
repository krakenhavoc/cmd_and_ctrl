package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feral Ridgewolf — Creature — Wolf {2}{R}, 1/2 (Innistrad, #5572):
//
//	"Trample
//	 {1}{R}: This creature gets +2/+0 until end of turn."
//
// Printed trample and Freejam Regent's repeatable {1}{R} pump.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3a799843-1304-4abf-b8d7-db616628ee3f",
		Name:            "Feral Ridgewolf",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{R}"),
			Effect:  thisGetsUntilEndOfTurn(2, 0, "Feral Ridgewolf — +2/+0"),
		}},
	})
}
