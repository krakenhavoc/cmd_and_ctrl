package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gnarled Effigy — Artifact {4} (Shadowmoor, S82 OLD b532, #5591):
//
//	"{4}, {T}: Put a -1/-1 counter on target creature."
//
// The counter goes on through the effect's placement, so a counter doubler
// applies (CR 614.16). A target that has left or stopped being legal does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5fa3cc1a-09ee-4152-8c50-ea719510fc20",
		Name:         "Gnarled Effigy",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Put a -1/-1 counter on target creature.",
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.R35NightmareSowerCounter,
		}},
	})
}
