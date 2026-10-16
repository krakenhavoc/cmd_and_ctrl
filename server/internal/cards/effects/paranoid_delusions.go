package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Paranoid Delusions — Sorcery {U}{B}:
//
//	"Target player mills three cards.
//	 Cipher (Then you may exile this spell card encoded on a creature you
//	 control. Whenever that creature deals combat damage to a player, its
//	 controller may cast a copy of the encoded card without paying its
//	 mana cost.)"
//
// A mill (CR 701.17), and cipher, the engine's keyword (game/cipher.go,
// CR 702.99, #3867). A copy cast through the encoded creature chooses
// its own target player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "129cec2c-fc1b-4ce6-98f2-2123e4dbd1ae",
		Name:            "Paranoid Delusions",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordCipher},
		Targets:         TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					return MillCards{Player: t.ID, N: 3}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
