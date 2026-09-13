package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lyra Dawnbringer — Legendary Creature — Angel {3}{W}{W}, 5/5
// (EDHREC rank 1817):
//
//	"Flying
//	 First strike
//	 Lifelink
//	 Other Angels you control get +1/+1 and have lifelink."
//
// The Angel lord. Her own three keywords ride PrintedKeywords; the
// grant is Death Baron's shape over "other Angels you control" —
// a Layer 7c +1/+1 and a Layer 6 lifelink, effective subtypes so a
// changeling Angel counts. Lyra is excluded by ID, as printed.
//
// No simplification.
func init() {
	applies := b16OtherCreaturesYouControlOfSubtype("Angel")
	Register(Spec{
		OracleID:        "592c91fc-6430-4c76-9460-65f047350f67",
		Name:            "Lyra Dawnbringer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", "lifelink"},
		Static: []game.StaticAbility{
			b16Anthem(applies, 1, 1),
			b16GrantKeywords(applies, "lifelink"),
		},
	})
}
