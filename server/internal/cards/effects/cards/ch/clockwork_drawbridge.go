package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clockwork Drawbridge — Artifact Creature — Wall {W}, 0/3 (Dominaria United, S82):
//
//	"Defender
//	 {2}{W}, {T}: Tap target creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e49eb21c-2a0d-4fd0-9f69-1bff6e94f9e7",
		Name:            "Clockwork Drawbridge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
