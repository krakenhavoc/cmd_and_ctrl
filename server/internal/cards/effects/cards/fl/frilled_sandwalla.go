package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "3315aa42-d3ae-425e-9c12-191d26e67a0a",
		Name:         "Frilled Sandwalla",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{1}{G}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect:    effects.ThisGetsUntilEndOfTurn(2, 2, "Frilled Sandwalla — +2/+2 until end of turn"),
		}},
	})
}
