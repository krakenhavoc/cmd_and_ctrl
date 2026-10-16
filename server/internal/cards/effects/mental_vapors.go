package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "e0151f37-790b-47f4-9928-cd8f32f68e4c",
		Name:            "Mental Vapors",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordCipher},
		Targets:         TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return targetPlayerDiscards(1)(ctx.Game, item)
		},
	})
}
