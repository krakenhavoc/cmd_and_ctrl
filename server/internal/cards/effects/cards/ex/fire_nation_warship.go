package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Nation Warship — Artifact — Vehicle {3}, 4/4 (Avatar: The Last
// Airbender, #4459):
//
//	"Reach
//	 When this Vehicle dies, create a Clue token.
//	 Crew 2"
//
// "Dies" is a creature put into a graveyard from the battlefield (CR
// 700.4), so the Clue comes only when the Warship was crewed as it left
// (avB18ThisVehicleDied). The Clue is created, not investigated (the
// ruling). Crew is the ordinary printed ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "987aab63-a889-400f-a3bd-feaef978b407",
		Name:            "Fire Nation Warship",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.AvB18ThisVehicleDied, "Fire Nation Warship — create a Clue token", effects.Do(effects.CreateToken{Template: effects.ClueToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Fire Nation Warship"),
		}},
	})
}
