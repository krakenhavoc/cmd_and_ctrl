package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Capashen Unicorn — Creature — Unicorn {1}{W}, 1/2 (Invasion, #5616):
//
//	"{1}{W}, {T}, Sacrifice this creature: Destroy target artifact or
//	 enchantment."
//
// The tap and the sacrifice are paid at announce, so the Unicorn is gone
// before the ability resolves; the destroy does not need it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52793400-bc83-402a-9609-928a4d3ac812",
		Name:         "Capashen Unicorn",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
