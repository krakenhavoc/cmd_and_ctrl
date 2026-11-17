package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skyshroud Behemoth — Creature — Beast {5}{G}{G}, 10/10:
//
//	"Fading 2 (This creature enters with two fade counters on it. At the
//	 beginning of your upkeep, remove a fade counter from it. If you
//	 can't, sacrifice it.)
//	 This creature enters tapped."
//
// The fade counters and "enters tapped" are one replacement
// (FadingEntersTapped, fading.go), so its controller is not asked to
// order two self-replacements that come out the same either way
// (CR 616.1). A copy of it enters tapped with two fade counters too,
// the example under CR 707.5.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77bee041-6f96-44e3-a18c-b52e508b9bdf",
		Name:         "Skyshroud Behemoth",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{FadingEntersTapped(2)},
		Triggered:    []game.TriggeredAbility{FadingUpkeep()},
	})
}
