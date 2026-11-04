package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sting-Slinger — Creature — Goblin Warrior {2}{R}, 3/3:
//
//	"{1}{R}, {T}, Blight 1: This creature deals 2 damage to each
//	 opponent. (To blight 1, put a -1/-1 counter on a creature you
//	 control.)"
//
// The blight is part of the cost (BlightN, #4684): any creature the
// activator controls, named as they activate. The damage comes from the
// creature as it last existed if it has left by then (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "945a4a4d-fdfd-4c39-9975-5b22d17a4e70",
		Name:         "Sting-Slinger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}, {T}, Blight 1: This creature deals 2 damage to each opponent.",
			Cost:    Plus(ManaCost("{1}{R}"), TapCost(), BlightN(1)),
			Purpose: game.Purpose{Answers: game.AnswerValue, DamageEachOpponent: 2},
			Effect:  Do(DamageEachOpponent{N: 2}),
		}},
	})
}
