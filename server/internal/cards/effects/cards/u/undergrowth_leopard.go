package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undergrowth Leopard — Creature — Cat {1}{G}, 2/2 (Tarkir: Dragonstorm,
// #4463):
//
//	"Vigilance
//	 {1}, Sacrifice this creature: Destroy target artifact or enchantment."
//
// Vigilance is a printed keyword; the activated ability is Caustic
// Caterpillar's with a {1} cost. The Leopard is sacrificed as the ability
// is activated, and the target is checked again as it resolves; an
// indestructible pick survives, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "511fa8a3-0a65-47fc-9f3b-57eaef2698cd",
		Name:            "Undergrowth Leopard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
