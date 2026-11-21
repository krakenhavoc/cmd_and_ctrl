package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bellows Lizard — Creature — Lizard {R}, 1/1 (Return to Ravnica, #5566):
//
//	"{1}{R}: This creature gets +1/+0 until end of turn."
//
// Firebreathing through the shared thisGetsUntilEndOfTurn body (Fathom
// Fleet Firebrand): no {T}, so it can be activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "667dae0b-5006-4e38-be80-eae79c2a2a95",
		Name:         "Bellows Lizard",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Bellows Lizard — +1/+0 until end of turn"),
		}},
	})
}
