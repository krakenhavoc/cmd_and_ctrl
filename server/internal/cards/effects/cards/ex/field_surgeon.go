package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Field Surgeon — Creature — Human Cleric {1}{W}, 1/1 (Urza's Destiny,
// #5621):
//
//	"Tap an untapped creature you control: Prevent the next 1 damage that
//	 would be dealt to target creature this turn."
//
// Opposition's cost (tapping an untapped creature you control) with
// Caregiver's one-damage shield (CR 615.7). Tapping a creature is not
// the {T} symbol (CR 302.6), so the Surgeon can pay with itself the
// turn it arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "730809d7-cb8c-495c-8972-7e85b967358e",
		Name:         "Field Surgeon",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped creature you control: Prevent the next 1 damage that would be dealt to target creature this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapAnUntapped("an untapped creature you control", effects.Creature()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.PreventNextOneDamageToTheTarget("Field Surgeon"),
		}},
	})
}
