package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flame Elemental — Creature — Elemental {2}{R}{R}, 3/2 (Mirage, #5633):
//
//	"{R}, {T}, Sacrifice this creature: It deals damage equal to its power to
//	 target creature."
//
// The damage is its last-known power as it left the battlefield (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "975226a4-2d03-4993-9337-cfd6595011bd",
		Name:         "Flame Elemental",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}, Sacrifice this creature: It deals damage equal to its power to target creature.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.SacrificedSourceDealsItsPowerToTheTarget,
		}},
	})
}
