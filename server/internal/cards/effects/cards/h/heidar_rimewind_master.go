package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heidar, Rimewind Master — Legendary Creature — Human Wizard {4}{U}, 3/3
// (Coldsnap, #5597):
//
//	"{2}, {T}: Return target permanent to its owner's hand. Activate only if
//	 you control four or more snow permanents."
//
// Temporal Adept's bounce with an activation condition (CR 602.5b) over the
// activator's snow permanents, as the ability is announced.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58ac4cce-506e-4d6e-a19c-0e875fbfcd81",
		Name:         "Heidar, Rimewind Master",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{2}, {T}: Return target permanent to its owner's hand. Activate only if you control four or more snow permanents.",
			Purpose:   game.Purpose{Answers: game.AnswerRemove},
			Cost:      effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Condition: effects.S82OldB567SnowPermanentsAtLeast(4),
			Targets:   effects.TargetPermanent("target permanent"),
			Effect:    effects.BounceTheTargetToHand,
		}},
	})
}
