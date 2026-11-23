package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orazca Frillback — Creature — Dinosaur {2}{G}, 4/2 (Rivals of Ixalan, #5531):
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
		OracleID:     "2da32ec3-f42f-4edc-b5b3-e5f39fcf370c",
		Name:         "Orazca Frillback",
		Completeness: effects.CompletenessFull,
	})
}
