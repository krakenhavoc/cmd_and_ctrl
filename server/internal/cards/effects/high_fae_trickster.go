package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// High Fae Trickster — Creature — Faerie Wizard {3}{U}, 4/2 (EDHREC
// rank 841):
//
//	"Flash
//	 Flying
//	 You may cast spells as though they had flash."
//
// Vedalken Orrery's sentence on a body: one Spec.CastTimings entry,
// read off the battlefield on every query, so it stops the moment the
// Trickster leaves or loses its abilities. It is about casting, so it
// does not let its controller play a land at instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "28c1a4a0-f58c-4d07-a918-51ff8a74f050",
		Name:            "High Fae Trickster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		CastTimings:     []game.CastTimingRule{CastAsThoughFlash()},
	})
}
