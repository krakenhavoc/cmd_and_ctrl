package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voice of Duty — Creature — Angel {3}{W}, 2/2 (Urza's Destiny, #5621):
//
//	"Flying, protection from green"
//
// Two printed keywords. Protection is one quality token per colour
// (CR 702.16), as on Paladin en-Vec: no green creature can block it,
// green spells and abilities can't target it, and damage from green
// sources is prevented.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b5d5b9c9-accc-4dbe-aaf4-c806f3e1c61c",
		Name:            "Voice of Duty",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.ProtectionFromColor("G")},
	})
}
