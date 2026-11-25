package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enchanted Carriage — Artifact — Vehicle {5}, 4/4 (Throne of Eldraine, S82 old
// sets):
//
//	"When this Vehicle enters, create two 1/1 white Mouse creature tokens.
//	 Crew 2"
//
// The enters trigger makes both tokens as one creation (CR 111.2). Crew 2 is the
// shared CrewCost / CrewEffect, as on Passenger Ferry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f4a3c9e-80c4-4fa1-940d-2c57065901a3",
		Name:         "Enchanted Carriage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Enchanted Carriage — create two 1/1 white Mouse creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Mouse"), N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Enchanted Carriage"),
		}},
	})
}
