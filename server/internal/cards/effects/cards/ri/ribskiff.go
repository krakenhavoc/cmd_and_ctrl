package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "8f761657-7127-4552-a200-532e063e2419",
		Name:            "Ribskiff",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"toxic 2"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ribskiff — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    effects.CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Ribskiff"),
		}},
	})
}
