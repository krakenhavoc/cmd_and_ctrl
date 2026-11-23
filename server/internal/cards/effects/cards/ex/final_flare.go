package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Final Flare — Instant {2}{R} (Theros Beyond Death, #5514):
//
//	"As an additional cost to cast this spell, sacrifice a creature
//	 or enchantment.
//	 Final Flare deals 5 damage to target creature."
//
// Final Vengeance's additional cost (a creature or an enchantment, paid
// as the spell is cast, CR 601.2f) and Scorching Shot's 5 damage to
// target creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "34f9a03e-ca95-4d3e-9e63-7297b5513481",
		Name:           "Final Flare",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature or enchantment", effects.Or(effects.Creature(), effects.Enchantment())),
		Targets:        effects.TargetCreature("target creature"),
		Purpose:        effects.ForTargets(effects.DamageToTarget(0, 5)),
		OnResolve:      effects.DamageToFirstTargetStep(5),
	})
}
