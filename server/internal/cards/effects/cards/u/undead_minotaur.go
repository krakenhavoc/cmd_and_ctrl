package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undead Minotaur — Creature — Zombie Minotaur {2}{B}, 2/3 (Magic 2014, #5563):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see ordinary_bear.go).
// The cost, types and body ride game.Card from Scryfall, so the Spec is the key
// and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19b25d09-563f-4159-9ce0-c0343564bd3b",
		Name:         "Undead Minotaur",
		Completeness: effects.CompletenessFull,
	})
}
