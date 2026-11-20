package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Duskmantle Prowler — Creature — Vampire Rogue {3}{B}, 2/2 (Magic 2013, #5568):
//
//	"Haste (This creature can attack and {T} as soon as it comes under your
//	 control.)
//	 Exalted (Whenever a creature you control attacks alone, that
//	 creature gets +1/+1 until end of turn.)"
//
// Haste (CR 702.10) and Akrasan Squire's exalted (CR 702.83a), both
// printed keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9784ccf1-2238-417a-b6d1-cfc2e3dcae5f",
		Name:            "Duskmantle Prowler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste", game.KeywordExalted},
	})
}
