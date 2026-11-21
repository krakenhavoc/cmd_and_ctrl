package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plundering Pirate — Creature — Orc Pirate {2}{R}, 3/2 (The Lost Caverns
// of Ixalan, #4470):
//
//	"When this creature enters, create a Treasure token. (It's an artifact
//	 with "{T}, Sacrifice this token: Add one mana of any color.")"
//
// Flamekin Gildweaver's entry trigger without trample: the shared
// TreasureToken template.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ab13412-1eaf-40c0-905d-8a3ca0c51be5",
		Name:         "Plundering Pirate",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Plundering Pirate — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
		},
	})
}
