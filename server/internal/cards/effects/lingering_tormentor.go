package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lingering Tormentor — Creature — Spirit {3}{B}, 2/2 (Eventide, #5590):
//
//	"Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)
//	 Persist (When this creature dies, if it had no -1/-1 counters on it,
//	 return it to the battlefield under its owner's control with a -1/-1
//	 counter on it.)"
//
// Fear (CR 702.36) and persist (CR 702.79), Razortooth Rats' and Persistent
// Constrictor's keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "276b1bd8-8918-4226-bab3-100cf5aa3c24",
		Name:            "Lingering Tormentor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear", game.KeywordPersist},
	})
}
