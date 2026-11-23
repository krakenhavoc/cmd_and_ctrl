package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Straw Soldiers — Creature — Scarecrow Soldier {1}{U}, 1/3
// (Portal Three Kingdoms, #5622):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// ordinary_bear.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8138a9d7-63cb-4d40-ac9e-1a9b10766113",
		Name:         "Straw Soldiers",
		Completeness: effects.CompletenessFull,
	})
}
