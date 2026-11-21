package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Looming Altisaur — Creature — Dinosaur {3}{W}, 1/7 (Ixalan, #5532):
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
		OracleID:     "af5302d3-782b-4d92-bc97-dccbfc76ac95",
		Name:         "Looming Altisaur",
		Completeness: effects.CompletenessFull,
	})
}
