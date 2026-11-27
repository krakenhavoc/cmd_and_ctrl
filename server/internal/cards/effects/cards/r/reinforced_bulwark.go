package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reinforced Bulwark — {3} Artifact Creature — Wall 0/4 (Rise of the Eldrazi, S82 OLD #5581):
//
//	Defender
//	{T}: Prevent the next 1 damage that would be dealt to you this turn.
//
// Security Blockade's prevention shield on its controller.

func init() {
	effects.Register(effects.Spec{
		OracleID:        "8916b3d4-6a9a-420a-8ee7-be67b2f47201",
		Name:            "Reinforced Bulwark",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Prevent the next 1 damage that would be dealt to you this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PreventNextDamage{Target: item.Controller, Amount: 1, Label: "Reinforced Bulwark: prevent the next 1 damage"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
