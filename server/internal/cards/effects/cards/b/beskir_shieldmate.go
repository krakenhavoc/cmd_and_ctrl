package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beskir Shieldmate — Creature — Human Warrior {1}{W}, 2/1 (Kaldheim,
// #5505):
//
//	"When this creature dies, create a 1/1 white Human Warrior creature
//	 token."
//
// Martyr of Dusk's dies trigger with a Human Warrior token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1090a50-2347-48d0-bb67-379accb8e02d",
		Name:         "Beskir Shieldmate",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Beskir Shieldmate — create a 1/1 white Human Warrior creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Warrior"), N: 1})),
		},
	})
}
