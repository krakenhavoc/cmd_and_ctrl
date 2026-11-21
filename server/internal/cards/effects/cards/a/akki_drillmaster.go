package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akki Drillmaster — Creature — Goblin Shaman {2}{R}, 2/2 (Saviors of
// Kamigawa, #5601):
//
//	"{T}: Target creature gains haste until end of turn."
//
// A layer-6 grant until the cleanup step (CR 514.2); a target that left
// in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa228a01-7de3-4539-9ebb-b3c546f72b41",
		Name:         "Akki Drillmaster",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature gains haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NeoB07TargetGetsUntilEOT("Akki Drillmaster — haste until end of turn", 0, 0, "haste"),
		}},
	})
}
