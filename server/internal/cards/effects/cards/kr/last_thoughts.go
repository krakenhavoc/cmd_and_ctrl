package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Last Thoughts — Sorcery {3}{U}:
//
//	"Draw a card.
//	 Cipher (Then you may exile this spell card encoded on a creature you
//	 control. Whenever that creature deals combat damage to a player, its
//	 controller may cast a copy of the encoded card without paying its
//	 mana cost.)"
//
// Cipher is the engine's keyword (game/cipher.go, CR 702.99, #3867):
// the declaration is the whole of it. A copy cast through the encoded
// creature draws for whoever cast it — the creature's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "649d983a-e861-4893-8ae8-fb0a22cfad0a",
		Name:            "Last Thoughts",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordCipher},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
