package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hoard Robber — Creature — Tiefling Rogue {1}{B}, 1/3 (Adventures in the
// Forgotten Realms):
//
//	"Whenever this creature deals combat damage to a player, create a
//	 Treasure token."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a04f70c-288d-4010-a6d5-24e9778b9b45",
		Name:         "Hoard Robber",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Hoard Robber — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
