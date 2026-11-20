package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scrapskin Drake — {2}{U} Creature — Zombie Drake 2/3 (Avacyn Restored, #5570):
//
//	"Flying (This creature can't be blocked except by creatures with
//	 flying or reach.)
//	 This creature can block only creatures with flying."
//
// Flying and Wanderlight Spirit's block rule: it may block only an
// attacker with flying (CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "910e8269-6001-479a-ab2e-5a935569c33c",
		Name:            "Scrapskin Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
