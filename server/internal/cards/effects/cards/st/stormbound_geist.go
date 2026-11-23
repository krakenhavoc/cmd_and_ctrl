package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "b5027d79-83e9-437a-baae-03cc3a13a608",
		Name:            "Stormbound Geist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordUndying},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})
}
