package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stormbound Geist — Creature — Spirit {1}{U}{U}, 2/2:
//
//	"Flying
//	 This creature can block only creatures with flying.
//	 Undying"
//
// Flying and undying are PrintedKeywords (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b5027d79-83e9-437a-baae-03cc3a13a608",
		Name:            "Stormbound Geist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordUndying},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
