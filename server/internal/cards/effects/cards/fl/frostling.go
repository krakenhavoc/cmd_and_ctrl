package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frostling — Creature — Spirit {R}, 1/1 (Betrayers of Kamigawa, #5602):
//
//	"Sacrifice this creature: It deals 1 damage to target creature."
//
// Ticking Gnomes' sacrifice ability with a narrower target. The sacrifice
// is part of the cost, so the damage comes from the Frostling as it last
// existed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f6ac768-4c97-44db-933a-f2d442e7f665",
		Name:         "Frostling",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: It deals 1 damage to target creature.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsOneToFirstTarget,
		}},
	})
}
