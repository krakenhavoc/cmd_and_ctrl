package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bag of Holding — Artifact {1}:
//
//	"Whenever you discard a card, exile that card from your graveyard.
//	 {2}, {T}: Draw a card, then discard a card.
//	 {4}, {T}, Sacrifice this artifact: Return all cards exiled with
//	 this artifact to their owner's hand."
//
// A one-mana looter that hands the whole pile back later. The three
// abilities are one machine: the trigger is a drawback on its own
// (your graveyard is emptied of everything you pitch, so no
// flashback, no delve, no reanimation off a discarded creature) and
// only pays out through the third ability.
//
// The record of WHAT was exiled with the Bag is the one part that is
// not a primitive. It is read back off the event log by
// b27ExiledWith, the same reader Duplicant, Angel of Serenity and the
// Oblivion Ring family use: a card counts as exiled with this Bag
// when its most recent move into exile happened while a resolution
// labelled bagOfHoldingExileLabel and sourced from this Bag was open.
// So the two abilities agree by construction — the label is a const
// rather than two string literals — and a card that has since left
// exile under some other effect is not pulled back out of wherever it
// went.
//
// The Bag is sacrificed as a COST of the third ability (CR 601.2h),
// so it is in the graveyard by the time that ability resolves. That
// costs the reader nothing: the stack item keeps the Bag's instance
// ID, and the record was written against that ID while the Bag was
// still on the battlefield.
//
// No simplification.

// bagOfHoldingExileLabel is the stack label the discard trigger
// carries. b27ExiledWith keys the "exiled with this artifact" record
// on it, so the trigger and the sacrifice ability must agree — a
// const rather than two string literals for exactly that reason.
const bagOfHoldingExileLabel = "Bag of Holding — exile the discarded card from your graveyard"

func init() {
	Register(Spec{
		OracleID:     "63c04040-e109-494e-baa6-c639a6c9a996",
		Name:         "Bag of Holding",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDiscardCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				discarded := ev.CardID
				return game.NewTriggeredItem(source, bagOfHoldingExileLabel,
					func(g *game.Game, item *game.StackItem) error {
						return bagOfHoldingExileFromGraveyard(g, item, discarded)
					})
			},
		}},
		Activated: []ActivatedAbility{
			{
				Label: "{2}, {T}: Draw a card, then discard a card.",
				Cost:  Plus(ManaCost("{2}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b16DrawThenDiscard(g, item, 1, 1)
				},
			},
			{
				Label:  "{4}, {T}, Sacrifice this artifact: Return all cards exiled with this artifact to their owner's hand.",
				Cost:   Plus(ManaCost("{4}"), TapCost(), SacrificeThis()),
				Effect: bagOfHoldingReturnExiledCards,
			},
		},
	})
}

// bagOfHoldingExileFromGraveyard is the trigger's body: exile the
// discarded card, but only if it is STILL in a graveyard. A discard
// answered in response with a reanimation, or a madness cast, leaves
// nothing to exile and the trigger does nothing — "exile that card
// from your graveyard" names an object in one zone (CR 400.7).
func bagOfHoldingExileFromGraveyard(g *game.Game, item *game.StackItem, cardID uuid.UUID) error {
	if z := g.FindCardZoneForEffect(cardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ExileTarget{Target: cardID}.Apply(NewContext(g, item))
}

// bagOfHoldingReturnExiledCards is the sacrifice ability's body:
// every card still in exile that this Bag's trigger put there returns
// to its OWNER's hand (CR 610.3 — the card goes to the player who
// owned it, not to whoever exiled it, which matters for a card
// discarded off an opponent's wheel).
func bagOfHoldingReturnExiledCards(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range b27ExiledWith(g, item.SourceCardID, bagOfHoldingExileLabel) {
		if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
