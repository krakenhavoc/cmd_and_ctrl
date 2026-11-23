package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saltfield Recluse — Creature — Human Rebel Cleric {2}{W}, 1/2 (Planar Chaos, #5595):
//
//	"{T}: Target creature gets -2/-0 until end of turn."
//
// Krosan Constrictor's -2/-0 on any creature. Declared as a harm, so it is
// aimed at an opponent's creature; the shrink ends at cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c87e056a-6067-49e6-b3f7-bba8234777de",
		Name:         "Saltfield Recluse",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature gets -2/-0 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -2, 0)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BoostTargetUntilEOT("Saltfield Recluse — -2/-0 until end of turn", -2, 0),
		}},
	})
}
