package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plated Wurm — Creature — Wurm {4}{G}, 4/5 (Portal Second Age,
// #5625):
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
		OracleID:     "d7316cc6-a2c2-479c-80b5-53c86165e3d1",
		Name:         "Plated Wurm",
		Completeness: effects.CompletenessFull,
	})
}
