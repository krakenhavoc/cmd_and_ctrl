package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kalonian Tusker — Creature — Beast {G}{G}, 3/3 (Magic 2014, #5563):
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
		OracleID:     "df7f697e-6886-4897-a024-61ae225c1b34",
		Name:         "Kalonian Tusker",
		Completeness: effects.CompletenessFull,
	})
}
