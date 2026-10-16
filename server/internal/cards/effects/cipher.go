package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cipher.go — CR 702.99 (#3867). Cipher is the engine's keyword
// (game/cipher.go): a card file declares it in PrintedKeywords and the
// engine asks, as the spell resolves, whether to exile it encoded on a
// creature its controller controls; the creature then has the triggered
// ability that copies the encoded card and offers the copy free. What
// the copy DOES is the card's own OnResolve, run again.
//
//	PrintedKeywords: []string{game.KeywordCipher},   // Trait Doctoring
//
// The one thing filed here is the granted ability's text. The layer
// pass grants the bundle game.CipherGrantKey to the creature, with the
// encoded card as grantor, and the view lists a granted bundle by its
// text — so the creature shows what it can now do, and which card it
// would copy, like any other granted ability. The trigger itself is the
// engine's (one per encoded card), so the bundle carries none.
func init() {
	fileDef(game.GrantKey(game.CipherGrantKey), &game.CardDef{
		GrantText: "Cipher — Whenever this creature deals combat damage to a player, you may copy the encoded card and you may cast the copy without paying its mana cost.",
	})
}
