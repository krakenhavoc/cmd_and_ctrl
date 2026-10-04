package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sightless Ghoul — Creature — Zombie Soldier {3}{B}, 2/2:
//
//	"This creature can't block.
//	 Undying"
//
// Undying is PrintedKeywords (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8791adf7-16b5-4692-9eaa-0d0c58867d9d",
		Name:            "Sightless Ghoul",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUndying},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
}
