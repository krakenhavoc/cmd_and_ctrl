package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keening Apparition — Creature — Spirit {1}{W}, 2/2 (Return to Ravnica,
// #5566):
//
//	"Sacrifice this creature: Destroy target enchantment."
//
// Ronom Unicorn's ability: the sacrifice is paid at announce, the target is
// re-checked at resolution, and an indestructible enchantment survives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9212596-3f65-41c9-9ee8-1f4cec7f0ecf",
		Name:         "Keening Apparition",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Destroy target enchantment.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetPermanent("target enchantment", effects.Enchantment()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
