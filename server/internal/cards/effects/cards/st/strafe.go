package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strafe — Sorcery {R} (Planeshift, S82 OLD #5615):
//
//	"Strafe deals 3 damage to target nonred creature."
//
// One target clause (a creature that is not red, judged at announce and again
// on resolution, CR 608.2b) and the shared damage-to-the-first-target body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ac7990b-b86e-4725-974c-d6affa2af2f9",
		Name:         "Strafe",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		Targets:      effects.TargetCreature("target nonred creature", effects.Not(effects.OfColor("R"))),
		OnResolve:    effects.DamageToFirstTargetStep(3),
	})
}
