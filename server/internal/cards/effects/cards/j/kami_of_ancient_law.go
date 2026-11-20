package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kami of Ancient Law — Creature — Spirit {1}{W}, 2/2:
//
//	"Sacrifice this creature: Destroy target enchantment."
//
// One CR 602 ability whose only cost is the sacrifice, paid at announce,
// so the Kami is gone before the destruction resolves; the target is
// re-checked at resolution and an indestructible enchantment survives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e937e0f5-ca35-4dce-9f05-8f2075731628",
		Name:         "Kami of Ancient Law",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice Kami of Ancient Law: Destroy target enchantment.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetPermanent("target enchantment", effects.Enchantment()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
