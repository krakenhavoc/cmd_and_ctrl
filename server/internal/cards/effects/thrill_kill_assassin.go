package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thrill-Kill Assassin — Creature — Human Assassin {1}{B}, 1/2 (Return to Ravnica, #5566):
//
//	"Deathtouch / Unleash (You may have this creature enter with a
//	 +1/+1 counter on it. It can't block as long as it has a +1/+1
//	 counter on it.)"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
// Unleash is the engine's keyword (ADR 0109 §10): the entering "may" and
// the can't-block while it has a +1/+1 counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f13f5705-e3e0-4e1e-a5b3-61e04e23a374",
		Name:            "Thrill-Kill Assassin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", game.KeywordUnleash},
	})
}
