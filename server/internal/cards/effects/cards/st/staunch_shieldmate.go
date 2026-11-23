package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Staunch Shieldmate — Creature — Dwarf Soldier {W}, 1/3 (Core Set 2021, #5511):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (Silverback Ape's
// shape). The cost, types and body ride game.Card from Scryfall.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a1a0911-dcd5-4a2f-8deb-abb5b273b0fb",
		Name:         "Staunch Shieldmate",
		Completeness: effects.CompletenessFull,
	})
}
