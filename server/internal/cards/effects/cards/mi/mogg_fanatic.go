package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mogg Fanatic — {R} Creature — Goblin 1/1 (Eternal Masters, S82 OLD, #5542):
//
//	"Sacrifice this creature: It deals 1 damage to any target."
//
// Fire Bowman's sacrifice-for-1 without the timing window: the sacrifice is a
// cost, so the damage is dealt by the creature as it last existed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7efd5d62-6da2-428a-98e3-5b842f668656",
		Name:         "Mogg Fanatic",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: It deals 1 damage to any target.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.B33DamageChosenTargetFromSource(1),
		}},
	})
}
