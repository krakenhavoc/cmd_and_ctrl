package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "987aab63-a889-400f-a3bd-feaef978b407",
		Name:            "Fire Nation Warship",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, avB18ThisVehicleDied, "Fire Nation Warship — create a Clue token", Do(CreateToken{Template: ClueToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Fire Nation Warship"),
		}},
	})
}
