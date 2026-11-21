package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Garrison Cat — Creature — Cat {W}, 1/1 (Ikoria: Lair of Behemoths, #5512):
//
//	"When this creature dies, create a 1/1 white Human Soldier creature
//	 token."
//
// Doomed Traveler's death trigger with a Human Soldier.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7bc7c44-946c-4131-b764-add0c5f21020",
		Name:         "Garrison Cat",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Garrison Cat — create a 1/1 white Human Soldier creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1})),
		},
	})
}
