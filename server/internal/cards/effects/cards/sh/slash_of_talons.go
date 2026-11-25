package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slash of Talons — Instant {W} (Ixalan, #5532):
//
//	"Slash of Talons deals 2 damage to target attacking or blocking
//	 creature."
//
// The target is judged again as the spell resolves: a creature removed
// from combat in response is illegal and the spell fizzles (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef331f13-bdd0-4123-9e61-b2e4b6b6d8b8",
		Name:         "Slash of Talons",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve:    effects.DamageToFirstTargetStep(2),
	})
}
