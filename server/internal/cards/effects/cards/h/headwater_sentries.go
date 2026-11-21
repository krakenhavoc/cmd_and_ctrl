package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Headwater Sentries — Creature — Merfolk Warrior {3}{U}, 2/5 (Ixalan,
// #5532):
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
		OracleID:     "17aa790d-d16a-4a90-b8ba-cd50e62f02f0",
		Name:         "Headwater Sentries",
		Completeness: effects.CompletenessFull,
	})
}
