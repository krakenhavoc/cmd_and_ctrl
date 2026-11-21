package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Those Who Serve — Creature — Zombie {2}{W}, 2/4
// (Amonkhet, #5535):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ee088dc-c7c7-4f7f-9d1f-be3d79550d95",
		Name:         "Those Who Serve",
		Completeness: effects.CompletenessFull,
	})
}
