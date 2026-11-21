package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smoldering Butcher — Creature — Elemental Warrior {3}{B}, 4/2 (Eventide, #5590):
//
//	"Wither (This deals damage to creatures in the form of -1/-1 counters.)"
//
// The printed keyword (CR 702.80).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b0f56eef-1abc-437e-8e2e-c079731fe783",
		Name:            "Smoldering Butcher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordWither},
	})
}
