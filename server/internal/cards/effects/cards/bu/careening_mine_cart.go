package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Careening Mine Cart — Artifact — Vehicle {3}, 3/3 (The Lost Caverns of
// Ixalan, #4470):
//
//	"Whenever this Vehicle attacks, create a Treasure token.
//	 Crew 1"
//
// Rocketeer Boostbuggy's attack trigger, which makes the catalog's
// Treasure, and the shared crew row (CR 702.122): tapping creatures with
// total power 1 or more makes the cart an artifact creature until end of
// turn, so it can attack and trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3775748-ff62-469f-916b-8e551ccce23c",
		Name:         "Careening Mine Cart",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Careening Mine Cart — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Careening Mine Cart"),
		}},
	})
}
