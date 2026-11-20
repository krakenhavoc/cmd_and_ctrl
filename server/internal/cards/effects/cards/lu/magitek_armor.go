package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magitek Armor — Artifact — Vehicle {3}{W}, 4/4 (Final Fantasy, #4462):
//
//	"When this Vehicle enters, create a 1/1 colorless Hero creature token.
//	 Crew 1"
//
// The Hero is the token job select makes, here on its own; it can crew
// the Armor straight away, since crewing taps it and needs no haste.
// Crew 1 is the ordinary CR 702.122 crew ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e9d8e12-b9da-4452-a758-5caffd8a752e",
		Name:         "Magitek Armor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Magitek Armor — create a 1/1 colorless Hero creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard(effects.JobSelectHeroToken), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Magitek Armor"),
		}},
	})
}
