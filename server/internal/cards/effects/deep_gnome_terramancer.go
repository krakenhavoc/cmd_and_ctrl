package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deep Gnome Terramancer — Creature — Gnome Wizard {1}{W}, 2/2:
//
//	"Flash
//	 Mold Earth — Whenever one or more lands enter under an opponent's
//	 control without being played, you may search your library for a
//	 Plains card, put it onto the battlefield tapped, then shuffle. Do
//	 this only once each turn."
//
// The proof card for #1326 (CR 305.4): "without being played" is a
// distinction the engine did not carry on the entry event at all
// until this issue. It now does — Event.Played, stamped by the
// shared entry finisher (entry_choice.go) from the settled event's
// landPlay flag, true for the one entry that is a land PLAY and
// false for every "put" path and for a token. Mold Earth is the
// first printed clause that reads the false side on purpose: an
// opponent's fetchland, reanimated land or ramp spell trigger it, and
// an opponent's ordinary land drop does not.
//
// Flash rides PrintedKeywords. The rest is three shared pieces:
// `dgtLandEnteredOpponentControlWithoutBeingPlayed` is the condition;
// `OncePerBatch` (#587) collapses a simultaneous "one or more" land
// entry (Scapeshift-style) into the one trigger the printed text
// describes; and `b11TriggeredThisTurn` (Exemplar of Light's "only
// once each turn" tally) gates the SECOND land-entry batch of the
// turn out entirely, which OncePerBatch alone does not do since it
// only dedupes within one batch. The search itself is Kor
// Cartographer's: "a Plains card" is the printed subtype filter
// (b39IsPlainsCard), not the literal name, "you may" forces the
// prompt (SearchLibrary.Optional), and the fetched land enters tapped
// (TappedOnEntry) and the library is reshuffled (Shuffle).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d4e3440d-4e34-40d7-8a42-c673225c0332",
		Name:            "Deep Gnome Terramancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny([]game.EventKind{game.EventZoneMove, game.EventTokenCreated},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if !dgtLandEnteredOpponentControlWithoutBeingPlayed(ev, source, g) {
						return false
					}
					return !b11TriggeredThisTurn(g, source.InstanceID, dgtMoldEarthLabel)
				}, dgtMoldEarthLabel, func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     b39IsPlainsCard,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						TappedOnEntry: true,
						Optional:      true,
						Shuffle:       true,
						Source:        item.SourceCardID,
						Reason:        dgtMoldEarthLabel,
					}.Apply(NewContext(g, item))
				})),
		},
	})
}

const dgtMoldEarthLabel = "Deep Gnome Terramancer — search for a Plains card"

// dgtLandEnteredOpponentControlWithoutBeingPlayed is Mold Earth's
// condition: a land landed under the control of someone other than
// the source's controller — via an ordinary battlefield entry
// (EventZoneMove) or a token's creation (EventTokenCreated) — and
// Event.Played (CR 305.4, #1326) says the entry was PUT there rather
// than PLAYED.
func dgtLandEnteredOpponentControlWithoutBeingPlayed(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Played || ev.CardID == uuid.Nil {
		return false
	}
	switch ev.Kind {
	case game.EventZoneMove:
		if ev.NewZone != game.ZoneBattlefield {
			return false
		}
	case game.EventTokenCreated:
		// fine: a created token has no OldZone/NewZone to check.
	default:
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsLand() {
		return false
	}
	return c.Controller != uuid.Nil && c.Controller != source.Controller
}
