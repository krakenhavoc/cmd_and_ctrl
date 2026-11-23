package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krenko's Enforcer — Creature — Goblin Warrior {1}{R}{R}, 2/2 (Magic
// 2015, #5556):
//
//	"Intimidate (This creature can't be blocked except by artifact
//	 creatures and/or creatures that share a color with it.)"
//
// Bladetusk Boar's intimidate, an engine keyword the block check reads.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0d9787e3-f482-4a98-95af-0438ed09c96c",
		Name:            "Krenko's Enforcer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
	})
}
