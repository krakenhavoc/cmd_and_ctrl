package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Penumbra Bobcat — Creature — Cat {2}{G}, 2/1 (Dominaria Remastered):
//
//	"When this creature dies, create a 2/1 black Cat creature token."
//
// A dies trigger that makes one token for its controller, last-known
// controller included (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0787397-8edb-4794-8ee1-22f97a057be8",
		Name:         "Penumbra Bobcat",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Penumbra Bobcat — create a 2/1 black Cat creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/1 black Cat"), N: 1})),
		},
	})
}
