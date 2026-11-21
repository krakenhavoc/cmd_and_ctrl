package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Touchstone — Artifact {2} (Weatherlight, #5629):
//
//	"{T}: Tap target artifact you don't control."
//
// Auriok Transfixer's tap with no mana and an artifact another player
// controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92382bdb-84cf-4456-8ca1-8a19badae7ad",
		Name:         "Touchstone",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Tap target artifact you don't control.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target artifact you don't control", effects.Artifact(), effects.OpponentControls()),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
