package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "010c660a-a67b-48f0-a63a-15470eb89f69",
		Name:            "Welkin Tern",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
