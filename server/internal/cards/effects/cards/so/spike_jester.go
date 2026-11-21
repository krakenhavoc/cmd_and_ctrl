package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spike Jester — Creature — Goblin Warrior {B}{R}, 3/1:
//
//	"Haste"
//
// Haste lets it attack the turn it arrives (CR 702.10b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "29f22168-21aa-4ca9-ab72-b09396a12c76",
		Name:            "Spike Jester",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
