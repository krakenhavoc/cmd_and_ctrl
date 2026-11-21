package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skywinder Drake — Creature — Drake {2}{U}, 3/1 (Magic 2012, #5573):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Cloud Spirit's block rule: it may block only an attacker with flying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e894ccf5-fc71-48bd-bca5-cfeb05315f42",
		Name:            "Skywinder Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
