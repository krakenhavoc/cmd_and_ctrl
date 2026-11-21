package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voice of the Woods — Creature — Elf {3}{G}{G}, 2/2 (Foundations
// Jumpstart, S82, #5929):
//
//	"Tap five untapped Elves you control: Create a 7/7 green Elemental
//	 creature token with trample."
//
// Ancestor's Prophet's tap-five cost (tapFiveUntapped) over Elves. The
// Voice may tap itself toward the five when it is an Elf, since the cost
// is not the {T} symbol and summoning sickness does not stop it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "12da8ba3-ca60-44c8-8627-1b4af2a2c4a3",
		Name:         "Voice of the Woods",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap five untapped Elves you control: Create a 7/7 green Elemental creature token with trample.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapFiveUntapped("five untapped Elves you control", "Elf"),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("7/7 green Elemental with trample"), N: 1}),
		}},
	})
}
