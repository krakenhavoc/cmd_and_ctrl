package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whirler Virtuoso — Creature — Vedalken Artificer {1}{U}{R}, 2/3:
//
//	"When this creature enters, you get {E}{E}{E} (three energy
//	 counters).
//	 Pay {E}{E}{E}: Create a 1/1 colorless Thopter artifact creature
//	 token with flying."
//
// ADR 0129 PR 1 (#1995).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f82da236-e567-4221-a0bc-439a0c7e2b03",
		Name:         "Whirler Virtuoso",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Whirler Virtuoso", 3),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay {E}{E}{E}: Create a 1/1 colorless Thopter artifact creature token with flying.",
			Cost:    effects.PayEnergy(3),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1}),
		}},
	})
}
