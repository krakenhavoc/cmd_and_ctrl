package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ageless Guardian — Creature — Spirit Soldier {1}{W}, 1/4 (Strixhaven:
// School of Mages, #5502):
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
		OracleID:     "468658bc-19d8-464a-ab8e-d36281680273",
		Name:         "Ageless Guardian",
		Completeness: effects.CompletenessFull,
	})
}
