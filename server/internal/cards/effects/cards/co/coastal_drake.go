package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coastal Drake — Creature — Drake {2}{U}, 2/1 (Apocalypse, #5614):
//
//	"Flying
//	 {1}{U}, {T}: Return target Kavu to its owner's hand."
//
// The target is any Kavu on the battlefield, the Drake's controller's own
// included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d4220231-e9ac-4390-8c7b-979549b26d33",
		Name:            "Coastal Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {T}: Return target Kavu to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			Targets: effects.TargetCreature("target Kavu", effects.HasSubtype("Kavu")),
			Effect:  effects.BounceTheTargetToHand,
		}},
	})
}
