package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ironclad Krovod — Creature — Beast {3}{W}, 2/5 (War of the Spark, #5520):
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
		OracleID:     "bd1c9b5f-3b70-457c-8dd8-2420ce1c0d7c",
		Name:         "Ironclad Krovod",
		Completeness: effects.CompletenessFull,
	})
}
