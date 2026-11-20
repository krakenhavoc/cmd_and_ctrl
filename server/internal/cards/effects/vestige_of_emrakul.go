package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vestige of Emrakul — Creature — Eldrazi Drone {3}{R}, 3/4 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Trample"
//
// Devoid is the engine's colour-removing keyword (CR 702.114a), so the
// card is colourless in every zone; trample is carried as printed data
// (CR 702.19b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "55187899-c445-48e9-80b9-235eaf42f803",
		Name:            "Vestige of Emrakul",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "trample"},
	})
}
