package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phantasmal Mount — Creature — Illusion Horse {1}{U}, 1/1 (Ice Age, S82 OLD,
// #5636):
//
//	"Flying
//	 {T}: Target creature you control with toughness 2 or less gets +1/+1 and
//	 gains flying until end of turn. When this creature leaves the battlefield
//	 this turn, sacrifice that creature. When the creature leaves the battlefield
//	 this turn, sacrifice this creature."
//
// Two delayed leaves-the-battlefield triggers (Runesword's shape), one watching
// the Mount and sacrificing the creature, one watching the creature and
// sacrificing the Mount, each lasting this turn and pinned to the object as it
// is now (CR 603.7c, 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8c7723a5-04c7-4241-b0ba-6a821449ca43",
		Name:            "Phantasmal Mount",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature you control with toughness 2 or less gets +1/+1 and gains flying until end of turn. When this creature leaves the battlefield this turn, sacrifice that creature. When the creature leaves the battlefield this turn, sacrifice this creature.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "flying")),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature you control with toughness 2 or less", effects.And(effects.YouControl(), effects.ToughnessLE(2))),
			Effect:  effects.OldB834PhantasmalMount,
		}},
	})
}
