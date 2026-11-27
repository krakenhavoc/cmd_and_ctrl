package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Advance Scout — Creature — Human Soldier Scout {1}{W}, 1/1 (Tempest, S82
// OLD b754, #5628):
//
//	"First strike
//	 {W}: Target creature gains first strike until end of turn."
//
// Battle Rampart's targeted grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "831e2185-0bb2-45bb-99cd-d2091bad86f0",
		Name:            "Advance Scout",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: Target creature gains first strike until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike")),
			Cost:    effects.ManaCost("{W}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NeoB07TargetGetsUntilEOT("Advance Scout — first strike until end of turn", 0, 0, "first strike"),
		}},
	})
}
