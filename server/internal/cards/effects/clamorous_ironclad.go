package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clamorous Ironclad — Artifact — Vehicle {3}{R}, 6/3 (Aetherdrift,
// #4464):
//
//	"Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 Crew 3 (Tap any number of creatures you control with total power 3
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)
//	 Cycling {R} ({R}, Discard this card: Draw a card.)"
//
// Thundering Broodwagon without its enters trigger: menace is the
// Vehicle's whether or not it is crewed (CR 702.111b matters only once it
// is a creature), crew is the engine's (CR 702.122a), and cycling is the
// hand ability (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a08580da-e5aa-424b-86bf-d4535f3480d1",
		Name:            "Clamorous Ironclad",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []ActivatedAbility{
			{
				Label:   "Crew 3",
				Cost:    CrewCost(3),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Clamorous Ironclad"),
			},
			Cycling("{R}"),
		},
	})
}
