package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rune-Sealed Wall — Artifact Creature — Wall {2}{U}, 0/4 (Foundations):
//
//	"Defender
//	 {T}: Surveil 1."
//
// Defender is a printed keyword; the surveil is the Surveil primitive
// behind a tap activation.
func init() {
	Register(Spec{
		OracleID:        "74d31e99-ab48-483a-a62e-5b424b27d017",
		Name:            "Rune-Sealed Wall",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true},
			Effect:  Do(Surveil{N: 1}),
		}},
	})
}
