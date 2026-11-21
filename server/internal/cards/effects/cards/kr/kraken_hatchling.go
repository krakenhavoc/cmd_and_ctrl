package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kraken Hatchling — Creature — Kraken {U}, 0/4 (Zendikar, #5583):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// ordinary_bear.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08e6c786-d553-4c23-9137-cfad84146739",
		Name:         "Kraken Hatchling",
		Completeness: effects.CompletenessFull,
	})
}
