package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cmdr2_b27_helpers.go — card-side helpers of the cmdr2 batch 27 cards.
//
// Each key is on disk (ledgered "choose <key>") and is never renamed or
// reused.

var (
	// cmdr2B27TakeRestOnTop is "put one of them into your hand and the rest
	// on top of your library in any order" (Diabolic Vision).
	cmdr2B27TakeRestOnTop = RegisterTake("choose-cards/cmdr2-b27-take-rest-top-any-order",
		TakeDef{Then: cmdr2B27RestOnTopInAnyOrder})

	// cmdr2B27TakeRestTopAndBottom is "put one of those cards into your hand,
	// one on top of your library, and one on the bottom" (Telling Time): once
	// a card is in hand, the player says which of the rest goes on top.
	cmdr2B27TakeRestTopAndBottom = RegisterTake("choose-cards/cmdr2-b27-take-rest-top-and-bottom",
		TakeDef{Then: cmdr2B27RestOneTopRestBottom})
)

// cmdr2B27RestOnTopInAnyOrder puts the rest of a look back on top of the
// library, in an order the player chooses. A pile of one has no order to
// ask for.
func cmdr2B27RestOnTopInAnyOrder(g *game.Game, res TakeFromLibraryResult) error {
	if len(res.Rest) == 0 {
		return nil
	}
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   res.Player,
		Source:    res.Source,
		Cards:     res.Rest,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceTop,
		Reason:    libraryOrderLabel(g, res.Source, "put the rest on top of your library in any order"),
	})
}

// cmdr2B27RestOneTopRestBottom puts exactly one of the rest on top of the
// library and the other on the bottom, the player's choice of which (Cream of
// the Crop's lane rule, TopCount 1).
func cmdr2B27RestOneTopRestBottom(g *game.Game, res TakeFromLibraryResult) error {
	if len(res.Rest) == 0 {
		return nil
	}
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   res.Player,
		Source:    res.Source,
		Cards:     res.Rest,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceTopOrBottom,
		TopCount:  1,
		Reason:    libraryOrderLabel(g, res.Source, "put one of those cards on top of your library and the other on the bottom"),
	})
}

// "then discard two cards unless you discard an enchantment card" —
// Thirst for Meaning (and Arm-Mounted Anchor's sentence): Thirst for
// Knowledge's prompt with enchantment cards where that card has artifacts.
var drawThenDiscardTwoUnlessEnchantment = discardTwoUnlessOne("discard-two-unless-enchantment",
	"discard two cards, or one enchantment card",
	func(c game.Card) bool { return c.IsEnchantment() })
