package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gore-House Chainwalker — Creature — Human Warrior {1}{R}, 2/1 (Return to Ravnica, #5566):
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
		OracleID:        "dbe3e317-3c39-4b38-ae06-eae02a910c37",
		Name:            "Gore-House Chainwalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUnleash},
	})
}
