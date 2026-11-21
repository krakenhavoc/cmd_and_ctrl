package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Gargantua — Creature — Phyrexian Horror {4}{B}{B}, 4/4:
//
//	"When this creature enters, you draw two cards and you lose 2 life."
//
// One ability: the draw, then the life loss (Mephitic Draught's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3925dfd-62b0-44ef-8545-7bee92cad5a2",
		Name:         "Phyrexian Gargantua",
		Purpose:      game.Purpose{Draws: 2},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Phyrexian Gargantua — you draw two cards and you lose 2 life",
				effects.Do(effects.DrawCards{N: 2}, effects.YouLoseLife(2))),
		},
	})
}
