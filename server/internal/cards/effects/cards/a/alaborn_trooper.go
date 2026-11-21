package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alaborn Trooper — Creature — Human Soldier {2}{W}, 2/3 (Portal Second Age,
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
		OracleID:     "17969b71-9612-45ff-b789-1c033f832154",
		Name:         "Alaborn Trooper",
		Completeness: effects.CompletenessFull,
	})
}
