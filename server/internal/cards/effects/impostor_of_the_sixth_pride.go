package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Impostor of the Sixth Pride — Creature — Shapeshifter {1}{W}, 3/1 (Modern Horizons, #5519):
//
//	"Changeling (This card is every creature type.)"
//
// Changeling is the engine's keyword (CR 702.73a): the card is every
// creature type in every zone, so a tribal lord or a "choose a creature
// type" effect counts it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a2b8185a-9ff2-4d65-9bdd-b060b7ce69bb",
		Name:            "Impostor of the Sixth Pride",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
	})
}
