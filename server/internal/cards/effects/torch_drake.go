package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Torch Drake — Creature — Drake {3}{U}, 2/2 (Guildpact, #5599):
//
//	"Flying
//	 {1}{R}: This creature gets +1/+0 until end of turn."
//
// Firebreathing at {1}{R} on a flyer: Hematite Golem's self-pump
// (thisGetsUntilEndOfTurn), activated as often as it is paid for.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8408f99a-aa3c-40d0-9e26-dd527343e62b",
		Name:            "Torch Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{R}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Torch Drake — +1/+0 until end of turn"),
		}},
	})
}
