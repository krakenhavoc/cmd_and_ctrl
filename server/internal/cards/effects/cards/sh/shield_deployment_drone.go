package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// S.H.I.E.L.D. Deployment Drone — {2}{U} Artifact Creature — Robot, 2/2
// (Marvel Super Heroes, #4455):
//
//	"Flying
//	 When this creature enters, create a 1/1 white Soldier creature token."
//
// Black Panther, Vanguard's Soldier token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c1669020-8f81-471a-9083-7e9ce3e8f387",
		Name:            "S.H.I.E.L.D. Deployment Drone",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("S.H.I.E.L.D. Deployment Drone — create a 1/1 white Soldier", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1})),
		},
	})
}
