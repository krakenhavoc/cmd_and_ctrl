package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seal of Primordium — Enchantment {1}{G} (EDHREC rank 3667):
//
//	"Sacrifice this enchantment: Destroy target artifact or
//	 enchantment."
//
// Seal of Cleansing in green, line for line: Naturalize on a
// permanent, paid for up front and cashed at instant speed later.
// A sacrifice-this activation with no mana and no tap, the target
// picked at announce and re-checked at resolution (CR 608.2b); the
// Seal is sacrificed at announce, so it can never be its own target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f14dbb39-c9f9-4f64-b22a-38dba28f5b1e",
		Name:         "Seal of Primordium",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Destroy target artifact or enchantment.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
