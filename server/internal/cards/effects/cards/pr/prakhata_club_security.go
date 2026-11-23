package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prakhata Club Security — Creature — Aetherborn Warrior {3}{B}, 3/4 (Kaladesh, #5539):
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
		OracleID:     "9c5f130d-5972-449e-83db-41fb8cf76004",
		Name:         "Prakhata Club Security",
		Completeness: effects.CompletenessFull,
	})
}
