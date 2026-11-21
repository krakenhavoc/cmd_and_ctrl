package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Minotaur Abomination — Creature — Zombie Minotaur {4}{B}{B}, 4/6 (Magic 2014, #5563):
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
		OracleID:     "017568ea-613f-4057-9a8d-61b087d6dc92",
		Name:         "Minotaur Abomination",
		Completeness: effects.CompletenessFull,
	})
}
