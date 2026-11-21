package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maraleaf Rider — Creature — Elf Knight {1}{G}, 3/1:
//
//	"Sacrifice a Food: Target creature blocks this creature this turn if
//	 able."
//
// The cost is Hollow Scavenger's "Sacrifice a Food" (any Food artifact
// you control); the effect is Burning-Tree Bloodscale's requirement,
// which waits for the Rider to attack and asks nothing if it never does
// (CR 509.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66412a7a-4ec0-41f0-80ad-eb483213ff81",
		Name:         "Maraleaf Rider",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a Food: Target creature blocks this creature this turn if able.",
			Cost:    game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Food", effects.HasSubtype("Food"))},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetBlocksThisCreature,
		}},
	})
}
