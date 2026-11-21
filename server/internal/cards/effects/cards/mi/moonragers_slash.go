package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moonrager's Slash — {2}{R} Instant (#2561, ADR 0132):
//
//	"This spell costs {2} less to cast if it's night.
//	 Moonrager's Slash deals 3 damage to any target."
//
// The discount reads the designation as the spell is priced, so it is
// there on the turn night begins and gone the moment it is day again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40a41774-fe50-438d-85e4-7f2beb29fb82",
		Name:         "Moonrager's Slash",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if it's night.", effects.ItsNightCost()),
		},
		OnResolve: effects.DamageToFirstTargetStep(3),
	})
}
