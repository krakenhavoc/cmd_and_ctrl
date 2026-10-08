package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Storm Fleet Swashbuckler — Creature — Human Pirate {1}{R}, 2/2:
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 This creature has double strike as long as you have the
//	 city's blessing."
//
// A layer 6 keyword grant, read live; the blessing is never lost once earned (CR 702.131c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6640bcfc-4bdc-4380-8b51-61b906f9aee6",
		Name:            "Storm Fleet Swashbuckler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		Static:          []game.StaticAbility{SelfKeywordWhileCitysBlessing("double strike")},
	})
}
