package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Protector of Gondor — Creature — Human Soldier {3}{W}, 3/3 (The Lord of
// the Rings: Tales of Middle-earth Commander):
//
//	"When this creature enters, create a 1/1 white Human Soldier creature
//	 token."
//
// A plain enters token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf6bb391-4fde-43c4-a0c5-990630d0ddcf",
		Name:         "Protector of Gondor",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Protector of Gondor — create a 1/1 white Human Soldier creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1})),
		},
	})
}
