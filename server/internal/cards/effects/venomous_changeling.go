package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venomous Changeling — Creature — Shapeshifter {2}{B}, 1/3 (Modern Horizons, #5519):
//
//	"Changeling (This card is every creature type.)
//	 Deathtouch"
//
// Two printed keywords: changeling (every creature type, CR 702.73a) and
// deathtouch, which the combat and damage code read.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f1929f1f-6aa3-4308-bbff-cab03883ac68",
		Name:            "Venomous Changeling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling, "deathtouch"},
	})
}
