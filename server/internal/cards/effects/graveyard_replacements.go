package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_replacements.go — the CR 614 replacement on "if a card
// would be put into a graveyard from anywhere". Its own file rather
// than helpers.go, per the convention enters_tapped.go set and
// discard_replacements.go followed.
//
// The whole family is one sentence with one clause moved around:
//
//	Rest in Peace        every graveyard, cards and tokens alike
//	Leyline of the Void  an OPPONENT's graveyard only
//	Dauthi Voidwalker    an opponent's, and the exiled card gets a
//	                     void counter and may be cast — its own work
//
// "From anywhere" is the load-bearing half, and it is why this could
// not be written before #931: a replacement only ever runs for a mover
// that pushes a move event through the CR 614 window, and until #931
// two graveyard arrivals did not — a library search that puts the card
// into a graveyard (Entomb, Buried Alive) and surveil's graveyard leg.
// Mill went through in #529, the battlefield exit has since S17, and a
// discard since #650. Every one of them now does, so "anywhere" is
// anywhere.
//
// The battlefield is included, which is the half worth saying out
// loud: a creature that would DIE is exiled instead, so it never died
// and its own dies-triggers never fire (CR 700.4) — the same shape
// Stone of Erech and Liesa, Forgotten Archangel already have, widened
// from "a creature an opponent controls" to every card in every zone.

// GraveyardBecomesExile describes "if a card would be put into a
// graveyard from anywhere, exile it instead" (CR 614.1a), with the one
// clause that differs between its members as a field.
type GraveyardBecomesExile struct {
	// OpponentsOnly restricts the effect to graveyards that are not
	// the source's controller's — Leyline of the Void's "into an
	// OPPONENT's graveyard". False is Rest in Peace: every graveyard,
	// the controller's own included.
	//
	// The graveyard a card is put into is always its owner's, so this
	// reads the destination the event names rather than deriving an
	// owner of its own.
	OpponentsOnly bool

	// NotControlledByYou restricts the effect to cards the source's
	// controller does NOT control — Valgavoth, Terror Eater's "if a
	// card YOU DIDN'T CONTROL would be put into an opponent's
	// graveyard". False leaves the family exactly as it was.
	//
	// It is a separate question from OpponentsOnly, and Valgavoth
	// asks both. The graveyard is always the card's OWNER's, so
	// OpponentsOnly alone would eat a creature you stole from an
	// opponent and then lost — which its owner's graveyard would
	// receive, but which you controlled. That is stronger than
	// printed, and stronger is the direction a simplification may
	// never go.
	//
	// Read off the card as it still sits in the zone it is leaving,
	// which is what makes the battlefield case work: control is only
	// ever different from ownership on the battlefield (Card.Controller
	// equals Owner everywhere else), and the CR 614 window runs before
	// the move.
	NotControlledByYou bool

	// Label is the CR 616 prompt header, shown when this and another
	// replacement both apply to the same move.
	Label string
}

// Build turns the description into the ReplacementEffect a Spec
// declares.
func (r GraveyardBecomesExile) Build() game.ReplacementEffect {
	return game.ReplacementEffect{
		// Two kinds, for the reason the CR 903.9 built-in declares two:
		// #650 split the discard off into its own event, so "from
		// anywhere" has to include a discarded card and a discard no
		// longer arrives as a plain move. The discard still HAPPENED
		// (CR 701.8a defines it by the move out of the hand), so
		// Megrim and the rest of the family still see it; only the
		// destination changes.
		Watches: []game.EventKind{game.EventZoneMove, game.EventDiscardCard},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove && ev.Kind != game.RepEventDiscard {
				return false
			}
			if ev.NewZone != game.ZoneGraveyard {
				return false
			}
			if r.OpponentsOnly && (ev.NewZoneOwner == uuid.Nil || ev.NewZoneOwner == src.Controller) {
				return false
			}
			if r.NotControlledByYou {
				moving, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || moving.Controller == src.Controller {
					return false
				}
			}
			return true
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.NewZone = game.ZoneExile
			ev.NewZoneOwner = uuid.Nil
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: r.Label,
	}
}
