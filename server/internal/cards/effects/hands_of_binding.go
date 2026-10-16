package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hands of Binding — Sorcery {1}{U}:
//
//	"Tap target creature an opponent controls. That creature doesn't
//	 untap during its controller's next untap step.
//	 Cipher (Then you may exile this spell card encoded on a creature you
//	 control. Whenever that creature deals combat damage to a player, its
//	 controller may cast a copy of the encoded card without paying its
//	 mana cost.)"
//
// Frost Breath's tap-and-freeze on one creature an opponent controls,
// and cipher, the engine's keyword (game/cipher.go, CR 702.99, #3867).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9bc82078-2f3e-44ee-906d-707f3430d414",
		Name:            "Hands of Binding",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordCipher},
		Targets:         TargetCreature("target creature an opponent controls", OpponentControls()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b751TapFreezeTargets(ctx, uuid.Nil, "Hands of Binding")
		},
	})
}
