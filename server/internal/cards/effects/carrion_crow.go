package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carrion Crow — Creature — Zombie Bird {2}{B}, 2/2 (Magic 2015, #5556):
//
//	"Flying
//	 This creature enters tapped."
//
// Piranha Fly's two lines: printed flying and the SelfEntersTapped
// replacement every tapland uses, so it is never untapped on the
// battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2f9de785-a80a-4356-a4b4-0e0243e8ea2d",
		Name:            "Carrion Crow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{SelfEntersTapped()},
	})
}
