package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spawn of Rix Maadi — Creature — Horror {3}{B}{R}, 5/3 (Return to Ravnica, #5566):
//
//	"Unleash (You may have this creature enter with a +1/+1 counter on
//	 it. It can't block as long as it has a +1/+1 counter on it.)"
//
// Unleash is the engine's keyword (ADR 0109 §10): the entering "may" and
// the can't-block while it has a +1/+1 counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "39d4373b-35b4-4ceb-b4d5-9ca474365eb0",
		Name:            "Spawn of Rix Maadi",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUnleash},
	})
}
