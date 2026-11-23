package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blade of the Sixth Pride — Creature — Cat Rebel {1}{W}, 3/1
// (Future Sight, #5594):
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
		OracleID:     "b7894010-fdba-498e-97de-22bd70332439",
		Name:         "Blade of the Sixth Pride",
		Completeness: effects.CompletenessFull,
	})
}
