package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gore Swine — Creature — Boar {2}{R}, 4/1 (Fate Reforged,
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
		OracleID:     "af4b5853-ae60-48c9-9be1-bf121fa2e03d",
		Name:         "Gore Swine",
		Completeness: effects.CompletenessFull,
	})
}
