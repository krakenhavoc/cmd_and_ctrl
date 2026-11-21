package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pillarfield Ox — Creature — Ox {3}{W}, 2/4 (Zendikar, #5583):
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
		OracleID:     "0f7c70c4-9795-4c76-8b27-043942a963c6",
		Name:         "Pillarfield Ox",
		Completeness: effects.CompletenessFull,
	})
}
