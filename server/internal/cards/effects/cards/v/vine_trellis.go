package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vine Trellis — Creature — Plant Wall {1}{G} 0/4 (Eighth Edition, S82 older sets batch 620, #5607):
//
//	"Defender
//	 {T}: Add {G}."
//
// Itlimoc's {T}: Add {G} with defender.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "57de8fe7-3d1b-41cd-8354-38aff3d2d052",
		Name:            "Vine Trellis",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "{T}: Add {G}",
			Answers:  game.AnswerValue,
		}},
	})
}
