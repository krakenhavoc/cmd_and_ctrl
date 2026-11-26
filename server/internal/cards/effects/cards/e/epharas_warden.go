package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ephara's Warden — Creature — Human Cleric {3}{W}, 1/2 (Theros):
//
//	"{T}: Tap target creature with power 3 or less."
//
// The target is legal only while its power, read through the layered view, is
// 3 or less (CR 608.2b re-checks it on resolution).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "805d1f48-dd91-4034-b42c-3f1f1cb75cc9",
		Name:         "Ephara's Warden",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Tap target creature with power 3 or less.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature with power 3 or less", effects.PowerLE(3)),
			Effect:  effects.TapTheFirstLegalTarget,
		}},
	})
}
