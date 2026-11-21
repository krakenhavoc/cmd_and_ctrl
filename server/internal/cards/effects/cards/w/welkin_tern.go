package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Welkin Tern — Creature — Bird {1}{U}, 2/1 (Zendikar, #5583):
//
//	"Flying (This creature can't be blocked except by creatures with
//	 flying or reach.)
//	 This creature can block only creatures with flying."
//
// Cloud Djinn's shape: flying is the printed keyword, and the block
// rule keeps it from blocking an attacker without flying (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "010c660a-a67b-48f0-a63a-15470eb89f69",
		Name:            "Welkin Tern",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
