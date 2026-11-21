package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vine Mare — Creature — Elemental Horse {2}{G}{G}, 5/3
// (Core Set 2019, #5527):
//
//	"Hexproof (This creature can't be the target of spells or abilities your opponents control.)
//	This creature can't be blocked by black creatures."
//
// Hexproof, and the block rule Dauthi Horror uses, for black creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "499c7ee3-08a3-4f40-809a-3f7087ab9b24",
		Name:            "Vine Mare",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.OfColor("B"), "black creatures"),
		},
	})
}
