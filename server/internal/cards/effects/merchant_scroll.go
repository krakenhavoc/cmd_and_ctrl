package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merchant Scroll — Sorcery {1}{U} (EDHREC rank 2557):
//
//	"Search your library for a blue instant card, reveal that card,
//	 put it into your hand, then shuffle."
//
// The blue deck's Counterspell tutor. Type-and-colour-filtered search
// to hand, revealed, through the S22 chooser — Idyllic Tutor's shape
// (b06TutorToHand) with "blue instant" as the predicate. Colour is
// the card's own; a colourless instant with a blue hybrid symbol is
// blue by Scryfall's computed colours, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "86cebe2a-95e7-4f22-99cc-e805aeaf347e",
		Name:         "Merchant Scroll",
		Completeness: CompletenessFull,
		OnResolve: b06TutorToHand("Merchant Scroll — a blue instant card", func(c game.Card) bool {
			return c.IsInstant() && c.HasColor("U")
		}),
	})
}
