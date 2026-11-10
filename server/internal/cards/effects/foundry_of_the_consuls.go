package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0498ea14-53d9-4655-a616-0ff3cf73de4e",
		Name:         "Foundry of the Consuls",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}, Sacrifice this land: Create two 1/1 colorless Thopter artifact creature tokens with flying.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{5}"), TapCost(), SacrificeThis()),
			Effect:  Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 2}),
		}},
	})
}
