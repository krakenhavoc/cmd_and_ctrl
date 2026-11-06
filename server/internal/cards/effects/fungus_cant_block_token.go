package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// fungus_cant_block_token.go — the 1/1 black Fungus creature token with
// "This token can't block." that The Lost Caverns of Ixalan's Fungi make
// (Synapse Necromage, The Mycotyrant, #4470).
//
// A catalog template and not a tokens_table row because the can't-block
// clause is a static the token always prints (the Phyrexian Mite's and the
// Tribble's shape). The colour is the oracle text's: "black Fungus".

// printedFungusCantBlockToken is the Fungus as printed, its static
// included. Its slug is an on-disk identity: never renamed, never reused.
func printedFungusCantBlockToken() tokenTemplate {
	return tokenTemplate{
		Slug: "fungus-cant-block",
		Card: game.Card{
			Name:      "Fungus",
			TypeLine:  "Token Creature — Fungus",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"B"},
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Text:   "This token can't block.",
	}
}

// FungusCantBlockToken is the Fungus's card value for CreateToken.
func FungusCantBlockToken() game.Card { return tokenFromCatalog(printedFungusCantBlockToken) }
