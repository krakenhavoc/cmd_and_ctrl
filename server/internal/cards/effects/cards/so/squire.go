package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Squire — Creature — Human Soldier {1}{W}, 1/2
// (The Dark, #5638):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// ordinary_bear.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa6cdcc7-f5ea-47bf-9448-1c63e36b18d1",
		Name:         "Squire",
		Completeness: effects.CompletenessFull,
	})
}
