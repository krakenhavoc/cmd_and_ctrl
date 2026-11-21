package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Neurok Replica — Artifact Creature — Wizard {3}, 1/4 (Scars of Mirrodin, #5577):
//
//	"{1}{U}, Sacrifice this creature: Return target creature to its owner's hand."
//
// Seal of Removal's bounce with a mana cost: the Replica is sacrificed
// as a cost, so it can return any creature but itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de8cdbd4-d8f7-4ce8-a69b-6d2020beff63",
		Name:         "Neurok Replica",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, Sacrifice this creature: Return target creature to its owner's hand.",
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BounceTheTargetToHand,
		}},
	})
}
