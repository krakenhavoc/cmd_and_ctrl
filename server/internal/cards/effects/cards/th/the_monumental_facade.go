package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Monumental Facade — Land — Sphere (Phyrexia: All Will Be One, #5479):
//
//	"This land enters with two oil counters on it.
//	 {T}: Add {C}.
//	 {T}, Remove an oil counter from this land: Put an oil counter on
//	 target artifact or creature you control. Activate only as a sorcery."
//
// The two counters are an entry replacement. The removal is a cost, so a
// land with none can't activate; the target is an artifact or creature
// you control, re-checked on resolution, and the counter is lost with the
// cost if it became illegal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63e8d282-d038-4a8c-a0cb-f51ddf87d8ea",
		Name:         "The Monumental Facade",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(effects.S82OneOilCounter, 2, "The Monumental Facade: enters with two oil counters"),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}, Remove an oil counter from this land: Put an oil counter on target artifact or creature you control. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis(effects.S82OneOilCounter, 1)),
			Targets:      effects.S82OneYourArtifactOrCreature(),
			SorcerySpeed: true,
			Effect:       effects.S82OnePutOilOnTarget,
		}},
	})
}
