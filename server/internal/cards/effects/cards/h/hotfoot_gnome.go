package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hotfoot Gnome — Artifact Creature — Gnome {2}{R}, 3/1 (The Lost Caverns
// of Ixalan, #4470):
//
//	"Haste
//	 {T}: Another target creature gains haste until end of turn."
//
// Axgard Cavalry's row with "another": the Gnome cannot target itself. A
// target that has gone by resolution gains nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3efb64a4-04f6-4b78-be1c-5b6320c02a71",
		Name:            "Hotfoot Gnome",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Another target creature gains haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.TapCost(),
			Targets: effects.Another(effects.TargetCreature("another target creature")),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("haste", "Hotfoot Gnome — haste until end of turn"),
		}},
	})
}
