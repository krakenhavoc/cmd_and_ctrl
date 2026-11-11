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

// --- batch 28: "whenever you surveil" payoffs --------------------------------

// cmdr2B27WheneverYouSurveil is "Whenever you surveil, <effect>": the
// payoff trigger of the surveil cards. EventSurveil is emitted once a
// surveil is complete (CR 701.25), so a surveil that looked at nothing
// does not trigger it.
func cmdr2B27WheneverYouSurveil(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventSurveil, ByYou, label, effect)
}

// cmdr2B27IsYourFirstSurveilThisTurn is "whenever you surveil for the first
// time each turn": the event is the controller's, and no earlier surveil of
// theirs is in the turn's log. The tally is the player's, so a permanent that
// enters after their first surveil of the turn sees no first one.
func cmdr2B27IsYourFirstSurveilThisTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSurveil || ev.Actor != source.Controller {
		return false
	}
	for _, prev := range g.EventsThisTurn() {
		if prev.Seq >= ev.Seq {
			break
		}
		if prev.Kind == game.EventSurveil && prev.Actor == ev.Actor {
			return false
		}
	}
	return true
}

// cmdr2B27SurveilPlusTwoMay is "You may look at an additional two cards each
// time you surveil" (Enhanced Surveillance): SurveilPlusOne's replacement with
// +2, and a "may" the player answers each time (CR 614.1a).
func cmdr2B27SurveilPlusTwoMay(label string) game.ReplacementEffect {
	r := KeywordActionBecomes(game.KeywordActionSurveil, cmdr2B27PlusTwo, label)
	r.Optional = true
	r.Label = label
	return r
}

func cmdr2B27PlusTwo(n int) int { return n + 2 }
