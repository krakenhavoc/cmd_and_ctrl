package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krovikan Elementalist — Creature — Human Wizard {B}{B}, 1/1 (Ice Age, S82 OLD,
// #5636):
//
//	"{2}{R}: Target creature gets +1/+0 until end of turn.
//	 {U}{U}: Target creature you control gains flying until end of turn.
//	 Sacrifice it at the beginning of the next end step."
//
// The second ability's sacrifice is a delayed trigger pinned to the creature as
// the object it is now (CR 603.7c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43e6fc44-8b23-4bff-91da-936019efe84b",
		Name:         "Krovikan Elementalist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: Target creature gets +1/+0 until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)),
			Cost:    effects.ManaCost("{2}{R}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.OldB834TargetGets(1, 0, "Krovikan Elementalist — +1/+0"),
		}, {
			Label:   "{U}{U}: Target creature you control gains flying until end of turn. Sacrifice it at the beginning of the next end step.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    effects.ManaCost("{U}{U}"),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect:  effects.OldB834TargetFliesThenSacrifice("Krovikan Elementalist"),
		}},
	})
}
