package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anointer of Champions — Creature — Human Cleric {W}, 1/1 (Magic Origins, #5549):
//
//	"{T}: Target attacking creature gets +1/+1 until end of turn."
//
// Infantry Veteran's tap: the target must be an attacking creature at
// activation and at resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f74ef5b0-9222-43d0-a919-aaf28dc21b59",
		Name:         "Anointer of Champions",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target attacking creature gets +1/+1 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
			Effect:  effects.PumpAndGrantFirstTarget(1, 1, "Anointer of Champions — +1/+1 until end of turn"),
		}},
	})
}
