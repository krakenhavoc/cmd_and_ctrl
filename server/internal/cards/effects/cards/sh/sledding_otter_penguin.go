package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sledding Otter-Penguin — Creature — Otter Bird {2}{W}, 2/3 (Avatar: The
// Last Airbender Eternal, #5075):
//
//	"{3}: Put a +1/+1 counter on this creature."
//
// Marketback Walker's ability at {3} (putCounterOnSelf).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b115051-a3c8-4c5f-9ad2-82b5a6897c51",
		Name:         "Sledding Otter-Penguin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}"),
			Effect:  effects.PutCounterOnSelf,
		}},
	})
}
