package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tamiyo's Immobilizer — Artifact {3}{U} (Phyrexia: All Will Be One):
//
//	"This artifact enters with four oil counters on it.
//	 {T}, Remove an oil counter from this artifact: Tap target artifact or
//	 creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2c7a9c2-28f8-4cf0-a042-420f6a4ba8bd",
		Name:         "Tamiyo's Immobilizer",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters(effects.OilCounter, 4, "Tamiyo's Immobilizer: enters with four oil counters on it")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove an oil counter from this artifact: Tap target artifact or creature.",
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis(effects.OilCounter, 1)),
			Targets: effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
