package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghost Warden — Creature — Spirit {1}{W}, 1/1 (Guildpact, #5599):
//
//	"{T}: Target creature gets +1/+1 until end of turn."
//
// Wyluli Wolf's ability, written with Icatian Priest's shared body
// (pumpAndGrantFirstTarget): the still-legal target gets the boost,
// nothing if it left in response (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66bd5999-742a-43e0-827c-ecb8ec7db242",
		Name:         "Ghost Warden",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature gets +1/+1 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.PumpAndGrantFirstTarget(1, 1, "Ghost Warden — +1/+1 until end of turn"),
		}},
	})
}
