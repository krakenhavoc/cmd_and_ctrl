package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foundry of the Consuls — Land:
//
//	"{T}: Add {C}.
//	 {5}, {T}, Sacrifice this land: Create two 1/1 colorless Thopter
//	 artifact creature tokens with flying."
//
// Treasure Vault's sacrifice-a-land shape making two Thopters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0498ea14-53d9-4655-a616-0ff3cf73de4e",
		Name:         "Foundry of the Consuls",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}, Sacrifice this land: Create two 1/1 colorless Thopter artifact creature tokens with flying.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 2}),
		}},
	})
}
