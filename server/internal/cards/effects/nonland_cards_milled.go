package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// nonland_cards_milled.go — the milled-card vocabulary of the second
// rad-counter batch (#2809): "whenever one or more nonland cards are
// milled", "this ability triggers only once each turn", and "a card in
// a graveyard that was milled this turn".
//
// Milled means a mill (CR 701.17a): the engine's EventMill, one per card
// that landed in a graveyard. A surveil's graveyard leg (CR 701.25a) and
// a search that puts a card there are not mills and emit an ordinary
// EventZoneMove (library_to_graveyard.go).

// WheneverOneOrMoreNonlandCardsAreMilled is "whenever one or more
// nonland cards are milled" (Mirelurk Queen, Screeching Scorchbeast, The
// Wise Mothman): ONE trigger for the whole mill, however many players
// milled and however many cards (CR 603.2c), not one per card as
// ANonlandCardWasMilled is.
//
// The engine emits one EventMill per card, so the ability waits for the
// event batch to settle (AtBatchEnd with ThatManyEvents, batch_total.go,
// ADR 0108 amendment 2026-10-13) and fires once for it, with
// `item.Trigger.Event.Amount` the number of NONLAND cards the batch
// milled — "that many" and "X" (nonlandCardsMilledThisWay). A batch
// that milled only lands does not trigger it. A batch is one
// resolution, one turn-based action or one special action
// (event_batch.go), so "each player mills three cards" is one trigger,
// and two mill instructions inside one resolution are one occurrence as
// they are for every "one or more" card.
//
// It is a battlefield ability: AtBatchEnd stages and settles only for a
// permanent, which is where all three cards print it.
func WheneverOneOrMoreNonlandCardsAreMilled(label string, effect Effect) game.TriggeredAbility {
	return withAtBatchEnd(On(game.EventMill, ANonlandCardWasMilled, label, effect), ThatManyEvents())
}

// nonlandCardsMilledThisWay is the settled count of a
// WheneverOneOrMoreNonlandCardsAreMilled trigger: the nonland cards its
// mill put into graveyards. At least 1.
func nonlandCardsMilledThisWay(item *game.StackItem) int {
	return settledEventAmount(item)
}

// TriggersOnlyOnceEachTurn is "This ability triggers only once each
// turn" (Mirelurk Queen): the ability's own condition is ANDed with "it
// has not been put on the stack yet this turn", per object (a new object
// is a new tally, CR 400.7). The harvester tallies the trigger as it
// queues the item, so one that was then countered still used up the
// turn, as printed. The check is at trigger time only: it is not an
// intervening if, so it is not asked again as the ability resolves.
func TriggersOnlyOnceEachTurn(t game.TriggeredAbility) game.TriggeredAbility {
	when, label := t.AppliesTo, t.Key
	t.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		if source == nil || b11TriggeredThisTurn(g, source.InstanceID, label) {
			return false
		}
		return when == nil || when(ev, source, ch, g)
	}
	return t
}

// milledThisTurn reports whether `cardID`'s most recent arrival in a
// graveyard this turn was a MILL (CR 701.17a) — "a card in a graveyard
// that was milled this turn" (Tato Farmer, The Master, Transcendent).
// discardedOrMilledThisTurn's walk of this turn's events
// (graveyard_provenance.go) with only an EventMill counting: a card that
// was surveilled, Entombed, discarded or died is not one, a later
// departure from the graveyard ends the arrival (CR 400.7), and a card
// milled, returned and put back some other way was put there by the
// later move. False for a card already there when the turn began.
//
// Caller holds g.mu.
func milledThisTurn(g *game.Game, cardID uuid.UUID) bool {
	inArrival, milled := false, false
	for _, ev := range g.EventsThisTurn() {
		if ev.CardID != cardID {
			continue
		}
		switch {
		case ev.NewZone == game.ZoneGraveyard && ev.OldZone != game.ZoneGraveyard:
			if !inArrival {
				inArrival, milled = true, false
			}
			if ev.Kind == game.EventMill {
				milled = true
			}
		case ev.OldZone == game.ZoneGraveyard && ev.NewZone != "" && ev.NewZone != game.ZoneGraveyard:
			inArrival, milled = false, false
		}
	}
	return inArrival && milled
}

// MilledThisTurn is the target predicate over milledThisTurn: "target …
// card in a graveyard that was milled this turn". Judged at announce and
// again as the ability resolves (CR 608.2b), so a card that left the
// graveyard in response and came back is no longer the milled card.
func MilledThisTurn() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return milledThisTurn(g, c.InstanceID)
	}
}

// thisCardIsStillInItsGraveyard is a graveyard ability's intervening
// "if this card is in your graveyard" asked again as it resolves (CR
// 603.4): the source is in a graveyard and is the same object the
// ability triggered from (StackItem.SourceObject, CR 400.7), so a card
// that left in response and came back is a new object and the ability
// does nothing. An item with no stamp only asks the zone.
func thisCardIsStillInItsGraveyard(g *game.Game, item *game.StackItem) bool {
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return false
	}
	if item.SourceObject.ID != item.SourceCardID {
		return true
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	return ok && c.ObjectEpoch == item.SourceObject.Epoch
}

// CondThisInYourGraveyard is "if this creature is in your graveyard"
// (Infesting Radroach) declared as an intervening if (InterveningIf), so
// an optional graveyard trigger whose card has left is removed as it
// resolves without asking its "you may" (CR 603.4, ADR 0197). A card's
// graveyard is always its owner's, which is "your" for an ability that
// works there (CR 108.4a). It reads the source (FootprintCondSource).
type CondThisInYourGraveyard struct{}

func (CondThisInYourGraveyard) Holds(ctx *Context) bool {
	return ctx.Item != nil && thisCardIsStillInItsGraveyard(ctx.Game, ctx.Item)
}

func (CondThisInYourGraveyard) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondSource, 0), true
}

// youGetRadCounters is "you get N rad counters" as a Do step (Tato
// Farmer, Strong, the Brutish Thespian): the resolving item's controller
// gives them to themself, the giver every rad-counter gift names.
type youGetRadCounters struct{ N int }

func (s youGetRadCounters) Apply(ctx *Context) error {
	return playerGetsRadCounters(ctx.Game, ctx.Controller(), ctx.Controller(), s.N)
}
