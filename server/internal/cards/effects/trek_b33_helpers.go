package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b33_helpers.go — shared pieces for Star Trek Commander batch 33
// (#4974). Names are prefixed trekB33 so they cannot collide with another
// batch's.

// --- Tribble token -----------------------------------------------------------

// trekB33TribbleSlug is the catalog identity of the Tribble token.
const trekB33TribbleSlug = "tribble"

// printedTrekB33TribbleToken is the 1/1 white Tribble creature token with
// "This token can't block." (Tribble Cluster). A catalog template and not
// a tokens_table row because the can't-block clause is a static the token
// always prints (the Phyrexian Mite's shape). The token card in the dump
// (ttrc 2) lists no colour; the oracle text of the card that makes it says
// white, and the template follows the oracle text (CR 111.4), as The
// Eagles Are Coming!'s Bird Soldier does.
func printedTrekB33TribbleToken() tokenTemplate {
	return tokenTemplate{
		Slug: trekB33TribbleSlug,
		Card: game.Card{
			Name:      "Tribble",
			TypeLine:  "Token Creature — Tribble",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"W"},
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Text:   "This token can't block.",
	}
}

// TribbleToken is the Tribble's card value for CreateToken.
func TribbleToken() game.Card { return tokenFromCatalog(printedTrekB33TribbleToken) }
