package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mental Vapors — Sorcery {3}{B}:
//
//	"Target player discards a card.
//	 Cipher (Then you may exile this spell card encoded on a creature you
//	 control. Whenever that creature deals combat damage to a player, its
//	 controller may cast a copy of the encoded card without paying its
//	 mana cost.)"
//
// The discarding player picks the card (CR 701.9a), and cipher is the
// engine's keyword (game/cipher.go, CR 702.99, #3867): the encoding
// question waits for the discard to be answered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e0151f37-790b-47f4-9928-cd8f32f68e4c",
		Name:            "Mental Vapors",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordCipher},
		Targets:         effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TargetPlayerDiscards(1)(ctx.Game, item)
		},
	})
}
