package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Razorfin Abolisher — {2}{U} Creature — Merfolk Wizard, 2/2:
//
//	"{1}{U}, {T}: Return target creature with a counter on it to its owner's
//	 hand."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38a2ded4-77c5-4d2b-b411-7dca8108aa02",
		Name:         "Razorfin Abolisher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {T}: Return target creature with a counter on it to its owner's hand.",
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature with a counter on it", effects.B64HasAnyCounter()),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  effects.BounceTheTargetToHand,
		}},
	})
}
