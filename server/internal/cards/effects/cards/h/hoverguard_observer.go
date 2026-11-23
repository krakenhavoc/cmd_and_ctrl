package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hoverguard Observer — Creature — Drone {2}{U}{U} (Darksteel, #5605):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Cloud Spirit's block rule: it may not block an attacker without flying
// (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5ca8d756-a55a-4b36-b0a6-562a39484952",
		Name:            "Hoverguard Observer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
