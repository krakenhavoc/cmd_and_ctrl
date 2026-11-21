package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rakdos Drake — Creature — Drake {2}{B}, 1/2:
//
//	"Flying
//	 Unleash (You may have this creature enter with a +1/+1 counter on
//	 it. It can't block as long as it has a +1/+1 counter on it.)"
//
// Flying plus unleash, the engine's keyword (CR 702.98): the "enter with a
// +1/+1 counter" choice is offered as it enters, and a Drake with the
// counter can't block. Hellhole Flailer's shape.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0715167b-00b1-45a4-94a0-60d4097a1428",
		Name:            "Rakdos Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordUnleash},
	})
}
