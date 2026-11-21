package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloud Spirit — Creature — Spirit {2}{U}, 3/1 (Portal, #5630):
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
	effects.Register(effects.Spec{
		OracleID:        "9e1a6481-f460-4551-96e8-30b289f2cb92",
		Name:            "Cloud Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
