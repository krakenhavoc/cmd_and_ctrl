package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archivist — Creature — Human Wizard {2}{U}{U}, 1/1 (Urza's Legacy,
// #5623):
//
//	"{T}: Draw a card."
//
// A tapping card-drawer. The {T} in the cost means it can't be activated the
// turn Archivist arrives (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d137586f-83b0-40af-8100-443460b07ac0",
		Name:         "Archivist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
