package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Southern Elephant — Creature — Elephant {3}{G}, 3/4
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
		OracleID:     "79de2319-078d-46db-9044-0088c623b73f",
		Name:         "Southern Elephant",
		Completeness: effects.CompletenessFull,
	})
}
