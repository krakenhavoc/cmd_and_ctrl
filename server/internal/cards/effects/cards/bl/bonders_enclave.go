package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bonders' Enclave — Land (EDHREC rank 544):
//
//	"{T}: Add {C}.
//	 {3}, {T}: Draw a card. Activate only if you control a creature
//	 with power 4 or greater."
//
// A colorless land that draws in a big-creature deck. The gate is the
// activation condition (CR 602.1b, #743), ControlsAtLeast(1, …) over
// the creature's CURRENT power — an anthem or +1/+1 counters count, as
// the printed "power 4 or greater" does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f33ce38a-34ec-4b65-a0fc-160484a02007",
		Name:         "Bonders' Enclave",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:     "{3}, {T}: Draw a card. Activate only if you control a creature with power 4 or greater.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Condition: effects.ControlsAtLeast(1, effects.MatchCreatureWithPowerAtLeast(4)),
			Effect:    effects.B36DrawOne,
		}},
	})
}
