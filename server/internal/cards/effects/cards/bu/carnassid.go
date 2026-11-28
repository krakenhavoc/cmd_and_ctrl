package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carnassid — Creature — Beast {4}{G}{G}, 5/4 (Stronghold, S82 OLD batch 746,
// #5627):
//
//	"Trample
//	 {1}{G}: Regenerate this creature."
//
// Trample is a printed keyword; the ability is the shared regenerate-this body
// (CR 701.19a) behind a mana cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "97b386da-4bc0-4c39-a5cc-17e441c45127",
		Name:            "Carnassid",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}{G}"),
			Effect:  effects.RegenerateThisSource,
		}},
	})
}
