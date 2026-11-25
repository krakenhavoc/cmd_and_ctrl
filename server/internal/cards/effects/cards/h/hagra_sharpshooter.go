package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hagra Sharpshooter — Creature — Human Assassin Ally {2}{B}, 2/2 (Battle for
// Zendikar):
//
//	"{4}{B}: Target creature gets -1/-1 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2f59fa6-8e61-4a70-8044-196d7b26ec9e",
		Name:         "Hagra Sharpshooter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{B}: Target creature gets -1/-1 until end of turn.",
			Cost:    effects.ManaCost("{4}{B}"),
			Targets: effects.Harmful(effects.TargetCreature("target creature")),
			Effect:  effects.BoostTargetUntilEOT("Hagra Sharpshooter — -1/-1 until end of turn", -1, -1),
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
		}},
	})
}
