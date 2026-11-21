package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brute Suit — Artifact — Vehicle {3}, 4/3 (Neon Dynasty):
//
//	"Vigilance
//	 Crew 1"
//
// Vigilance is the printed keyword. The Vehicle is not a creature until a crew ability resolves
// (CR 702.122), so any printed keyword sits inert on the artifact until then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bc64e396-7fa1-4a4e-a528-352737714a9c",
		Name:            "Brute Suit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Brute Suit"),
		}},
	})
}
