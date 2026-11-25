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
//
// # The other moments (OpeningHandReveal.At)
//
// The Chancellors of New Phyrexia name three more (#6101):
//
//   - RevealAtTheFirstUpkeep, "at the beginning of the first upkeep"
//     (Chancellor of the Dross, of the Forge, of the Spires): the game's
//     first upkeep, the starting player's, whoever revealed. It has begun
//     by the time any seat answers, so the trigger goes straight onto
//     PendingTriggers for every seat, as the starting player's does above.
//   - RevealAtYourFirstMainPhase, "at the beginning of your first main
//     phase of the game" (Chancellor of the Tangle): a delayed trigger at
//     the beginning of the revealing player's next precombat main phase,
//     which is still to come for every seat, the starting player's
//     included (the table is in turn one's upkeep).
//   - RevealAtEachOpponentsFirstSpell, "when each opponent casts their
//     first spell of the game" (Chancellor of the Annex): one delayed
//     trigger per opponent, waiting with no duration for that opponent's
//     next cast (CR 603.7b: it triggers only once). No player can cast a
//     spell before the reveal (nobody has had priority yet), so that next
//     cast is their first spell of the game. The cast spell rides as the
//     fired item's Payload, which is "that spell".

// OpeningHandReveal is what a revealed card arms: a delayed trigger at the
// beginning of its owner's first upkeep. Plain data, like the rest of
// OpeningHandAction: the body is a registered key, so the catalog entry
// gains no closure route and the armed trigger is a restore point.
type OpeningHandReveal struct {
	// At is the moment the trigger is for. The zero value is "at the
	// beginning of your first upkeep".
	At OpeningHandRevealAt
	// Label is the delayed trigger's stack copy, "<card> — <what
	// happens>".
	Label string
	// Body is what the trigger does when it resolves.
	Body BodyRef
}

// OpeningHandRevealAt names the moment an opening-hand reveal's trigger is
// for (see the file comment). Catalog data, never persisted: what is armed
// is an ordinary DelayedTrigger.
type OpeningHandRevealAt string

const (
	// RevealAtYourFirstUpkeep is "at the beginning of your first upkeep"
	// (Devourer of Destiny).
	RevealAtYourFirstUpkeep OpeningHandRevealAt = ""
	// RevealAtTheFirstUpkeep is "at the beginning of the first upkeep"
	// (Chancellor of the Dross).
	RevealAtTheFirstUpkeep OpeningHandRevealAt = "the-first-upkeep"
	// RevealAtYourFirstMainPhase is "at the beginning of your first main
	// phase of the game" (Chancellor of the Tangle).
	RevealAtYourFirstMainPhase OpeningHandRevealAt = "your-first-main-phase"
	// RevealAtEachOpponentsFirstSpell is "when each opponent casts their
	// first spell of the game" (Chancellor of the Annex).
	RevealAtEachOpponentsFirstSpell OpeningHandRevealAt = "each-opponents-first-spell"
)

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

// armOpeningHandRevealLocked arms the revealed card's trigger for the
// moment its reveal names (see the file comment). A first upkeep that has
// already begun gets the trigger on PendingTriggers now.
//
// Caller must hold g.mu.
func (g *Game) armOpeningHandRevealLocked(seat, cardID uuid.UUID, rv *OpeningHandReveal) {
	dt := DelayedTrigger{
		Controller:   seat,
		SourceCardID: cardID,
		Label:        rv.Label,
		Body:         rv.Body,
	}
	firstUpkeepBegun := g.Turn.Seq <= 1 && g.Turn.Step != StepUntap
	switch rv.At {
	case RevealAtTheFirstUpkeep:
		if g.Turn.Seq > 1 {
			return
		}
		dt.At = StepUpkeep
		if firstUpkeepBegun {
			g.queueOpeningHandRevealNowLocked(dt)
			return
		}
	case RevealAtYourFirstMainPhase:
		dt.At = StepPrecombatMain
		dt.ControllerTurnOnly = true
	case RevealAtEachOpponentsFirstSpell:
		dt.On = []EventKind{EventCast}
		dt.Condition = openingHandRevealThatPlayersNextSpell
		forever := IndefiniteDuration()
		dt.Duration = &forever
		for _, opp := range g.Seats {
			if opp == nil || opp.ID == seat || opp.Eliminated {
				continue
			}
			each := dt
			each.CondParams = EffectParams{Player: opp.ID}
			g.ScheduleDelayedTriggerForEffect(each)
		}
		return
	default:
		dt.At = StepUpkeep
		dt.ControllerTurnOnly = true
		if firstUpkeepBegun && g.activePlayerIDLocked() == seat {
			g.queueOpeningHandRevealNowLocked(dt)
			return
		}
	}
	g.ScheduleDelayedTriggerForEffect(dt)
}

// queueOpeningHandRevealNowLocked puts a reveal's upkeep trigger straight
// onto PendingTriggers: the upkeep it is for began before the answer came.
//
// Caller must hold g.mu.
func (g *Game) queueOpeningHandRevealNowLocked(dt DelayedTrigger) {
	dt.ID = uuid.New()
	dt.CreatedSeq = g.Turn.Seq
	dt.SourceObject = g.delayedSourceObjectLocked(dt.SourceCardID)
	dt.SourceTurnOvers = g.TurnOversForEffect(dt.SourceCardID)
	g.queueHarvestedTriggerLocked(dt.stackItem())
}

// openingHandRevealThatPlayersNextSpell is the event condition of "when
// each opponent casts their first spell of the game": a cast by the player
// CondParams names. An on-disk key: never renamed, never reused.
var openingHandRevealThatPlayersNextSpell = DelayedCondition("cast/opening-hand-reveal-that-players-first-spell",
	func(ev Event, _ *DelayedTrigger, _ *Game, p EffectParams) bool {
		return ev.Kind == EventCast && p.Player != uuid.Nil && ev.Actor == p.Player
	})
