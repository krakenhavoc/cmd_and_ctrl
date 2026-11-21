package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grizzled Outrider — Creature — Elf Warrior {4}{G}, 5/5 (Kaldheim, #5505):
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
		OracleID:     "0d3d0da2-73fe-4f8a-b13c-23981eb7e2ee",
		Name:         "Grizzled Outrider",
		Completeness: effects.CompletenessFull,
	})
}
