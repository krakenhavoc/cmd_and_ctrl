package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forgestoker Dragon — Creature — Dragon {4}{R}{R}, 5/4 (Born of the Gods, S82
// OLD b378, #5559):
//
//	"Flying
//	 {1}{R}: This creature deals 1 damage to target creature. That creature
//	 can't block this combat. Activate only if this creature is attacking."
//
// Glint-Horn Buccaneer's "only if this creature is attacking" condition, the
// ping, and a CR 509.1b can't-block that ends with combat, applied even when
// the damage is prevented (S82OldB378ForgestokerPing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "53de1e6f-9471-4cdc-ac67-66d47da4d7eb",
		Name:            "Forgestoker Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     "{1}{R}: This creature deals 1 damage to target creature. That creature can't block this combat. Activate only if this creature is attacking.",
			Cost:      effects.ManaCost("{1}{R}"),
			Targets:   effects.TargetCreature("target creature"),
			Purpose:   effects.ForTargets(effects.DamageToTarget(0, 1)),
			Condition: effects.SourceIsAttacking(),
			Effect:    effects.S82OldB378ForgestokerPing,
		}},
	})
}
