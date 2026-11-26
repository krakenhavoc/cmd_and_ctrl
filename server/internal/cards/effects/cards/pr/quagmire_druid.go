package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quagmire Druid — Creature — Zombie Druid {2}{B}, 2/2 (Commander 2013, S82
// older sets):
//
//	"{G}, {T}, Sacrifice a creature: Destroy target enchantment."
//
// The sacrificed creature is chosen as a cost (CR 118.3, 602.2b) and may be
// the Druid itself; the target enchantment is chosen as the ability goes on
// the stack and judged again as it resolves (CR 608.2b). A black creature
// that taps for {G} needs another source of green mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "560e1e81-6675-4d08-8e82-cdf1abd4b3d0",
		Name:         "Quagmire Druid",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}, Sacrifice a creature: Destroy target enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost(), effects.SacrificeN(1, "a creature", effects.Creature())),
			Targets: effects.TargetPermanent("target enchantment", effects.Enchantment()),
			Effect:  effects.DestroyFirstLegalCardTarget,
		}},
	})
}
