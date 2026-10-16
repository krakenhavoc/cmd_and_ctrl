package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_by_opponent_bodies.go — the delayed-trigger bodies and the one
// event condition ADR 0153's "this card" triggers and Pure Intentions
// need. They are keyed, append-only effect identities like every body
// in delayed_bodies.go (the ledger is
// server/internal/game/testdata/effect_keys.txt); they live in their own
// file so a concurrent card batch does not collide on that one.

var (
	// "Return this card from your graveyard to your hand at the beginning
	// of the next end step" (Mangara's Blessing, Pure Intentions). The
	// card is pinned as an OBJECT (CR 400.7): a card that has left the
	// graveyard and come back is a new object and is left alone.
	returnDiscardedToHandBody = game.DelayedBody("discard-by-opponent/return-to-hand",
		returnTheObjectFromGraveyardToHand)

	// "Return this card from your graveyard to the battlefield with a
	// +1/+1 counter on it at the beginning of the next end step" (Sand
	// Golem). The counter is a CR 614.1c enters-with clause.
	returnDiscardedWithCounterBody = game.DelayedBody("discard-by-opponent/return-with-counter",
		returnTheObjectFromGraveyardWithACounter)

	// Pure Intentions' "return those cards from your graveyard to your
	// hand": the card the discard event named rides on the item as its
	// payload.
	returnTheDiscardedCardBody = game.SimpleDelayedBody("discard-by-opponent/return-the-discarded-card",
		returnThePayloadCardFromGraveyardToHand)

	// The condition of Pure Intentions' "whenever a spell or ability an
	// opponent controls causes you to discard cards this turn": a discard
	// by the delayed trigger's controller that a resolving spell or
	// ability an opponent controls caused.
	opponentCausedYourDiscardCondition = game.DelayedCondition("discard/opponent-caused-yours",
		opponentCausedYourDiscard)
)

// returnTheObjectFromGraveyardToHand is returnDiscardedToHandBody.
func returnTheObjectFromGraveyardToHand(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if !objectStillInAGraveyard(g, p.Object) {
		return nil
	}
	return ReturnFromGraveyard{Target: p.Object.ID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
}

// returnTheObjectFromGraveyardWithACounter is returnDiscardedWithCounterBody:
// the card returns under its owner's control with one +1/+1 counter.
func returnTheObjectFromGraveyardWithACounter(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if !objectStillInAGraveyard(g, p.Object) {
		return nil
	}
	_, err := g.ReturnFromGraveyardWithCountersForEffect(p.Object.ID, item.Controller, false,
		map[string]int{game.CounterPlusOne: 1})
	if errors.Is(err, game.ErrCardNotFound) {
		return nil
	}
	return err
}

// returnThePayloadCardFromGraveyardToHand is returnTheDiscardedCardBody.
func returnThePayloadCardFromGraveyardToHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range ctx.PayloadCards() {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// objectStillInAGraveyard reports whether the card an ObjectRef names is
// still that object, in a graveyard.
func objectStillInAGraveyard(g *game.Game, o game.ObjectRef) bool {
	c, ok := g.LookupCardForEffect(o.ID)
	if !ok || c.ObjectEpoch != o.Epoch {
		return false
	}
	z := g.FindCardZoneForEffect(o.ID)
	return z != nil && z.Kind == game.ZoneGraveyard
}

// opponentCausedYourDiscard is opponentCausedYourDiscardCondition.
func opponentCausedYourDiscard(ev game.Event, dt *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
	return game.DiscardCausedByOpponentOf(ev, dt.Controller)
}

// scheduleReturnOfThisCard schedules a "return this card from your
// graveyard … at the beginning of the next end step" for the card a
// discard trigger is about, if it is still in a graveyard as the trigger
// resolves.
func scheduleReturnOfThisCard(g *game.Game, item *game.StackItem, label string, body game.BodyRef) error {
	id := item.SourceCardID
	if g.AbilitySourceLeftBattlefieldForEffect(item) {
		// The card is a different object than the one that was discarded
		// (CR 400.7): madness's "if you don't cast it, put it into your
		// graveyard" exiled and then buried it, so there is no card in the
		// graveyard to follow.
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label:  label,
		Cards:  []uuid.UUID{id},
		Body:   body,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(NewContext(g, item))
}

// returnDiscardedCardToHand is Quagnoth's "return it to your hand": the
// discarded card, as the object that was discarded (CR 400.7), comes back
// from the graveyard, exile (madness, Rest in Peace), the top of a library
// (Library of Leng) or the command zone, wherever the discard left it.
func returnDiscardedCardToHand(g *game.Game, item *game.StackItem) error {
	if g.AbilitySourceLeftBattlefieldForEffect(item) {
		return nil
	}
	id := item.SourceCardID
	if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneGraveyard {
		return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(NewContext(g, item))
	}
	_, err := g.ReturnFromAnyZoneToHandForEffect(item.Controller, id, id)
	return err
}
