package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krovikan Scoundrel — Creature — Human Rogue {1}{B}, 2/1 (Coldsnap, #5597):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it. The cost, types and body
// ride game.Card from Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44c437ca-5d36-4bb3-9450-baf0c93ff4ce",
		Name:         "Krovikan Scoundrel",
		Completeness: effects.CompletenessFull,
	})
}
