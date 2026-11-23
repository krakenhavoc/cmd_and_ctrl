package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scepter of Dominance — Artifact {1}{W}{W} (Conflux, #5587):
//
//	"{W}, {T}: Tap target permanent."
//
// Auriok Transfixer's row with any permanent as the target. An
// artifact's {T} is not held back by summoning sickness (CR 302.6), and
// it may target a permanent that is already tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60bf80f9-78b1-4869-afd9-e769b1b0f5ea",
		Name:         "Scepter of Dominance",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Tap target permanent.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target permanent"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
