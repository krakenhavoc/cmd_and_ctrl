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
// Sandbox simplification, declared — the Urabrask / Kismet gap: an
// opponent's creature TOKEN enters untapped. Token creation does not
// go through the CR 614 zone-move pipeline at all (a token has no
// previous zone to move from), so no enters-tapped replacement in
// the catalog sees one. Weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:        "b9243163-b726-432e-830f-86132aa7f34a",
		Name:            "Kinjalli's Sunwing",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"An opponent's creature token enters untapped — only cast, reanimated or returned creatures enter tapped."},
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			b31OpponentsCreaturesEnterTapped("Kinjalli's Sunwing: enters tapped"),
		},
	})
}
