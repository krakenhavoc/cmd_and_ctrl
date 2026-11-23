package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leatherback Baloth — Creature — Beast {G}{G}{G}, 4/5 (Worldwake, #5582):
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
		OracleID:     "0ee14128-3bec-4b65-8ee4-619337d4ed45",
		Name:         "Leatherback Baloth",
		Completeness: effects.CompletenessFull,
	})
}
