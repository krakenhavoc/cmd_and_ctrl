package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voidwalk — Sorcery {3}{U}:
//
//	"Exile target creature. Return it to the battlefield under its
//	 owner's control at the beginning of the next end step.
//	 Cipher (Then you may exile this spell card encoded on a creature you
//	 control. Whenever that creature deals combat damage to a player, its
//	 controller may cast a copy of the encoded card without paying its
//	 mana cost.)"
//
// Eerie Interlude's delayed blink (exileTargetsThenScheduleReturn: the
// exile now, a CR 603.7 delayed trigger returning it under its owner's
// control at the next end step) on one target creature, and cipher
// (game/cipher.go, CR 702.99, #3867). Blinking the creature it would be
// encoded on leaves nothing to encode it on: the creature is in exile
// when cipher asks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6706ebe4-4f96-4788-a934-f9b9aac5746f",
		Name:            "Voidwalk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordCipher},
		Targets:         TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return exileTargetsThenScheduleReturn(ctx, "Voidwalk — return the exiled creature")
		},
	})
}
