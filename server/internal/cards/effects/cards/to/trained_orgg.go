package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trained Orgg — Creature — Orgg {6}{R}, 6/6
// (Starter 1999, #5620):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (Grizzly Bears'
// shape). The cost, types and body ride game.Card from Scryfall, so the
// Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da67de0f-6b0d-4100-9260-334f8c235c59",
		Name:         "Trained Orgg",
		Completeness: effects.CompletenessFull,
	})
}
