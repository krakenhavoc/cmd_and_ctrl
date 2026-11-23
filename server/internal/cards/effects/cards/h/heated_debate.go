package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heated Debate — Instant {2}{R} (Strixhaven: School of Mages, #5502):
//
//	"This spell can't be countered. (This includes by the ward ability.)
//	 Heated Debate deals 4 damage to target creature or planeswalker."
//
// Sear's 4 damage to a creature or planeswalker, with Abrupt Decay's
// can't-be-countered flag, which a ward trigger's counter respects too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "06c3e85c-17d8-44e1-a0f8-07ef3f1624cb",
		Name:            "Heated Debate",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Targets:         effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:         effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve:       effects.DamageToFirstTargetStep(4),
	})
}
