package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Selenia, Dark Angel — {3}{W}{B} Legendary Creature — Phyrexian Angel, 3/3
// (Tempest, S82 OLD batch 769, #5628):
//
//	"Flying
//	 Pay 2 life: Return Selenia to its owner's hand."
//
// Flying is a printed keyword. The life is a cost, paid on activation (CR
// 118.4), so it can be answered; the return is to its OWNER's hand and does
// nothing to a Selenia that has already left (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3db1ac09-e35a-4f50-97c4-f65dbc28daef",
		Name:            "Selenia, Dark Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay 2 life: Return Selenia to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    game.AbilityCost{Life: 2},
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}
