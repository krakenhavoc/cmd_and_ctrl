package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Noble Templar — Creature — Human Cleric Soldier {5}{W}, 3/6 (Scourge):
//
//	"Vigilance
//	 Plainscycling {2} ({2}, Discard this card: Search your library for a
//	 Plains card, reveal it, put it into your hand, then shuffle.)"
//
// A printed keyword and the shared typecycling ability (CR 702.29e). The
// search finds any card with the Plains subtype, a nonbasic Plains
// included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e7079969-8d88-4f18-a872-2661bed6ff74",
		Name:            "Noble Templar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Plainscycling", "{2}", "a Plains card", effects.IsLandWithSubtype("Plains")),
		},
	})
}
