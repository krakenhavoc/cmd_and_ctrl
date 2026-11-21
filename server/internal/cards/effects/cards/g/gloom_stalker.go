package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gloom Stalker — Creature — Dwarf Ranger {2}{W}, 2/3:
//
//	"As long as you've completed a dungeon, this creature has double
//	 strike."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16201410-86e9-40a7-88a6-c6c2fb2ea6bd",
		Name:         "Gloom Stalker",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordWhileCompletedADungeon(effects.ThisPermanent, "double strike"),
		},
	})
}
