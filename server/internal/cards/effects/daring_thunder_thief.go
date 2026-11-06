package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daring Thunder-Thief — Creature — Turtle Rogue {3}{U}, 4/4 (Outlaws of
// Thunder Junction, #4468):
//
//	"Flash
//	 This creature enters tapped."
//
// Flash is the printed keyword (CR 702.8a), which also lets the card be cast
// from hand at instant speed. "Enters tapped" is the shared self-entry
// replacement (CR 614.1d), so it applies on every way onto the battlefield,
// not only a cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0f5401c2-0189-4a07-8e7b-a765ccd7c954",
		Name:            "Daring Thunder-Thief",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements:    []game.ReplacementEffect{SelfEntersTapped()},
	})
}
