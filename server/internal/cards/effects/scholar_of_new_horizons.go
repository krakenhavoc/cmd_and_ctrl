package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scholar of New Horizons — Creature — Human Scout {1}{W}, 1/1:
//
//	"This creature enters with a +1/+1 counter on it.
//	 {T}, Remove a counter from a permanent you control: Search your
//	 library for a Plains card and reveal it. If an opponent controls
//	 more lands than you, you may put that card onto the battlefield
//	 tapped. If you don't put the card onto the battlefield, put it into
//	 your hand. Then shuffle."
//
// The counter cost is Fain, the Broker's "a counter" of any kind
// (RemoveCountersFrom with no kind), off any permanent you control, the
// Scholar included, paid at announce.
//
// The search finds and reveals the Plains and leaves it on top of the
// shuffled library (SearchLibrary ToTop), so the card's destination is
// chosen after it is found, as printed: with an opponent ahead on lands
// (counted as the ability resolves) the controller may put it onto the
// battlefield tapped (Risen Reef's library-to-battlefield move), and
// otherwise, or on a decline, it goes to hand. Taking the card off the
// top of a library that was just shuffled is the same as shuffling
// after it left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c7ea603-c985-4107-806b-a467b5fcca36",
		Name:         "Scholar of New Horizons",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 1, "Scholar of New Horizons enters with a +1/+1 counter"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Remove a counter from a permanent you control: Search your library for a Plains card and reveal it. If an opponent controls more lands than you, you may put that card onto the battlefield tapped. If you don't put the card onto the battlefield, put it into your hand. Then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(TapCost(), RemoveCountersFrom("", 1, "a permanent you control", Permanent())),
			Effect:  scholarOfNewHorizonsSearch,
		}},
	})
}

func scholarOfNewHorizonsSearch(g *game.Game, item *game.StackItem) error {
	controller, source := item.Controller, item.SourceCardID
	return SearchLibrary{
		Player:    controller,
		Predicate: IsLandWithSubtype("Plains"),
		Dest:      game.ZoneLibrary,
		ToTop:     true,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Scholar of New Horizons — a Plains card",
		Source:    source,
		Then: func(g *game.Game, found []uuid.UUID) error {
			if len(found) == 0 {
				return nil
			}
			if !b03OpponentControlsMoreLands(g, controller) {
				return scholarOfNewHorizonsToHand(g, found)
			}
			return PutFromLibraryOntoBattlefield{
				Player:   controller,
				Cards:    found,
				Match:    Land(),
				Max:      1,
				Optional: true,
				Tapped:   true,
				Label:    "Scholar of New Horizons — you may put the Plains onto the battlefield tapped (otherwise it goes into your hand)",
				Then: func(g *game.Game, res PutFromLibraryResult) error {
					return scholarOfNewHorizonsToHand(g, res.Rest)
				},
			}.Apply(NewContext(g, &game.StackItem{Controller: controller, SourceCardID: source}))
		},
	}.Apply(NewContext(g, item))
}

// scholarOfNewHorizonsToHand puts the found Plains still in the library
// into its owner's hand.
func scholarOfNewHorizonsToHand(g *game.Game, ids []uuid.UUID) error {
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneLibrary {
			continue
		}
		if err := g.BounceToHandForEffect(id); err != nil {
			return err
		}
	}
	return nil
}
