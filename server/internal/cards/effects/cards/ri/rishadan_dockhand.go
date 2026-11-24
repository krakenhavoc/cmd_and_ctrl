package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rishadan Dockhand — Creature — Merfolk {U}, 1/2 (Modern Horizons 2,
// #5501):
//
//	"Islandwalk
//	 {1}, {T}: Tap target land."
//
// Islandwalk is the engine's landwalk (CR 702.14c); the tap ability is
// Rishadan Port's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3d199876-09f3-4e14-8598-79f0ec2afafa",
		Name:            "Rishadan Dockhand",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Tap target land.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
