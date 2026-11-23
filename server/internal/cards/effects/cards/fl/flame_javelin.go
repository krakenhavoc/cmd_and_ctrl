package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flame Javelin — Instant {2/R}{2/R}{2/R} (Shadowmoor, #5591):
//
//	"({2/R} can be paid with any two mana or with {R}. This card's mana
//	 value is 6.)
//	 Flame Javelin deals 4 damage to any target."
//
// Puncture Blast's damage row without wither: 4 to any target, the target
// re-checked at resolution (CR 608.2b). Each {2/R} is paid with {R} or two
// mana (CR 107.4e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58363090-bf89-4174-bbb2-636afc1c7a06",
		Name:         "Flame Javelin",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve:    effects.DamageToFirstTargetStep(4),
	})
}
