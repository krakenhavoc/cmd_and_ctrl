package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloud Dragon — Creature — Illusion Dragon {5}{U}, 5/4 (Portal, #5630):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Flying is printed data (CR 702.9b). "Can block only creatures with
// flying" is Stormbound Geist's block rule, a restriction checked when
// blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3c46f309-69ae-43b1-adf6-bdb26599c1f4",
		Name:            "Cloud Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
