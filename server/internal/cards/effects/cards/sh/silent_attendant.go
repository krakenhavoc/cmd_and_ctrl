package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silent Attendant — Creature — Human Cleric {2}{W}, 0/2 (Urza's Saga, S82 OLD
// batch 732, #5624):
//
//	"{T}: You gain 1 life."
//
// A plain activated ability (it is not a mana ability), so it uses the stack;
// the {T} is part of the cost, so it needs the Attendant to have been under
// its controller's control since the turn began (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4fa30e29-94c1-4537-bebd-dc7124e4ef71",
		Name:         "Silent Attendant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect:  effects.GainOneLifeEffect,
		}},
	})
}
