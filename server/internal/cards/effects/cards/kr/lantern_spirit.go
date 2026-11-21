package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lantern Spirit — Creature — Spirit {2}{U}, 2/1 (Innistrad, #5572):
//
//	"Flying
//	 {U}: Return this creature to its owner's hand."
//
// Printed flying and Ephemeron's self-bounce (returnThisPermanentToOwnersHand)
// for {U}: the creature goes to its OWNER's hand (CR 400.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cf590bb7-2ea4-446a-9ca2-1e0706c2c41e",
		Name:            "Lantern Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{U}"),
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}
