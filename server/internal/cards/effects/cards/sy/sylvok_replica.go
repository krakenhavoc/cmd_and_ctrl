package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sylvok Replica — Artifact Creature — Shaman {3}, 1/3 (Scars of Mirrodin, #5577):
//
//	"{G}, Sacrifice this creature: Destroy target artifact or enchantment."
//
// Thrashing Brontodon's ability at {G}. The Replica is sacrificed as a
// cost, so it can't be its own target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97818ed4-3590-4a3b-92ce-716f9cfdc78e",
		Name:         "Sylvok Replica",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
