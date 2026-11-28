package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heartwood Giant — Creature — Giant {3}{G}{G}, 4/4 (Tempest, S82 OLD
// batch 758, #5628):
//
//	"{T}, Sacrifice a Forest: This creature deals 2 damage to target
//	 player or planeswalker."
//
// Foratog's "Sacrifice a Forest" cost joined to a tap, and Zealot of the
// God-Pharaoh's damage to a player or planeswalker. The Forest is sacrificed
// as a cost, so the damage happens even if the Giant is gone (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b7d7418d-222b-4ae0-b0d9-9ea71d6ad800",
		Name:         "Heartwood Giant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice a Forest: This creature deals 2 damage to target player or planeswalker.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "a Forest", effects.Subtype("Forest"))),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Effect:  effects.DealDamageToTheTarget(2),
		}},
	})
}
