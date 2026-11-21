package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tasseled Dromedary — Creature — Camel {W}, 0/4 (Kaladesh, #5539):
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
		OracleID:     "3121eecd-5f8a-408f-8b2d-75e13aa8a1ea",
		Name:         "Tasseled Dromedary",
		Completeness: effects.CompletenessFull,
	})
}
