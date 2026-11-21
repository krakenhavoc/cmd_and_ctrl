package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thundering Rebuke — Sorcery {1}{R} (Zendikar Rising, #5508):
//
//	"Thundering Rebuke deals 4 damage to target creature or planeswalker."
//
// Sear's text at sorcery speed (the speed rides the type line).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb4d9770-f230-4413-a7a4-bb1f1303661a",
		Name:         "Thundering Rebuke",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve:    effects.DamageToFirstTargetStep(4),
	})
}
