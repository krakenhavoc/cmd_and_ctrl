package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elf Replica — Artifact Creature — Elf {3}, 2/2 (Mirrodin, #5606):
//
//	"{1}{G}, Sacrifice this creature: Destroy target enchantment."
//
// Felidar Cub's ability with mana added to the cost: the sacrifice is
// paid at announce, so the Replica is gone before the destruction
// resolves; the target is checked again as it resolves (CR 608.2b), and
// an indestructible enchantment survives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2dda5011-aaa1-48d0-afa8-76998278ab75",
		Name:         "Elf Replica",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, Sacrifice this creature: Destroy target enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target enchantment", effects.Enchantment()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
