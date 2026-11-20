package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "5ca8d756-a55a-4b36-b0a6-562a39484952",
		Name:            "Hoverguard Observer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
