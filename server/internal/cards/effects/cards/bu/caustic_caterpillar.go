package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Caustic Caterpillar — Creature — Insect {G}, 1/1 (EDHREC rank
// 3436):
//
//	"{1}{G}, Sacrifice this creature: Destroy target artifact or
//	 enchantment."
//
// A Naturalize on legs. One CR 602 ability with a mana-plus-self-
// sacrifice cost, paid at announce so the Caterpillar's own
// dies-triggers resolve above the destruction; the target is
// re-checked at resolution and an indestructible pick survives, as
// printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45d35128-76e6-43f9-8d23-41f7506c3a71",
		Name:         "Caustic Caterpillar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, Sacrifice Caustic Caterpillar: Destroy target artifact or enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
