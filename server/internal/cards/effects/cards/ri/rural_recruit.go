package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rural Recruit — Creature — Human Peasant {3}{G}, 1/1 (Crimson Vow):
//
//	"Training (Whenever this creature attacks with another creature with
//	 greater power, put a +1/+1 counter on this creature.)
//	 When this creature enters, create a 3/1 green Boar creature token."
//
// Training is the shared Training() trigger (training.go, CR 702.149a); the
// Boar is a plain token row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fda06d9e-d17f-4bb6-916b-06116e47c8e8",
		Name:         "Rural Recruit",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.Training("Rural Recruit — training: put a +1/+1 counter on it"),
			effects.WhenThisEnters("Rural Recruit — create a 3/1 green Boar creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/1 green Boar"), N: 1})),
		},
	})
}
