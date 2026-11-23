package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Catacomb Crocodile — Creature — Crocodile {4}{B}, 3/7 (Ravnica Allegiance,
// #5521):
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
		OracleID:     "edc52483-19c7-4d3a-8915-6b1f41b2d6e5",
		Name:         "Catacomb Crocodile",
		Completeness: effects.CompletenessFull,
	})
}
