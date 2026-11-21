package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horizon Drake — Creature — Drake {1}{U}{U}, 3/1 (Worldwake, #5582):
//
//	"Flying, protection from lands"
//
// Protection from lands is the card-type quality game/protection.go
// parses (CR 702.16a): a land's ability can't target it, deal damage
// to it, enchant or equip it, and a land creature can't block it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9c724912-5afd-4813-affb-4dbda64fd4ee",
		Name:            "Horizon Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from lands"},
	})
}
