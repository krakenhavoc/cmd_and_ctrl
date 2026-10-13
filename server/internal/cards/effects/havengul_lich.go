package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Havengul Lich — Creature — Zombie Wizard {3}{U}{B}, 4/4:
//
//	"{1}: You may cast target creature card in a graveyard this turn.
//	 When you cast it this turn, this creature gains all activated
//	 abilities of that card until end of turn."
//
// Two halves, both made as the ability resolves:
//
//   - The permission is a ScopeCards cast permission over the target, as
//     the object it is now (Locke's shape): cast only, until end of turn,
//     at the card's normal timing and cost, from whichever graveyard it is
//     in. It is the activator's, so losing the Lich does not take it away.
//   - "When you cast it this turn" is a delayed trigger (CR 603.7) watching
//     for that card to be cast by you from a graveyard, by this permission
//     or any other, once. It ends with the turn.
//
// When it resolves, the Lich gains the rows the card prints, read off the
// card as it was cast (ADR 0155 Decision 9): the spell may have resolved
// or been countered by then, and "that card" is still what it lends. The
// borrow lasts until end of turn and survives the card leaving (CR
// 608.2h). A card the engine does not automate lends nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1fb02bad-0d1f-437f-a269-f51cc19b280f",
		Name:         "Havengul Lich",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}: You may cast target creature card in a graveyard this turn. When you cast it this turn, this creature gains all activated abilities of that card until end of turn.",
			Cost:    ManaCost("{1}"),
			Targets: TargetCardInGraveyard("target creature card in a graveyard", Creature()),
			Effect:  havengulLichPermission,
		}},
	})
}

var (
	// "When you cast it this turn": an EventCast by the trigger's
	// controller of the card the trigger lists, out of a graveyard.
	havengulLichCastItCondition = game.DelayedCondition("havengul-lich/you-cast-it", havengulLichCastIt)

	// "This creature gains all activated abilities of that card until end
	// of turn."
	havengulLichBorrowBody = game.SimpleDelayedBody("havengul-lich/gains-its-abilities", havengulLichBorrow)
)

// havengulLichPermission grants the cast and schedules the trigger.
func havengulLichPermission(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		card, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:     item.Controller,
			Zone:       game.ZoneGraveyard,
			CastOnly:   true,
			Source:     item.SourceCardID,
			SourceName: "Havengul Lich",
			Label:      "Havengul Lich — cast " + card.Name + " this turn",
		}, []game.Card{card})
		return DelayedOnEvent{
			Label:     "Havengul Lich — gains all activated abilities of " + card.Name + " until end of turn",
			On:        []game.EventKind{game.EventCast},
			Condition: havengulLichCastItCondition,
			Cards:     []uuid.UUID{card.InstanceID},
			Body:      havengulLichBorrowBody,
		}.Apply(ctx)
	}
	return nil
}

// havengulLichCastIt is the delayed trigger's condition: you cast the
// listed card, and it came from a graveyard (CR 601.2a: the event's
// OldZone is where it was cast from).
func havengulLichCastIt(ev game.Event, dt *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
	return ev.Kind == game.EventCast && ev.Actor == dt.Controller && ev.OldZone == game.ZoneGraveyard &&
		len(dt.Cards) > 0 && ev.CardID == dt.Cards[0]
}

// havengulLichBorrow lends the cast card's printed activated abilities to
// the Lich until end of turn.
func havengulLichBorrow(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cast := ctx.PayloadCards()
	if len(cast) == 0 {
		return nil
	}
	card, ok := g.LookupCardForEffect(cast[0])
	if !ok {
		return nil
	}
	return BorrowAbilitiesFor{
		Target:  item.SourceCardID,
		Lender:  &card,
		Printed: true,
		Label:   "Havengul Lich — gains all activated abilities of " + card.Name + " until end of turn",
	}.Apply(ctx)
}
