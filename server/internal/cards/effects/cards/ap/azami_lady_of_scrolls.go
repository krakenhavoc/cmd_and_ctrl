package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Azami, Lady of Scrolls — Legendary Creature — Human Wizard {2}{U}{U}{U},
// 0/2 (Commander Masters):
//
//	"Tap an untapped Wizard you control: Draw a card."
//
// The cost is a tap of an untapped Wizard, not the {T} symbol (CR 302.6):
// a Wizard that arrived this turn, Azami herself included, may pay it. It
// does not target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f8b97fe-3e5e-47c2-9a9d-7f77482aa159",
		Name:         "Azami, Lady of Scrolls",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped Wizard you control: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapAnUntapped("an untapped Wizard you control", effects.HasSubtype("Wizard")),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
