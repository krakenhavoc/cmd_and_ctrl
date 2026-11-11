package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Midnight Banshee — Creature — Spirit {3}{B}{B}{B}, 5/5 (Lorwyn Eclipsed
// Commander, #5368):
//
//	"Wither
//	 At the beginning of your upkeep, put a -1/-1 counter on each
//	 nonblack creature."
//
// Wither is the engine's damage tail (game/infect_wither_toxic.go). The
// upkeep trigger is "your" upkeep only, and reads each creature's colour
// through the layered view (NonBlack).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dd8b6067-a97f-402f-bba1-70ebc6ee1599",
		Name:            "Midnight Banshee",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordWither},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Midnight Banshee — put a -1/-1 counter on each nonblack creature", cmdr2B25CountersOnEach(false, NonBlack())),
		},
	})
}
