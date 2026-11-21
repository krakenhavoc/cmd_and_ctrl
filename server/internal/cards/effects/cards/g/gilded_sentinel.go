package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gilded Sentinel — Artifact Creature — Golem {4}, 3/3 (Ixalan, #5532):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a382165d-d715-4836-b1ce-9e5d8dc56fb4",
		Name:         "Gilded Sentinel",
		Completeness: effects.CompletenessFull,
	})
}
