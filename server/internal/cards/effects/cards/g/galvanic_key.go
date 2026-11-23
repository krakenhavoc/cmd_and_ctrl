package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galvanic Key — Artifact {2} (Mirrodin, #5606):
//
//	"Flash
//	 {3}, {T}: Untap target artifact."
//
// Voltaic Key at a higher price, with flash (CR 702.8a): it can be cast
// whenever an instant could. "Target artifact" admits the Key itself, as
// printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d8a552ca-2c7b-410e-bd7e-1bb81465277a",
		Name:            "Galvanic Key",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Untap target artifact.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect:  effects.UntapFirstLegalTarget,
		}},
	})
}
