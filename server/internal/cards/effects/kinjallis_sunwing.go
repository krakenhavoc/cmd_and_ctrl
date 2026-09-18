package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kinjalli's Sunwing — Creature — Dinosaur {2}{W}, 2/3 (EDHREC rank
// 3290):
//
//	"Flying
//	 Creatures your opponents control enter tapped."
//
// The Dinosaur with Urabrask's second line. Flying rides
// PrintedKeywords; the tapped entry is the CR 614 replacement on
// the entering creature's move, applied when its controller is not
// the Sunwing's (b31OpponentsCreaturesEnterTapped). A cast,
// reanimated, fetched or flickered creature runs the entry pipeline
// and arrives tapped, as printed.
//
// An opponent's creature TOKEN enters tapped too, since #762: a
// created token now runs the same battlefield-entry pipeline every
// other permanent runs, so this replacement sees it exactly as it
// sees a cast creature.
func init() {
	Register(Spec{
		OracleID:        "b9243163-b726-432e-830f-86132aa7f34a",
		Name:            "Kinjalli's Sunwing",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			b31OpponentsCreaturesEnterTapped("Kinjalli's Sunwing: enters tapped"),
		},
	})
}
