package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gigantic Big Bear — Creature — Bear {5}{G}{G}, 10/7 (The Hobbit):
//
//	"This spell can't be countered.
//	 Hexproof, haste"
//
// The uncounterable half is Spec.CantBeCountered (Thrun, the Last
// Troll's shape); hexproof and haste ride PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "88d7bd11-7b1d-426c-8552-01448531c6ad",
		Name:            "Gigantic Big Bear",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"hexproof", "haste"},
	})
}
