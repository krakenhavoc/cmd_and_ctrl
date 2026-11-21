package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stronghold Zeppelin — Creature — Human {2}{U}{U}, 3/3
// (Nemesis, #5618):
//
//	"Flying
//	 This creature can block only creatures with flying."
//
// Flying is printed data (CR 702.9b). "Can block only creatures with
// flying" is Cloud Spirit's block rule (nemBlocksOnlyFlyers), judged as
// blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dbd86a37-8ac1-4bdb-9ba1-7caee99ab5f7",
		Name:            "Stronghold Zeppelin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules:      effects.NemBlocksOnlyFlyers(effects.OnSelf()),
	})
}
