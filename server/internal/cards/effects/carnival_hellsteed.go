package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carnival Hellsteed — Creature — Nightmare Horse {4}{B}{R}, 5/4 (Return to Ravnica, #5566):
//
//	"First strike, haste / Unleash (You may have this creature enter
//	 with a +1/+1 counter on it. It can't block as long as it has a
//	 +1/+1 counter on it.)"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
// Unleash is the engine's keyword (ADR 0109 §10): the entering "may" and
// the can't-block while it has a +1/+1 counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "34ae5b49-f8bf-4aa8-bb62-92d6d3118b26",
		Name:            "Carnival Hellsteed",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste", game.KeywordUnleash},
	})
}
