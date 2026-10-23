package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Field of Dreams — World Enchantment {U}:
//
//	"Players play with the top card of their libraries revealed."
//
// #1997, one of the world enchantments ADR 0109 PR 7 left out. Every
// player's top library card is revealed to every player while Field of
// Dreams is on the battlefield, whoever controls it
// (game.LibraryTopRevealedEachPlayer, read by
// game.LibraryTopVisibilityLocked for every library). The reveal is
// derived on every frame, so a new top card is revealed the moment it
// is on top (CR 401.5). It grants no permission to play the card.
// A world permanent: the world rule (CR 704.5k, ADR 0109 §8) puts it
// into its owner's graveyard when a newer one enters.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:          "6dcc62f8-2fd9-473b-9e6a-2dd7e610f855",
		Name:              "Field of Dreams",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopRevealedEachPlayer,
	})
}
