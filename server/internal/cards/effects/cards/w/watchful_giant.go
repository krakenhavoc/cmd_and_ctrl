package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Watchful Giant — Creature — Giant Soldier {5}{W}, 3/6 (Ravnica Allegiance,
// #5521):
//
//	"When this creature enters, create a 1/1 white Human creature token."
//
// Clarion Cathars's enters trigger and token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8239c63d-097a-449a-953a-cd480e594aa8",
		Name:         "Watchful Giant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Watchful Giant — create a 1/1 white Human creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: 1})),
		},
	})
}
