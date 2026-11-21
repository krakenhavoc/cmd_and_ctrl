package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Screeching Griffin — Creature — Griffin {3}{W}, 2/2 (Ravnica: City of
// Guilds, #5600):
//
//	"Flying
//	 {R}: Target creature can't block this creature this turn."
//
// Burning-Tree Bloodscale's "can't block this creature" row: the one
// target can't block the Griffin this turn, and may still block others.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "beae83d2-20b8-47ef-83d9-25a7f90d1da2",
		Name:            "Screeching Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Target creature can't block this creature this turn.",
			Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{R}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetCantBlockThisCreature("Screeching Griffin — can't block it"),
		}},
	})
}
