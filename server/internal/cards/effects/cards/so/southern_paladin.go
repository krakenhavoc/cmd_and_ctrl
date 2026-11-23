package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Southern Paladin — Creature — Human Knight {2}{W}{W}, 3/3 (Weatherlight, #5629):
//
//	"{W}{W}, {T}: Destroy target red permanent."
//
// The target's colour is read at activation and again at resolution
// (CR 608.2b), so a permanent that stopped being red is not destroyed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cab96aa8-13e3-4dd4-9d44-9e298f807388",
		Name:         "Southern Paladin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{W}, {T}: Destroy target red permanent.",
			Cost:    effects.Plus(effects.ManaCost("{W}{W}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target red permanent", effects.OfColor("R")),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
