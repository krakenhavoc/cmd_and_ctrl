package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rotting Mastodon — Creature — Zombie Elephant {4}{B}, 2/8 (Khans of
// Tarkir, #5554):
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
		OracleID:     "d6e90e75-ef12-49d0-8dd8-0c6cb718ce28",
		Name:         "Rotting Mastodon",
		Completeness: effects.CompletenessFull,
	})
}
