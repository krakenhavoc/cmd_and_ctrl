package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skybox Ferry — Artifact — Vehicle {5}, 4/4 (Aetherdrift, #4464):
//
//	"Flying
//	 Crew 2 (Tap any number of creatures you control with total power 2
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Air Response Unit's crew with a printed keyword, plus the engine's
// cycling from hand (#660). Crew works on the battlefield, cycling in
// the hand, so the two rows never compete.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "09726c3d-f700-4bf4-a9e7-326fe41b0a4a",
		Name:            "Skybox Ferry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Crew 2",
				Cost:    effects.CrewCost(2),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  effects.CrewEffect("Skybox Ferry"),
			},
			effects.Cycling("{2}"),
		},
	})
}
