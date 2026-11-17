package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Paladin en-Vec — Creature — Human Knight {1}{W}{W}, 2/2
// (Exodus, #5626):
//
//	"First strike, protection from black and from red"
//
// Three printed keywords. Protection is one quality token per colour
// (CR 702.16), as on Akroma, Angel of Fury: no black or red creature
// can block it, black and red spells and abilities can't target it, and
// damage from black or red sources is prevented.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd8722e1-9c41-4037-ab7e-47e2aa9858a0",
		Name:         "Paladin en-Vec",
		Completeness: CompletenessFull,
		PrintedKeywords: []string{
			"first strike",
			game.ProtectionFromColor("B"), game.ProtectionFromColor("R"),
		},
	})
}
