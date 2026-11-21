package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alpine Grizzly — Creature — Bear {2}{G}, 4/2 (Khans of Tarkir, #5554):
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
		OracleID:     "86e3c86c-d50e-4aa9-b080-f68b6afee87c",
		Name:         "Alpine Grizzly",
		Completeness: effects.CompletenessFull,
	})
}
