package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orazca Raptor — Creature — Dinosaur {2}{R}{R}, 3/4 (Rivals of Ixalan, #5531):
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
		OracleID:     "507ffc26-6c45-4167-9e0d-3c88b065720a",
		Name:         "Orazca Raptor",
		Completeness: effects.CompletenessFull,
	})
}
