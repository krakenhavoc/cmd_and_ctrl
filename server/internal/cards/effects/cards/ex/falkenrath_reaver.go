package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Falkenrath Reaver — Creature — Vampire {1}{R}, 2/2 (Eldritch Moon, #5541):
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
		OracleID:     "6cfd9d88-f450-4408-bbe2-94a2d295b24a",
		Name:         "Falkenrath Reaver",
		Completeness: effects.CompletenessFull,
	})
}
