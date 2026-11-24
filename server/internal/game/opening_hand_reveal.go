package game

import (
	"fmt"

	"github.com/google/uuid"
)

// opening_hand_reveal.go — the reveal shape of CR 103.6 (#6101, ADR 0133
// amendment 2026-11-24): "You may reveal this card from your opening hand.
// If you do, at the beginning of your first upkeep, …" (Devourer of
// Destiny).
//
// It is asked at the same moment and in the same order as the entry shape
// (offerOpeningHandActionsLocked), as a private keyed confirm. A yes
// reveals the card to the table (CR 701.20a; it stays in its owner's
// hand) and arms a delayed trigger, controlled by that player, for the
// beginning of their first upkeep (CR 603.7a). A no does nothing.
//
// # The starting player's first upkeep
//
// ADR 0133 decision 3 runs the first step's entry hooks while the
// opening-hand prompts are still open: by the time the starting player
// answers, turn one's upkeep has already begun and its delayed triggers
// have already been read. In paper the reveal comes before the game
// begins, so the trigger exists when that upkeep begins and triggers
// then. The answer therefore puts the trigger straight onto
// PendingTriggers when the revealing player is the active player on the
// game's first turn and the untap step is over: no player has received
// priority yet (the table is parked on the prompts), so it reaches the
// stack exactly where a trigger at the beginning of that upkeep would
// (CR 603.3). Every other seat's first upkeep is still to come, and an
// ordinary "at the beginning of your next upkeep" delayed trigger is it.

// OpeningHandReveal is what a revealed card arms: a delayed trigger at the
// beginning of its owner's first upkeep. Plain data, like the rest of
// OpeningHandAction: the body is a registered key, so the catalog entry
// gains no closure route and the armed trigger is a restore point.
type OpeningHandReveal struct {
	// Label is the delayed trigger's stack copy, "<card> — <what
	// happens>".
	Label string
	// Body is what the trigger does when it resolves.
	Body BodyRef
}

// openingHandRevealQuestion is the reveal prompt's header.
func openingHandRevealQuestion(name string) string {
	return fmt.Sprintf("Reveal %s from your opening hand?", name)
}

// queueOpeningHandRevealOfferLocked queues one card's reveal question,
// private to its chooser for the reason queueOpeningHandOfferLocked's is:
// declining reveals nothing, so the table must not learn the card from the
// question.
//
// Caller must hold g.mu.
func (g *Game) queueOpeningHandRevealOfferLocked(seat, cardID uuid.UUID, name string) {
	id, err := g.QueueConfirmKeyedForEffect(ConfirmPrompt{
		Chooser:      seat,
		Source:       cardID,
		Question:     openingHandRevealQuestion(name),
		AcceptLabel:  "Reveal it",
		DeclineLabel: "Keep it hidden",
	}, openingHandRevealAnswer, nil)
	if err != nil {
		return
	}
	if _, c := g.findChoiceLocked(id); c != nil {
		c.private = true
	}
}

// openingHandRevealAnswer is the answer to a reveal offer. An on-disk key:
// never renamed, never reused. The card is the prompt's source and must
// still be in the chooser's hand; its action is read from the catalog
// again, as the entry answer reads it.
var openingHandRevealAnswer = RegisterConfirmThen("confirm/opening-hand-reveal", func(g *Game, r ConfirmAnswer) error {
	if !r.Accepted || CatalogOpeningHand == nil {
		return nil
	}
	p := g.playerByIDLocked(r.Chooser)
	if p == nil || p.Hand == nil {
		return nil
	}
	c, ok := g.cardInZoneLocked(p.Hand, r.Source)
	if !ok {
		return nil
	}
	act := CatalogOpeningHand(catalogKeyOf(&c))
	if act == nil || act.Reveal == nil {
		return nil
	}
	g.RevealForEffect(RevealSpec{
		Player: r.Chooser,
		Source: r.Source,
		Reason: c.Name + " — revealed from an opening hand",
		Cards:  []uuid.UUID{r.Source},
	})
	g.armOpeningHandRevealLocked(r.Chooser, r.Source, act.Reveal)
	return nil
})

// armOpeningHandRevealLocked schedules the revealed card's "at the
// beginning of your first upkeep" trigger, or, when that upkeep has
// already begun (the starting player, see the file comment), puts it on
// PendingTriggers now.
//
// Caller must hold g.mu.
func (g *Game) armOpeningHandRevealLocked(seat, cardID uuid.UUID, rv *OpeningHandReveal) {
	dt := DelayedTrigger{
		Controller:         seat,
		SourceCardID:       cardID,
		Label:              rv.Label,
		At:                 StepUpkeep,
		ControllerTurnOnly: true,
		Body:               rv.Body,
	}
	if g.Turn.Seq <= 1 && g.activePlayerIDLocked() == seat && g.Turn.Step != StepUntap {
		dt.ID = uuid.New()
		dt.CreatedSeq = g.Turn.Seq
		dt.SourceObject = g.delayedSourceObjectLocked(cardID)
		dt.SourceTurnOvers = g.TurnOversForEffect(cardID)
		g.queueHarvestedTriggerLocked(dt.stackItem())
		return
	}
	g.ScheduleDelayedTriggerForEffect(dt)
}
