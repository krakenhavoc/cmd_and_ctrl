package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Muraganda Raceway — Land:
//
//	"Start your engines!
//	 {T}: Add {C}.
//	 Max speed — {T}: Add {C}{C}."
//
// ADR 0138 (#2122). Two mana abilities; the second can be activated
// only while its controller has max speed (CR 702.178a), and the
// auto-tapper plans it only then (it reads the ability's Condition).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b5fa5651-d714-44d6-867b-be0e3224b7ed",
		Name:            "Muraganda Raceway",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		ManaAbilities: []effects.ManaAbility{
			{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "{T}: Add {C}."},
			effects.MaxSpeedMana(effects.ManaAbility{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{C}{C}", Label: "Max speed — {T}: Add {C}{C}."}),
		},
	})
}
