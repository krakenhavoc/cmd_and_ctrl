package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// satyr_cant_block_token.go — the 1/1 red Satyr creature token with "This
// token can't block." that Anax, Hardened in the Forge makes (Commander
// Masters, S82 GAP batch 104, #5948).
//
// A catalog template and not a tokens_table row because the can't-block
// clause is a static the token always prints (the Fungus's shape).
// Its slug is an on-disk identity: never renamed, never reused.
func printedGapB104SatyrCantBlockToken() TokenTemplate {
	return TokenTemplate{
		Slug: "satyr-cant-block",
		Card: game.Card{
			Name:      "Satyr",
			TypeLine:  "Token Creature — Satyr",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"R"},
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Text:   "This token can't block.",
	}
}

// GapB104SatyrCantBlockToken is the Satyr's card value for CreateToken.
func GapB104SatyrCantBlockToken() game.Card {
	return TokenFromCatalog(printedGapB104SatyrCantBlockToken)
}
