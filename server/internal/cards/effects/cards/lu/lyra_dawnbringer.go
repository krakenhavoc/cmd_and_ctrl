package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	applies := effects.B16OtherCreaturesYouControlOfSubtype("Angel")
	effects.Register(effects.Spec{
		OracleID:        "592c91fc-6430-4c76-9460-65f047350f67",
		Name:            "Lyra Dawnbringer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", "lifelink"},
		Static: []game.StaticAbility{
			effects.B16Anthem(applies, 1, 1),
			effects.B16GrantKeywords(applies, "lifelink"),
		},
	})
}
