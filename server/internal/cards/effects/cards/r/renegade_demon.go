package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Renegade Demon — {3}{B}{B} Creature — Demon 5/3 (Avacyn Restored, #5570):
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
		OracleID:     "5f444efc-3bc2-4ca8-986c-f804f9e1a95e",
		Name:         "Renegade Demon",
		Completeness: effects.CompletenessFull,
	})
}
