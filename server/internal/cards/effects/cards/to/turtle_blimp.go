package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Turtle Blimp — {5} Artifact — Vehicle 3/4 (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"Flying
//	 When this Vehicle enters, create a 2/2 red Mutant creature token.
//	 Crew 2"
//
// Slash's 2/2 red Mutant on an enters trigger, and the engine's crew
// (CR 702.122a) as on every Vehicle in the catalog. The Mutant can crew
// the Blimp the turn it is made: crewing taps, and summoning sickness
// does not stop a creature from paying a crew cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bb8c6dd2-1abe-4eb2-be1f-b0c078e5f257",
		Name:            "Turtle Blimp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Turtle Blimp — create a 2/2 red Mutant creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 red Mutant"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Turtle Blimp"),
		}},
	})
}
