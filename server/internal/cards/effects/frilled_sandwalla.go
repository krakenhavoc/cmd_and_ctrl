package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frilled Sandwalla — Creature — Lizard {G}, 1/1 (Hour of Devastation, #5534):
//
//	"{1}{G}: This creature gets +2/+2 until end of turn. Activate only once
//	 each turn."
//
// Basking Rootwalla's row: the once-a-turn limit is OncePerTurnActivation
// keyed on the label, and the pump is the shared thisGetsUntilEndOfTurn.
//
// No simplification.
func init() {
	const label = "{1}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn."
	Register(Spec{
		OracleID:     "3315aa42-d3ae-425e-9c12-191d26e67a0a",
		Name:         "Frilled Sandwalla",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      ManaCost("{1}{G}"),
			Condition: OncePerTurnActivation(label),
			Effect:    thisGetsUntilEndOfTurn(2, 2, "Frilled Sandwalla — +2/+2 until end of turn"),
		}},
	})
}
