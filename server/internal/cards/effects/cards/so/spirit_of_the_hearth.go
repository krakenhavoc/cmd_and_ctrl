package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spirit of the Hearth — Creature — Cat Spirit {4}{W}{W}, 4/5 (Eventide, #5590):
//
//	"Flying
//	 You have hexproof. (You can't be the target of spells or abilities
//	 your opponents control.)"
//
// Leyline of Sanctity's player keyword, granted while the Spirit is on the
// battlefield and gone when it leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "30cd69a8-7893-4075-94ca-04450ff821d3",
		Name:            "Spirit of the Hearth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		PlayerKeywords:  []string{effects.KeywordHexproof},
	})
}
