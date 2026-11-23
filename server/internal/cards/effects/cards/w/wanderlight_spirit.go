package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wanderlight Spirit — Creature — Spirit {2}{U}, 2/3 (Innistrad: Crimson
// Vow, S82):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Flying is printed data (CR 702.9b). "Can block only creatures with
// flying" is Cloud Spirit's block rule, checked as blockers are declared
// (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7b4c2dcf-86f4-409d-90c3-8fe83966cac0",
		Name:            "Wanderlight Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
