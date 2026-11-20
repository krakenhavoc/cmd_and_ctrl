package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dead Reveler — Creature — Zombie {2}{B}, 2/3 (Return to Ravnica, #5566):
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
		OracleID:        "5cee3505-0291-4f1b-9e5f-2909a31acd8e",
		Name:            "Dead Reveler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordUnleash},
	})
}
