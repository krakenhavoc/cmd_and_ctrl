package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feral Krushok — Creature — Beast {4}{G}, 5/4 (Fate Reforged,
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
		OracleID:     "8c42883d-8b70-4691-a1e7-4659e7954b88",
		Name:         "Feral Krushok",
		Completeness: effects.CompletenessFull,
	})
}
