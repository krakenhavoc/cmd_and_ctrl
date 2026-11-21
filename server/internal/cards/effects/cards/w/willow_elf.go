package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Willow Elf — Creature — Elf {G}, 1/1
// (Starter 1999, #5620):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (Grizzly Bears'
// shape). The cost, types and body ride game.Card from Scryfall, so the
// Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f88db161-3d00-4f45-b2a4-cbdc4f8e5312",
		Name:         "Willow Elf",
		Completeness: effects.CompletenessFull,
	})
}
