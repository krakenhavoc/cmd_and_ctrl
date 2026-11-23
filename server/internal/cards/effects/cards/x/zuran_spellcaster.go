package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zuran Spellcaster — Creature — Human Wizard {2}{U}, 1/1 (Ice Age,
// #5636):
//
//	"{T}: This creature deals 1 damage to any target."
//
// Capricious Sorcerer's ping with no timing window: any target, at
// instant speed, the shared sourceDealsDamageToEachLegalTarget body. A
// target that is no longer legal as it resolves is left alone
// (CR 608.2b). The tap means summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc7b90b1-3517-4e5d-9bd8-68b4d8a259fd",
		Name:         "Zuran Spellcaster",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
