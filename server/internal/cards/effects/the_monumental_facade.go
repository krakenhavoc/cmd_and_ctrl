package effects

import (
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
	Register(Spec{
		OracleID:     "63e8d282-d038-4a8c-a0cb-f51ddf87d8ea",
		Name:         "The Monumental Facade",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(s82OneOilCounter, 2, "The Monumental Facade: enters with two oil counters"),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{T}, Remove an oil counter from this land: Put an oil counter on target artifact or creature you control. Activate only as a sorcery.",
			Cost:         Plus(TapCost(), RemoveCountersFromThis(s82OneOilCounter, 1)),
			Targets:      s82OneYourArtifactOrCreature(),
			SorcerySpeed: true,
			Effect:       s82OnePutOilOnTarget,
		}},
	})
}
