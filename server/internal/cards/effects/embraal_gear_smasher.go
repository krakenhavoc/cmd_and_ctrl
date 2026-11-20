package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Embraal Gear-Smasher — Creature — Human Warrior {2}{R}, 2/3 (Aether Revolt,
// #5536):
//
//	"{T}, Sacrifice an artifact: This creature deals 2 damage to each
//	 opponent."
//
// Intruding Soulrager's tap-and-sacrifice cost (an artifact, picked at
// announce) with Fire Nation Archers' DamageEachOpponent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "01e81249-6c45-4fae-86ee-9ef5fbec99ec",
		Name:         "Embraal Gear-Smasher",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: This creature deals 2 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), SacrificeN(1, "an artifact", Artifact())),
			Effect:  Do(DamageEachOpponent{N: 2}),
		}},
	})
}
