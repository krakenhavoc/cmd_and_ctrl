package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Waterlogged Teachings // Inundated Archive — modal double-faced card.
// This file is the FRONT face, Instant {3}{U/B}:
//
//	"Search your library for an instant card or a card with flash,
//	 reveal it, put it into your hand, then shuffle."
//
// The back face, Inundated Archive, is registered with the MDFC land
// cycle in mdfc_lands.go under "<oracle>#1".
//
// The S22 tutor body (b06TutorToHand). "A card with flash" is the
// keyword as the card in the library has it, which is its printed
// keyword list: game.HasKeyword falls back to the card's own Keywords
// and the catalog's printed ones off the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e6ad1be9-f13d-4590-b3db-e2d0fff46f03",
		Name:         "Waterlogged Teachings",
		Completeness: CompletenessFull,
		OnResolve: b06TutorToHand("Waterlogged Teachings — an instant card or a card with flash, revealed, to your hand",
			waterloggedTeachingsFinds),
	})
}

// waterloggedTeachingsFinds is "an instant card or a card with flash".
func waterloggedTeachingsFinds(c game.Card) bool {
	return c.IsInstant() || game.HasKeyword(&c, "flash")
}
