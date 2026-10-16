package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shadow Slice — Sorcery {4}{B}:
//
//	"Target opponent loses 3 life.
//	 Cipher (Then you may exile this spell card encoded on a creature you
//	 control. Whenever that creature deals combat damage to a player, its
//	 controller may cast a copy of the encoded card without paying its
//	 mana cost.)"
//
// Cipher is the engine's keyword (game/cipher.go, CR 702.99, #3867). A
// copy cast through the encoded creature targets an opponent of the
// player casting it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f4d2eca4-4645-4044-b75f-442cd726cedf",
		Name:            "Shadow Slice",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordCipher},
		Targets:         TargetPlayer("target opponent", Opponent()),
		OnResolve:       shadowSliceLoseLife,
	})
}

// shadowSliceLoseLife is "Target opponent loses 3 life": the target if
// it is still legal (CR 608.2b).
func shadowSliceLoseLife(_ *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return GainLife{Player: t.ID, Amount: -3}.Apply(ctx)
		}
	}
	return nil
}
