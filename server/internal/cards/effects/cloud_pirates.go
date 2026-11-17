package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloud Pirates — Creature — Human Pirate {U}, 1/1 (Portal, #5630):
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
		OracleID:        "d334aa85-3470-4f5d-9cbc-b88bf991a5af",
		Name:            "Cloud Pirates",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
