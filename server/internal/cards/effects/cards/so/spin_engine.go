package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spin Engine — Artifact Creature — Construct {3}, 3/1 (Mirrodin Besieged, S82
// OLD b465, #5576):
//
//	"{R}: Target creature can't block this creature this turn."
//
// Burning-Tree Bloodscale's {2}{R} row: the engine's record bars only blocks
// of this creature, and the target is read again at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88e98269-fe0e-42f8-830d-776aa06a58dc",
		Name:         "Spin Engine",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Target creature can't block this creature this turn.",
			Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{R}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetCantBlockThisCreature("Spin Engine — can't block it"),
		}},
	})
}
