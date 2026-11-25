package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gargoyle Castle — Land (Commander 2019, #5517):
//
//	"{T}: Add {C}.
//	 {5}, {T}, Sacrifice this land: Create a 3/4 colorless Gargoyle artifact
//	 creature token with flying."
//
// Spawning Bed's shape with one token. The cost is paid whole before
// anything resolves, so the Gargoyle arrives even though the land is gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "963f2848-15bd-441b-a55c-635f53b7b63f",
		Name:         "Gargoyle Castle",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}, Sacrifice this land: Create a 3/4 colorless Gargoyle artifact creature token with flying.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("3/4 colorless Gargoyle artifact with flying"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
