package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rune-Sealed Wall — Artifact Creature — Wall {2}{U}, 0/4 (Foundations):
//
//	"Defender
//	 {T}: Surveil 1."
//
// Defender is a printed keyword; the surveil is the Surveil primitive
// behind a tap activation.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "74d31e99-ab48-483a-a62e-5b424b27d017",
		Name:            "Rune-Sealed Wall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true},
			Effect:  effects.Do(effects.Surveil{N: 1}),
		}},
	})
}
