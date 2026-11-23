package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ember Hauler — Creature — Goblin {R}{R}, 2/2 (Magic 2011, #5579):
//
//	"{1}, Sacrifice this creature: It deals 2 damage to any target."
//
// Goblin Legionnaire's damage row at {1}: the sacrifice is part of the
// cost, and the damage is dealt by the Hauler as it last existed on the
// battlefield (CR 113.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d313958-49ca-4592-97c7-20f12c8468a8",
		Name:         "Ember Hauler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: It deals 2 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeThis()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.B33DamageChosenTargetFromSource(2),
		}},
	})
}
