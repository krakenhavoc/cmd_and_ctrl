package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stone Haven Medic — Creature — Kor Cleric {1}{W}, 1/3 (Battle for
// Zendikar, #5547):
//
//	"{W}, {T}: You gain 1 life."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4d6ed72-f40f-4e4a-bec0-19b1d4f99ebe",
		Name:         "Stone Haven Medic",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: You gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Effect:  effects.GainOneLifeEffect,
		}},
	})
}
