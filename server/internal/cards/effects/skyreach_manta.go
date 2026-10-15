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
		Caveats:             []string{"When “Charge mana costs” is off, or you choose Cast anyway, only mana spent from your mana pool counts, so it can enter with fewer counters than the colors you paid."},
		WantsDistinctColors: true,
		PrintedKeywords:     []string{game.KeywordSunburst, "flying"},
	})
}
