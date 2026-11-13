package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faithful Watchdog — Creature — Dog {G}{W}, 0/0 (Modern Horizons 3,
// #5467):
//
//	"Vigilance
//	 This creature enters with three +1/+1 counters on it."
//
// Afiya Grove's entry counters on a creature: the counters are part of
// the event that puts it onto the battlefield (CR 614.1c), so the 0/0
// never sees a state-based check without them. Vigilance is the engine's
// keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "139fb542-89fd-4b9e-87e6-ec925bcca93a",
		Name:            "Faithful Watchdog",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 3, "Faithful Watchdog: enters with three +1/+1 counters"),
		},
	})
}
