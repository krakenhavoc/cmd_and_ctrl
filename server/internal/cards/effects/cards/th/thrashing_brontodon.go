package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thrashing Brontodon — Creature — Dinosaur {1}{G}{G}, 3/4 (EDHREC
// rank 2799):
//
//	"{1}, Sacrifice this creature: Destroy target artifact or
//	 enchantment."
//
// The green deck's body that is also its Naturalize. One activated
// ability, sacrificing the Brontodon as its cost — so it dies with
// the ability on the stack and a Blood Artist drains before the
// artifact is destroyed — through the single-target verb that
// honours indestructible. A target that left in response makes the
// ability do nothing, and the Brontodon is gone either way, as
// printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60bc63dc-ac9f-4a2f-aef5-c90d0aa31553",
		Name:         "Thrashing Brontodon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
