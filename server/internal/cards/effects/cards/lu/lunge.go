package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lunge — Instant {2}{R} (Mercadian Masques, S82 OLD, #5619):
//
//	"Lunge deals 2 damage to target creature and 2 damage to target player or
//	 planeswalker."
//
// Hungry Flames's two clauses at 2 and 2: one damage instance, and a clause
// whose target has left is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e642112-9592-475f-ba03-d71544270cfa",
		Name:         "Lunge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.Clauses(effects.TargetCreature("target creature"), effects.TargetPlayerOrPlaneswalker()),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2), effects.DamageToTarget(1, 2)),
		OnResolve:    effects.DamageEachClauseTarget(2, 2),
	})
}
