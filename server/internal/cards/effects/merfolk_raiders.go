package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merfolk Raiders — Creature — Merfolk Soldier {1}{U}, 2/3 (Mirage,
// #5633):
//
//	"Islandwalk (This creature can't be blocked as long as defending
//	 player controls an Island.)
//	 Phasing (This phases in or out before you untap during each of your
//	 untap steps. While it's phased out, it's treated as though it doesn't
//	 exist.)"
//
// Both are printed keywords the engine honours on its own: islandwalk in
// the block legality check (CR 702.14c), phasing in its controller's
// untap step (CR 702.26a, game/phasing.go, ADR 0084), as on Teferi's
// Drake.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "19edc98c-edc2-47bf-b2da-00f536d7c158",
		Name:            "Merfolk Raiders",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"islandwalk", game.KeywordPhasing},
	})
}
