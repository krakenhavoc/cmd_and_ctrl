package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// S.H.I.E.L.D. Helicarrier — {4} Artifact — Vehicle, 4/5 (Marvel Super
// Heroes, #4455):
//
//	"Flying
//	 When this Vehicle enters, create two 1/1 white Soldier creature tokens.
//	 Crew 6"
//
// The two tokens are one CreateToken with N: 2, so they enter together.
// Crew is Aethersphere Harvester's row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0f2335bf-6a13-4483-997a-fd19d30c4d48",
		Name:            "S.H.I.E.L.D. Helicarrier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("S.H.I.E.L.D. Helicarrier — create two 1/1 white Soldiers", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 2})),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Crew 6",
				Cost:    effects.CrewCost(6),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  effects.CrewEffect("S.H.I.E.L.D. Helicarrier"),
			},
		},
	})
}
