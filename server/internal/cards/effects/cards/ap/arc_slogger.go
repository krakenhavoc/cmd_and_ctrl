package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arc-Slogger — Creature — Beast {3}{R}{R}, 4/5:
//
//	"{R}, Exile the top ten cards of your library: This creature deals 2
//	 damage to any target."
//
// ADR 0109 §7 (#1902): "Exile the top N cards of your library" is a
// cost with nothing to choose. A library of fewer than N cards can't pay
// it (CR 118.3), and it is paid after every other cost (CR 601.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a28fc506-a98c-4b4d-a0f4-7d971489ccbb",
		Name:         "Arc-Slogger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Exile the top ten cards of your library: This creature deals 2 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.ExileTopOfLibrary(10)),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
