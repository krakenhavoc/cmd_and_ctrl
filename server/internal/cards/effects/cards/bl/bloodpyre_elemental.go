package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodpyre Elemental — {4}{R} Creature — Elemental 4/1 (Shards of Alara, S82
// OLD b515):
//
//	"Sacrifice this creature: It deals 4 damage to target creature. Activate
//	 only as a sorcery."
//
// Sacrificed as a cost, so the damage is dealt by its last-known information
// (CR 608.2h); the target is chosen before the sacrifice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c683855-f32e-42f9-8681-c3335d5eaf79",
		Name:         "Bloodpyre Elemental",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "Sacrifice this creature: It deals 4 damage to target creature. Activate only as a sorcery.",
			Cost:         effects.SacrificeThis(),
			Targets:      effects.TargetCreature("target creature"),
			SorcerySpeed: true,
			Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
			Effect:       effects.DealDamageToTheTarget(4),
		}},
	})
}
