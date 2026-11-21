package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Axebane Beast — Creature — Beast {3}{G}, 3/4 (Ravnica Allegiance, #5521):
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
		OracleID:     "8f1d5c8d-98e1-47b7-8347-38887253e34a",
		Name:         "Axebane Beast",
		Completeness: effects.CompletenessFull,
	})
}
