package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plagued Rusalka — Creature — Spirit {B}, 1/1 (Jumpstart, #4892):
//
//	"{B}, Sacrifice a creature: Target creature gets -1/-1 until end of
//	 turn."
//
// The sacrifice is a cost, paid at announcement; a target that has left
// by resolution makes the ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6912f52-e478-4cc6-a34c-67e372abc4fa",
		Name:         "Plagued Rusalka",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Sacrifice a creature: Target creature gets -1/-1 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.SacrificeACreature()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BoostTargetUntilEOT("Plagued Rusalka — -1/-1", -1, -1),
		}},
	})
}
