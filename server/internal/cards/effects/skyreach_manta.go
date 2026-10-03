package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skyreach Manta — Artifact Creature — Fish {5}, 0/0:
//
//	"Sunburst (This creature enters with a +1/+1 counter on it for each
//	 color of mana spent to cast it.)
//	 Flying"
//
// Sunburst is a keyword the engine reads off the resolving spell (ADR
// 0109 §11, #1552); flying is printed.
func init() {
	Register(Spec{
		OracleID:            "58eab745-9afa-4db7-9d8a-f5eb0b9dca6c",
		Name:                "Skyreach Manta",
		Completeness:        CompletenessCaveats,
		Caveats:             []string{"With strict mana off, the game doesn't track which mana you spent, so it enters with no counters at all."},
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst, "flying"},
	})
}
