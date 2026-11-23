package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Queen's Bay Soldier — Creature — Vampire Soldier {1}{B}, 2/2 (Ixalan,
// #5532):
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
		OracleID:     "b288c863-5377-473d-a628-bb199993a29c",
		Name:         "Queen's Bay Soldier",
		Completeness: effects.CompletenessFull,
	})
}
