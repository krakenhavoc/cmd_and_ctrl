package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dune Diviner — {2}{G} Creature — Snake Cleric 2/3 (Hour of Devastation):
//
//	"{1}, Tap an untapped Desert you control: You gain 1 life."
//
// Tapping a Desert is a cost that is not the {T} symbol (CR 302.6): a Desert
// that arrived this turn may pay it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4bf308fc-a54e-4fba-b0fd-3bedb4db3b18",
		Name:         "Dune Diviner",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Tap an untapped Desert you control: You gain 1 life.",
			Purpose: game.Purpose{LifeGain: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapAnUntapped("an untapped Desert you control", effects.Subtype("Desert"), effects.YouControl())),
			Effect:  effects.Do(effects.GainLife{Amount: 1}),
		}},
	})
}
