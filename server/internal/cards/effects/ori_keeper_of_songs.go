package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ori, Keeper of Songs — Legendary Creature — Dwarf Bard {2}{W}, 3/3:
//
//	"Storied (If you control three or more artifacts, legendaries,
//	 and/or Sagas, you have an enduring story for the rest of the game.)
//	 As long as you have an enduring story, Ori gets +1/+0 and has
//	 vigilance."
//
// Storied is the engine's (CR 702.195a, game/enduring_story.go); Ori
// counts as one of the three, being legendary. The rest reads the
// designation, which is never lost once had.
//
// No simplification. The Hobbit (#4475).
func init() {
	Register(Spec{
		OracleID:        "69cc089a-abd3-4edb-bc79-f02daac70837",
		Name:            "Ori, Keeper of Songs",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordStoried},
		Static: []game.StaticAbility{
			SelfPumpWhileEnduringStory(1, 0),
			SelfKeywordWhileEnduringStory("vigilance"),
		},
	})
}
