package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "66412a7a-4ec0-41f0-80ad-eb483213ff81",
		Name:         "Maraleaf Rider",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a Food: Target creature blocks this creature this turn if able.",
			Cost:    game.AbilityCost{SacrificeOther: sacrificeSpec("a Food", HasSubtype("Food"))},
			Targets: TargetCreature("target creature"),
			Effect:  TargetBlocksThisCreature,
		}},
	})
}
