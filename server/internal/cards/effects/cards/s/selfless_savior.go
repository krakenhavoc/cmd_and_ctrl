package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Selfless Savior — Creature — Dog {W}, 1/1:
//
//	"Sacrifice this creature: Another target creature you control gains
//	 indestructible until end of turn."
//
// The sacrifice is paid at announce. "Another" is object identity
// (CR 109.1, effects.Another), so the Savior can't name itself; the
// target must still be a creature you control as the ability resolves
// (CR 608.2b), or nothing happens.
//
// No simplification.
func init() {
	const label = "Selfless Savior — another creature you control gains indestructible until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "49dc3434-2501-4788-84f2-93e55730f67c",
		Name:         "Selfless Savior",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice Selfless Savior: Another target creature you control gains indestructible until end of turn.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
			Effect:  effects.FirstTargetGainsUntilEndOfTurn(0, label, "indestructible"),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")),
		}},
	})
}
