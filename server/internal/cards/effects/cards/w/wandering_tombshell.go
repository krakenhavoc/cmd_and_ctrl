package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wandering Tombshell — Creature — Zombie Turtle {3}{B}, 1/6 (Dragons of Tarkir,
// #5551):
//
//	"(no rules text)"
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "768c78f3-f09c-4b77-974d-514925e8df5e",
		Name:         "Wandering Tombshell",
		Completeness: effects.CompletenessFull,
	})
}
