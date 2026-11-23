package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "01e81249-6c45-4fae-86ee-9ef5fbec99ec",
		Name:         "Embraal Gear-Smasher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: This creature deals 2 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Effect:  effects.Do(effects.DamageEachOpponent{N: 2}),
		}},
	})
}
