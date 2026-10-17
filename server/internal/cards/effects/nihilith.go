package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nihilith — Creature — Horror {4}{B}{B}, 4/4:
//
//	"Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)
//	 Suspend 7—{1}{B} (Rather than cast this card from your hand, you
//	 may pay {1}{B} and exile it with seven time counters on it. At the
//	 beginning of your upkeep, remove a time counter. When the last is
//	 removed, you may cast it without paying its mana cost. It has
//	 haste.)
//	 Whenever a card is put into an opponent's graveyard from anywhere,
//	 if this card is suspended, you may remove a time counter from this
//	 card."
//
// Fear is an enforced evasion keyword; suspend is the shared special
// action (Suspend), which also grows the countdown and the free-cast
// triggers from the declaration. What is new is the third ability: a
// trigger that works FROM EXILE (CR 113.6 — the card says it does, by
// naming itself as suspended), declared with InExile, the same zone
// dimension suspend's own triggers use. "You" there is the card's owner
// (CR 108.4), so "an opponent's graveyard" is any graveyard but the
// owner's.
//
// "A card is put into a graveyard from anywhere" is read the way every
// such card in the catalog reads it (ThisWasPutIntoAGraveyardFromAnywhere,
// creatureCardPutIntoAGraveyard): one event per arrival, whichever
// route it took — a death or sacrifice (EventZoneMove), a mill, a
// discard, a countered spell — and never a card that was already in a
// graveyard, never a token (CR 108.2, a token is not a card), and
// never a card a replacement sent somewhere else. Each card is its own
// trigger, so a mill of three cards offers three removals, one "may" per
// card.
//
// The intervening "if this card is suspended" (CR 603.4) is checked when
// the ability would trigger (CardIsSuspended on the card in exile): a
// copy of the card exiled by anything else carries no time counters and
// so is not suspended. As the ability resolves it acts on the same
// OBJECT (CR 400.7): a Nihilith that left exile and came back suspended
// again is a new object that this ability cannot find, and one whose
// last counter came off in response has none left to lose. Taking the
// LAST counter off is an ordinary counter removal, so the free-cast
// offer arrives by the same road the upkeep tick uses
// (SuspendLastCounterTrigger watches the counter event).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c235fda2-1a8d-4ac7-9a23-784529a520f0",
		Name:            "Nihilith",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear"},
		SpecialActions: []game.SpecialAction{
			Suspend(7, "{1}{B}"),
		},
		Triggered: []game.TriggeredAbility{
			Optional(InExile(OnAny(
				[]game.EventKind{game.EventZoneMove, game.EventMill, game.EventDiscardCard, game.EventCounterSpell},
				nihilithCardInOpponentsGraveyard,
				"Nihilith — remove a time counter from this card", nihilithRemoveTimeCounter)),
				"Nihilith — remove a time counter from this card?"),
		},
	})
}

// nihilithCardInOpponentsGraveyard is the trigger condition: the event
// put a card into a graveyard other than the owner's, and the source is
// suspended.
func nihilithCardInOpponentsGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !game.CardIsSuspended(*source) {
		return false
	}
	_, ok := cardPutIntoAnOpponentsGraveyard(ev, source.Controller, g)
	return ok
}

// cardPutIntoAnOpponentsGraveyard reports whether the event put a CARD
// into the graveyard of a player other than `you`, and whose graveyard
// it was. The engine emits one event per arrival
// (game/zone_route.go): a zone move, a discard and a mill each name
// their destination; a countered spell carries none, so the card is
// looked for where it is now. A card that was already in a graveyard
// has not been put into one, and a token is not a card.
func cardPutIntoAnOpponentsGraveyard(ev game.Event, you uuid.UUID, g *game.Game) (uuid.UUID, bool) {
	if ev.CardID == uuid.Nil {
		return uuid.Nil, false
	}
	switch ev.Kind {
	case game.EventZoneMove, game.EventDiscardCard, game.EventMill:
		if ev.NewZone != game.ZoneGraveyard || ev.OldZone == game.ZoneGraveyard {
			return uuid.Nil, false
		}
	case game.EventCounterSpell:
		if ev.Target != ev.CardID {
			return uuid.Nil, false
		}
	default:
		return uuid.Nil, false
	}
	z := g.FindCardZoneForEffect(ev.CardID)
	if z == nil || z.Kind != game.ZoneGraveyard || z.Owner == uuid.Nil || z.Owner == you {
		return uuid.Nil, false
	}
	if c, ok := g.LookupCardForEffect(ev.CardID); !ok || c.IsToken() {
		return uuid.Nil, false
	}
	if p := g.PlayerByIDForEffect(z.Owner); p == nil {
		return uuid.Nil, false
	}
	return z.Owner, true
}

// nihilithRemoveTimeCounter is the ability's body: remove one time
// counter from this card if it is still the exiled object it was when
// the ability triggered. "Suspended" is "in exile with a time counter"
// (CR 702.62b), so a card whose counters are gone has nothing to lose:
// removing one from it changes nothing and emits no event, which is why
// the body needs no second suspended check.
func nihilithRemoveTimeCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	z := g.FindCardZoneForEffect(ref.ID)
	if z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	c, ok := g.LookupCardForEffect(ref.ID)
	if !ok || c.ObjectEpoch != ref.Epoch {
		return nil
	}
	return g.AddCounterForEffect(ref.ID, game.CounterTime, -1)
}
