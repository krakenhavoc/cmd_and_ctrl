package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chaos Imps — Creature — Imp {4}{R}{R}, 6/5:
//
//	"Flying
//	 Unleash (You may have this creature enter with a +1/+1 counter on
//	 it. It can't block as long as it has a +1/+1 counter on it.)
//	 This creature has trample as long as it has a +1/+1 counter on it."
//
// Unleash is the engine's keyword (ADR 0109 §10); the trample is a
// layer-6 self grant read off the live counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "31e99241-592d-4be4-9ef3-6dde522b1885",
		Name:            "Chaos Imps",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordUnleash},
		Static:          []game.StaticAbility{ThisHasWhileItHasCounter(game.CounterPlusOne, "trample")},
	})
}
