package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arrows of Justice — Instant {2}{R/W} (Gatecrash, #5565):
//
//	"Arrows of Justice deals 4 damage to target attacking or blocking
//	 creature."
//
// Divine Arrow's spell: 4 damage to the one target, which must still be
// attacking or blocking as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46e311ef-c86e-4aad-8b42-ef0d68a90c2e",
		Name:         "Arrows of Justice",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		OnResolve:    effects.DamageToFirstTargetStep(4),
	})
}
