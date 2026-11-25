package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wydwen, the Biting Gale — {2}{U}{B} Legendary Creature — Faerie Wizard 3/3
// (Commander 2020):
//
//	"Flash
//	 Flying
//	 {U}{B}, Pay 1 life: Return Wydwen to its owner's hand."
//
// The life is a cost paid as the ability is activated (CR 118.4); the effect
// is Aethertide Whale's return of the source (a source that has already left
// is not returned).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7891ec87-8d1c-42fc-b00a-4c64ee5edbff",
		Name:            "Wydwen, the Biting Gale",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{B}, Pay 1 life: Return Wydwen to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{U}{B}"), effects.PayLife(1)),
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}
