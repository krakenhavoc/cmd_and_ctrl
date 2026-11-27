package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghitu War Cry — Enchantment {2}{R} (Urza's Legacy, S82 OLD b715, #5623):
//
//	"{R}: Target creature gets +1/+0 until end of turn."
//
// Sacred Armory's pump at {R} (NeoB07TargetGetsUntilEOT): no {T}, so it
// stacks with each activation, and a target that left is skipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0439df39-0324-4110-b3f4-4a32393021de",
		Name:         "Ghitu War Cry",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Target creature gets +1/+0 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)),
			Cost:    effects.ManaCost("{R}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NeoB07TargetGetsUntilEOT("Ghitu War Cry — +1/+0 until end of turn", 1, 0),
		}},
	})
}
