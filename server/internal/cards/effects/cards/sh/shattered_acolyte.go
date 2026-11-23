package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shattered Acolyte — Creature — Dwarf Warlock {1}{W}, 2/2 (Secrets of
// Strixhaven, #4456):
//
//	"Lifelink
//	 {1}, Sacrifice this creature: Destroy target artifact or enchantment."
//
// Cathar Commando's ability (destroyChosenPermanent): the sacrifice is a cost,
// so the destroy still happens after the Acolyte is gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "57f58d52-0014-4cac-94fc-668063cc147a",
		Name:            "Shattered Acolyte",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
