package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skywinder Drake — Creature — Drake {2}{U}, 3/1 (Magic 2012, #5573):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Cloud Spirit's block rule: it may block only an attacker with flying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e894ccf5-fc71-48bd-bca5-cfeb05315f42",
		Name:            "Skywinder Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
