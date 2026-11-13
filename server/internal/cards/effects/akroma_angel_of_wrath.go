package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Akroma, Angel of Wrath — Legendary Creature — Angel {5}{W}{W}{W}, 6/6
// (Legions, #5609):
//
//	"Flying, first strike, vigilance, trample, haste, protection from
//	 black and from red"
//
// Seven printed keywords. Protection is one quality token per colour
// (CR 702.16), as on Paladin en-Vec: black and red spells and abilities
// can't target it, no black or red creature can block it, and damage
// from black or red sources is prevented.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "107f204c-9fd3-490d-8887-be63d14fc6a4",
		Name:         "Akroma, Angel of Wrath",
		Completeness: CompletenessFull,
		PrintedKeywords: []string{
			"flying", "first strike", "vigilance", "trample", "haste",
			game.ProtectionFromColor("B"), game.ProtectionFromColor("R"),
		},
	})
}
