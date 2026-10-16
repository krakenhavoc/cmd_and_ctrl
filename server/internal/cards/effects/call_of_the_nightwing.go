package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Call of the Nightwing — Sorcery {2}{U}{B}:
//
//	"Create a 1/1 blue and black Horror creature token with flying.
//	 Cipher (Then you may exile this spell card encoded on a creature you
//	 control. Whenever that creature deals combat damage to a player, its
//	 controller may cast a copy of the encoded card without paying its
//	 mana cost.)"
//
// Cipher is the engine's keyword (game/cipher.go, CR 702.99, #3867).
// The Horror is created before the cipher question is asked, so the
// token itself may be the creature the card is encoded on, as printed.
// A copy cast through the encoded creature makes its Horror for the
// creature's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "919d21ad-3d1e-4f80-8005-64e3f8bf64d1",
		Name:            "Call of the Nightwing",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordCipher},
		Purpose:         game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: ctx.Controller(), Template: callOfTheNightwingHorror, N: 1}.Apply(ctx)
		},
	})
}

// callOfTheNightwingHorror is the 1/1 blue and black Horror with flying.
// Not a tokens_table row: no other catalog card prints this Horror.
var callOfTheNightwingHorror = game.Card{
	Name: "Horror", TypeLine: "Token Creature — Horror", Power: 1, Toughness: 1,
	Colors: []string{"U", "B"}, Keywords: []string{"flying"},
}
