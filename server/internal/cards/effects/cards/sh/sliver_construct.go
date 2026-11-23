package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sliver Construct — Artifact Creature — Sliver Construct {3}, 2/2 (Magic 2014, #5563):
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
		OracleID:     "7d592d5b-95cf-4416-9ade-1c23a8ba01d1",
		Name:         "Sliver Construct",
		Completeness: effects.CompletenessFull,
	})
}
