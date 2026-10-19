package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gloom Stalker — Creature — Dwarf Ranger {2}{W}, 2/3:
//
//	"As long as you've completed a dungeon, this creature has double
//	 strike."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "16201410-86e9-40a7-88a6-c6c2fb2ea6bd",
		Name:         "Gloom Stalker",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordWhileCompletedADungeon(thisPermanent, "double strike"),
		},
	})
}
