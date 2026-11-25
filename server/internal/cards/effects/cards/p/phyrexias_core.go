package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexia's Core — Land (Commander 2021):
//
//	"{T}: Add {C}.
//	 {1}, {T}, Sacrifice an artifact: You gain 1 life."
//
// A colourless land whose second ability sacrifices an artifact as a cost
// (CR 602.2b); the artifact is chosen as the ability is activated, and may
// be any artifact its controller has, tokens included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6cc062c-eb39-46ee-bd6d-17f1db0ac50d",
		Name:         "Phyrexia's Core",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice an artifact: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Effect:  effects.Do(effects.GainLife{Amount: 1}),
		}},
	})
}
