package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flamecast Wheel — Artifact {1} (Theros):
//
//	"{5}, {T}, Sacrifice this artifact: It deals 3 damage to target creature."
//
// The sacrifice is a cost, so the damage is dealt by the Wheel as it last
// existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58f70225-17b5-4aa6-a62a-89c57e31ecc1",
		Name:         "Flamecast Wheel",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}, Sacrifice this artifact: It deals 3 damage to target creature.",
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Effect:  effects.B33DamageChosenTargetFromSource(3),
		}},
	})
}
