package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Golgari Longlegs — Creature — Insect {3}{B/G}{B/G}, 5/4 (Return to Ravnica, #5566):
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
		OracleID:     "f1534fbf-eab6-4f12-8037-e71d6861b636",
		Name:         "Golgari Longlegs",
		Completeness: effects.CompletenessFull,
	})
}
