package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necropotence — Enchantment {B}{B}{B} (EDHREC rank 510):
//
//	"Skip your draw step.
//	 Whenever you discard a card, exile that card from your graveyard.
//	 Pay 1 life: Exile the top card of your library face down. Put
//	 that card into your hand at the beginning of your next end step."
//
// The second exit criterion of S22 and the reason the sprint needed
// delayed triggers at all. All three clauses are implemented.
//
// The card is three separate rules objects that happen to share a
// piece of cardboard, and modelling them as one would get each of
// them wrong:
//
//   - **The skip is a replacement effect**, not a flag on the player
//     (CR 614.10). It exists while the enchantment is on the
//     battlefield and stops the moment it leaves — a Necropotence
//     destroyed in your upkeep gives you the draw step it was going
//     to eat. See SkipYourDrawStep.
//
//   - **The discard clause is a trigger, not a replacement.** The
//     card really does reach the graveyard first, so anything
//     watching for a card hitting a graveyard sees it, and the
//     exile happens later, off the stack, in response to which a
//     player can flash the card back. Writing it as "discard
//     straight to exile" would be a strictly different card.
//
//   - **The pay-1-life ability is not a draw.** It exiles face down,
//     so the controller does not learn what they bought until the
//     end step, and an effect that counts cards drawn this turn
//     never counts it. That is also why running the library out to
//     it does not lose the game: CR 704.5b kills you for drawing
//     from an empty library, and this is not a draw. The engine's
//     ExileTopFaceDownForEffect is the only exile in the codebase
//     that does not mark the table as knowers.
//
// The CR 603.7 delayed trigger is scheduled once per activation and
// carries the exiled card as its payload, so five activations in one
// turn are five triggers that all fire at the same end step. It is
// ControllerTurnOnly because the printed text says "YOUR next end
// step": a Necropotence activated on an opponent's turn holds the
// card until your own end step comes round, which is the difference
// between the current Oracle text and the original printing.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "94a844d2-0574-45a7-b347-e0e329767c42",
		Name:         "Necropotence",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SkipYourDrawStep()},
		Triggered: []game.TriggeredAbility{{
			// "Whenever you discard a card" — EventDiscardCard is
			// emitted after the card is already in the graveyard,
			// with Actor set to the discarding player, which is
			// exactly the shape this clause wants: it names a card
			// that is in a graveyard by the time the trigger exists.
			Watches: []game.EventKind{game.EventDiscardCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID != uuid.Nil && ev.Actor == source.Controller
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				discarded := ev.CardID
				return game.NewTriggeredItem(source, "Necropotence — exile the discarded card",
					func(g *game.Game, item *game.StackItem) error {
						return necropotenceExileDiscarded(g, item.Controller, discarded)
					})
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "Pay 1 life: Exile the top card of your library face down. Put that card into your hand at the beginning of your next end step.",
			Cost:  PayLife(1),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				var exiled []uuid.UUID
				if err := (ExileTopFaceDown{
					Player: item.Controller,
					N:      1,
					Exiled: &exiled,
				}).Apply(ctx); err != nil {
					return err
				}
				if len(exiled) == 0 {
					// Empty library. The life is still paid (it was
					// a cost, paid at announce) and there is nothing
					// to schedule a delivery for.
					return nil
				}
				return ScheduleDelayedTrigger{
					At:                 game.StepEnd,
					ControllerTurnOnly: true,
					Label:              "Necropotence — put the exiled card into your hand",
					Cards:              exiled,
					Body:               necropotenceDeliverBody,
				}.Apply(ctx)
			},
		}},
	})
}

// necropotenceExileDiscarded exiles a just-discarded card from its
// owner's graveyard.
//
// Both guards are load-bearing rather than defensive. The card may
// have moved on between the discard and this trigger resolving —
// flashed back, reanimated, exiled by something else — and CR 400.7
// makes that a different object, so "exile that card" has nothing to
// act on. And the printed text says "from YOUR graveyard": a discard
// whose card ended up in some other player's graveyard (a
// control-changed card returning to its owner) is not the clause's
// business.
func necropotenceExileDiscarded(g *game.Game, controller, cardID uuid.UUID) error {
	if cardID == uuid.Nil {
		return nil
	}
	z := g.FindCardZoneForEffect(cardID)
	if z == nil || z.Kind != game.ZoneGraveyard || z.Owner != controller {
		return nil
	}
	return g.ExileCardForEffect(cardID)
}

// necropotenceDeliverExiled is the delayed trigger's effect: put the
// cards it carries into their owner's hand.
//
// A package-level func with no captures, per the delayed-trigger
// contract — the queue survives Clone / undo by sharing this
// function pointer with the snapshot, so everything it needs has to
// arrive on `item`. The payload rides as TargetCard refs.
//
// A card no longer in exile is skipped rather than dragged back from
// wherever it went (CR 608.2b). That is the case where an opponent
// answered the exile — a Scavenger Grounds-style graveyard/exile
// sweep, another Necropotence-shaped effect moving it on — and the
// instruction does nothing rather than reaching into a zone it was
// never pointed at.
func necropotenceDeliverExiled(g *game.Game, item *game.StackItem) error {
	for _, ref := range item.Targets {
		if ref.Kind != game.TargetCard || ref.ID == uuid.Nil {
			continue
		}
		z := g.FindCardZoneForEffect(ref.ID)
		if z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if err := g.BounceToHandForEffect(ref.ID); err != nil {
			return err
		}
	}
	return nil
}
