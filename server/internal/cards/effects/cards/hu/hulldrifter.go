package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hulldrifter — Artifact — Vehicle {3}{U}{U}, 3/2 (Aetherdrift, #4464):
//
//	"Flying
//	 When this Vehicle enters, draw two cards.
//	 Crew 3 (Tap any number of creatures you control with total power 3
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)"
//
// Mulldrifter's enters draw on a Vehicle, with Broadcast Rambler's crew.
// The draw is the Vehicle's whether or not it is ever crewed; flying
// matters only while it is a creature (CR 702.9b), and crew is the
// engine's (CR 702.122a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a0a7319d-9441-4aed-a351-9ff3fb3b1d9e",
		Name:            "Hulldrifter",
		Purpose:         game.Purpose{Draws: 2},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hulldrifter — draw two cards", effects.Do(effects.DrawCards{N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    effects.CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Hulldrifter"),
		}},
	})
}
