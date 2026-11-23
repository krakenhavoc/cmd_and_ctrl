package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inspired Insurgent — Creature — Human Peasant Ally {1}{W}, 2/2
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"{1}, Sacrifice this creature: Destroy target artifact or enchantment."
//
// Cathar Commando's ability (destroyChosenPermanent): the sacrifice is a
// cost, so the destroy still happens after the Insurgent is gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5863c67b-a47c-4466-8bd1-210ab26e0d3f",
		Name:         "Inspired Insurgent",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
