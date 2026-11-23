package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whispering Madness — Sorcery {2}{U}{B}:
//
//	"Each player discards their hand, then draws cards equal to the
//	 greatest number of cards a player discarded this way.
//	 Cipher (Then you may exile this spell card encoded on a creature you
//	 control. Whenever that creature deals combat damage to a player, its
//	 controller may cast a copy of the encoded card without paying its
//	 mana cost.)"
//
// Jace's Archivist's wheel (b17WheelToGreatestDiscard: every hand goes
// first, then every player draws the greatest count), and cipher
// (game/cipher.go, CR 702.99, #3867).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7bbc5f3a-f8fb-47e8-837e-ce319bc489de",
		Name:            "Whispering Madness",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordCipher},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B17WheelToGreatestDiscard(ctx.Game, item)
		},
	})
}
