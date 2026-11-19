package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragon Hatchling — Creature — Dragon {1}{R}, 0/1:
//
//	"Flying
//	 {R}: This creature gets +1/+0 until end of turn."
func init() {
	Register(Spec{
		OracleID:        "3f48f6a7-906d-4287-8440-06bfabc93b89",
		Name:            "Dragon Hatchling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{R}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Dragon Hatchling — +1/+0 until end of turn"),
		}},
	})
}
