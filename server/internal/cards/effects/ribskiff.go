package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ribskiff — Artifact — Vehicle {4}, 4/4 (Phyrexia: All Will Be One, #5479):
//
//	"Toxic 2
//	 When this Vehicle enters, draw a card.
//	 Crew 3"
//
// Toxic is printed on the Vehicle and matters once it is crewed and deals
// combat damage to a player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8f761657-7127-4552-a200-532e063e2419",
		Name:            "Ribskiff",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"toxic 2"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ribskiff — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Ribskiff"),
		}},
	})
}
