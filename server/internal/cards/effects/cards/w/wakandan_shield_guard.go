package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wakandan Shield Guard — Creature — Human Soldier {1}{W}, 1/1 (Marvel
// Super Heroes Commander, #4455): "When this creature enters, create a
// 1/1 white Soldier creature token." A one-token enters trigger from the
// catalog's token table.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9268e61-2697-46ff-b219-4f2be646cbe4",
		Name:         "Wakandan Shield Guard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wakandan Shield Guard — create a 1/1 white Soldier creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1})),
		},
	})
}
