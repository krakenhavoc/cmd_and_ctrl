package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rotting Fensnake — Creature — Zombie Snake {3}{B}, 5/1 (Innistrad, #5572):
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
		OracleID:     "be22faba-07e0-4036-b700-82769989250b",
		Name:         "Rotting Fensnake",
		Completeness: effects.CompletenessFull,
	})
}
