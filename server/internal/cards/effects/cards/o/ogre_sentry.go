package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ogre Sentry — Creature — Ogre Warrior {1}{R}, 3/3 (Conspiracy: Take the
// Crown, S82 older sets batch 272, #5540):
//
//	"Defender"
//
// Defender is the engine's keyword (CR 702.3): it can't attack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "263b3472-b437-4889-af5e-c7537abb6cb6",
		Name:            "Ogre Sentry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
