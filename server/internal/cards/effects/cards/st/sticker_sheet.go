package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sticker sheet — Secret Lair (SLD), type line "Stickers":
//
//	(no mana cost, no rules text)
//
// A Secret Lair sticker sheet printed as a card with no characteristics beyond its
// name: no mana cost, no card type a game can use, no abilities. It is registered
// so the catalog counts it, and there is nothing for it to do: it has no
// abilities to resolve and nothing to cast. It never appears in a deck's playable
// pool.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f55a5b8a-ee73-4248-90c7-d2456928448e",
		Name:         "Sticker sheet",
		Completeness: effects.CompletenessFull,
	})
}
