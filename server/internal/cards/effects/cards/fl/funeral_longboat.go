package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Funeral Longboat — Artifact — Vehicle {2}, 3/3 (Kaldheim, #5505):
//
//	"Vigilance
//	 Crew 1"
//
// Brute Suit's text exactly: vigilance rides PrintedKeywords and crew is
// the shared CrewCost / CrewEffect row (CR 702.122).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a994de7b-0128-4f03-922f-c8cd80cb5852",
		Name:            "Funeral Longboat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Funeral Longboat"),
		}},
	})
}
