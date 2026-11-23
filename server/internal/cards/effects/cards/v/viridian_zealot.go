package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viridian Zealot — Creature — Elf Warrior {G}{G} (Darksteel, #5605):
//
//	"{1}{G}, Sacrifice this creature: Destroy target artifact or
//	 enchantment."
//
// Caustic Caterpillar's ability: the sacrifice is part of the cost, so
// the Zealot is gone before the ability resolves (CR 602.2b), and the
// shared destroyFirstLegalTarget body re-checks the target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9410ae41-b6db-47c6-ae0a-4e33aad1d3ce",
		Name:         "Viridian Zealot",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
