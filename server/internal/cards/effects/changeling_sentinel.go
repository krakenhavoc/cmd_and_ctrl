package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Changeling Sentinel — Creature — Shapeshifter {3}{W}, 3/2 (Morningtide,
// #5592):
//
//	"Changeling (This card is every creature type.)
//	 Vigilance"
//
// Two printed keywords: changeling is every creature type in every zone
// (CR 702.73a), and vigilance means attacking doesn't tap it (CR 702.20b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2977aae2-225d-4f3f-8bf7-535f542a7391",
		Name:            "Changeling Sentinel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling, "vigilance"},
	})
}
