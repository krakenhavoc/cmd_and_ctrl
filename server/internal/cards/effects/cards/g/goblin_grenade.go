package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Grenade — Sorcery {R} (Jumpstart 2022, #5948):
//
//	"As an additional cost to cast this spell, sacrifice a Goblin.
//	 Goblin Grenade deals 5 damage to any target."
//
// The sacrifice is paid as the spell is cast (Betrayer's Bargain's cost
// shape), so a table with no Goblin cannot cast it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "20aecb7b-7e37-4efe-ac27-f9f9042203f4",
		Name:           "Goblin Grenade",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a Goblin", effects.HasSubtype("Goblin")),
		Targets:        effects.TargetAny(),
		Purpose:        effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve:      effects.DamageToFirstTargetStep(5),
	})
}
