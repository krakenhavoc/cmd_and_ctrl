package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Great-Horn Krushok — Creature — Beast {4}{W}, 3/5 (Fate Reforged,
// #5552):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see grizzly_bears.go).
// The cost, types and body ride game.Card from Scryfall, so the Spec is the
// key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b193a6df-da22-4e3c-a086-5b02c604089e",
		Name:         "Great-Horn Krushok",
		Completeness: effects.CompletenessFull,
	})
}
