package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raptor Companion — Creature — Dinosaur {1}{W}, 3/1 (Ixalan, #5532):
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
		OracleID:     "20e65887-c68c-4dc0-b74b-9ce093f40b06",
		Name:         "Raptor Companion",
		Completeness: effects.CompletenessFull,
	})
}
