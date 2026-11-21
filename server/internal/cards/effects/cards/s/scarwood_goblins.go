package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scarwood Goblins — Creature — Goblin {R}{G}, 2/2
// (The Dark, #5638):
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
		OracleID:     "5314e57b-107c-4478-9cdb-51d1732f9468",
		Name:         "Scarwood Goblins",
		Completeness: effects.CompletenessFull,
	})
}
