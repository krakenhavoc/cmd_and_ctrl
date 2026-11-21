package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tocasia's Dig Site — Land:
//
//	"{T}: Add {C}.
//	 {3}, {T}: Surveil 1."
//
// Broodheart Engine's plain surveil as an activated ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "a91f93fd-e428-4a36-b1b3-604b47a34287",
		Name:          "Tocasia's Dig Site",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.PainlessColorless()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Surveil{N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
