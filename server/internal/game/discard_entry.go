package game

import "github.com/google/uuid"

// discard_entry.go — a discard that a replacement settles on the
// battlefield (ADR 0153 §2 and §3, #1861).
//
// Obstinate Baloth: "If a spell or ability an opponent controls causes
// you to discard Obstinate Baloth, put it onto the battlefield instead
// of putting it into your graveyard." The discard window
// (RepEventDiscard) rewrites the destination to ZoneBattlefield, and
// executeZoneRouteLocked hands the settled event here instead of moving
// the card, because an exit's MoveCard would skip everything a
// battlefield entry owes: its own CR 614 window (CR 614.12: enters
// tapped, enters with counters, "exile it instead"), the controller
// stamp, summoning sickness, EventETB, the AsEnters hook and the entry
// ordinal.
//
// So the card enters through the one entry primitive every other "put
// it from your hand onto the battlefield" uses, under its OWNER's
// control (CR 110.2: the discard replacement is the owner's, CR 616.1,
// and the card is in the owner's hand). That is a second, chained
// CR 614 window: the discard window has settled and cannot see an
// entry, and the entry replacements watch a battlefield RepEventMove
// that no discard replacement watches, so nothing applies twice
// (CR 614.5). ADR 0153 §3 says why two windows are rules-equivalent to
// one here.
//
// # Order and the batch
//
// The entry can PAUSE (a shockland-style question, a CR 616 ordering
// prompt). The card is still in hand while it does, and the rest of the
// discard batch, and through it the discard effect's own "then", waits:
// the route's continuation is detached and handed to the entry's, which
// runs exactly once at the entry's terminal outcome (entered,
// cancelled, redirected, or its prompt taken away). It is the contract
// a paused CR 903.9 exit already gives the batch (#853).
//
// EventDiscardCard is emitted from that same continuation, AFTER
// EventETB, with NewZone set to where the card really is (owner
// decision 3). Both are one event under CR 614.6 and their triggers are
// put on the stack together (CR 603.3), so the log order changes
// nothing a card reads, and emitting late is what lets NewZone be true
// when the entry window redirected the card (ADR 0098 Decision 4: it
// lands in exile, and the discard still happened).
//
// "Discarded this way" needs no help: discardedThisWayLocked reads "no
// longer in the hand", so a card on the battlefield or in exile counts
// and a card the entry window cancelled outright, still in hand, does
// not.

// discardOntoBattlefieldLocked is executeZoneRouteLocked's arm for a
// discard settled on the battlefield. ev is the settled RepEventDiscard
// and carries the route.
//
// Caller must hold g.mu.
func (g *Game) discardOntoBattlefieldLocked(ev *ReplacementEvent) error {
	r := ev.zoneRoute
	if legs := r.discardLegs; legs != nil && r.then == nil {
		// #4253 slice 9: a keyed discard batch hands its rest to the
		// entry as data (discardEntryRun), so the entry's prompt is a
		// restore point too.
		r.discardLegs = nil
		return g.discardOntoBattlefieldKeyedLocked(ev, *r, legs)
	}
	// Detach the rest of the batch from the route. routeCardToZoneLocked
	// runs the route's tail when this returns unpaused, and an entry
	// that pauses returns unpaused from here, so the tail has to belong
	// to the entry's continuation instead. Cleared through the pointer,
	// as runRouteTailLocked does; an undo snapshot holds its own copy of
	// the route (cloneReplacementResume).
	tail := r.then
	r.then = nil
	route := *r
	finish := func(g *Game) error {
		if tail == nil {
			return nil
		}
		return tail(g)
	}

	player := ev.DiscardPlayer
	if player == uuid.Nil {
		player = r.Actor
	}
	p := g.playerByIDLocked(player)
	if p == nil || p.Hand == nil || !p.Hand.Contains(ev.CardID) {
		// The card left the hand before the handoff ran: nothing moves,
		// no discard happened (CR 701.9a), and the batch carries on, as
		// it does for a stale exit.
		return finish(g)
	}

	cardID := ev.CardID
	discard := Event{
		Kind:         EventDiscardCard,
		Actor:        player,
		Source:       route.Source,
		CardID:       cardID,
		OldZone:      ZoneHand,
		DiscardCause: ev.DiscardCause,
	}
	if route.Actor != uuid.Nil {
		discard.Actor = route.Actor
	}
	route.Cause.stampCause(&discard)

	opts := HandEntryOptions{
		Controller:         player,
		EntersWithCounters: copyCounterMap(ev.EntersWithCounters),
	}
	return g.PutFromHandOntoBattlefieldThenForEffect(cardID, opts, func(g *Game, _ uuid.UUID) error {
		// Where the card is now. A hand card keeps its ID through the
		// entry and through a redirect out of the entry window, so the
		// live zone is the answer. Still in hand: the entry window
		// cancelled it, the card never left, and there was no discard.
		if z := g.findCardZoneLocked(cardID); z != nil && z.Kind != ZoneHand {
			out := discard
			out.NewZone = z.Kind
			g.EmitEvent(out)
		}
		return finish(g)
	})
}

// DiscardCausedByOpponent reports whether a discard replacement event
// was caused by a spell or ability an opponent of the discarding player
// controls — the Obstinate Baloth family's condition (ADR 0153 §1). A
// cost, the cleanup step and a manual move have no cause controller and
// are never an opponent's. Every other player at a Commander table is
// an opponent (CR 102.3).
func DiscardCausedByOpponent(ev *ReplacementEvent) bool {
	return ev != nil && ev.Kind == RepEventDiscard &&
		ev.DiscardCauseController != uuid.Nil &&
		ev.DiscardCauseController != ev.DiscardPlayer
}

// discardEntryRun is a keyed discard batch's leg that became an entry
// (#4253 slice 9): the card, what its EventDiscardCard names, and the rest
// of the batch. It rides the entry batch (entryBatch.discard), which runs
// it at the entry's terminal outcome, as the closure form's `finish` did.
type discardEntryRun struct {
	Card            uuid.UUID        `json:"card"`
	Actor           uuid.UUID        `json:"actor"`
	DiscardCause    DiscardCause     `json:"discardCause,omitempty"`
	CauseKind       MoveCauseKind    `json:"causeKind,omitempty"`
	CauseController *uuid.UUID       `json:"causeController,omitempty"`
	CauseItem       *uuid.UUID       `json:"causeItem,omitempty"`
	Legs            *discardBatchRun `json:"legs"`
}

func cloneDiscardEntryRun(d *discardEntryRun) *discardEntryRun {
	if d == nil {
		return nil
	}
	out := *d
	out.Legs = cloneDiscardBatchRun(d.Legs)
	return &out
}

// discardOntoBattlefieldKeyedLocked is discardOntoBattlefieldLocked with
// the rest of the batch as data.
//
// Caller must hold g.mu.
func (g *Game) discardOntoBattlefieldKeyedLocked(ev *ReplacementEvent, route zoneRoute, legs *discardBatchRun) error {
	player := ev.DiscardPlayer
	if player == uuid.Nil {
		player = route.Actor
	}
	p := g.playerByIDLocked(player)
	if p == nil || p.Hand == nil || !p.Hand.Contains(ev.CardID) {
		return g.runDiscardBatchLegLocked(legs)
	}
	actor := player
	if route.Actor != uuid.Nil {
		actor = route.Actor
	}
	d := &discardEntryRun{
		Card: ev.CardID, Actor: actor, DiscardCause: ev.DiscardCause,
		CauseKind: route.Cause.Kind, CauseController: uuidPtrOrNil(route.Cause.Controller), CauseItem: uuidPtrOrNil(route.Cause.Item),
		Legs: legs,
	}
	opts := HandEntryOptions{
		Controller:         player,
		EntersWithCounters: copyCounterMap(ev.EntersWithCounters),
	}
	return g.startPreparedEntryBatchLocked([]BatchEntry{{CardID: ev.CardID, From: ZoneHand}}, &entryBatch{opts: opts, discard: d})
}

// finishDiscardEntryLocked is a keyed discard entry's terminal outcome:
// the discard event, naming where the card really is, then the rest of the
// batch. Still in hand: the entry window cancelled it and there was no
// discard.
//
// Caller must hold g.mu.
func (g *Game) finishDiscardEntryLocked(d *discardEntryRun) error {
	if z := g.findCardZoneLocked(d.Card); z != nil && z.Kind != ZoneHand {
		out := Event{
			Kind:         EventDiscardCard,
			Actor:        d.Actor,
			Source:       d.Legs.Source,
			CardID:       d.Card,
			OldZone:      ZoneHand,
			NewZone:      z.Kind,
			DiscardCause: d.DiscardCause,
		}
		MoveCause{Kind: d.CauseKind, Controller: uuidOrNil(d.CauseController), Item: uuidOrNil(d.CauseItem)}.stampCause(&out)
		g.EmitEvent(out)
	}
	return g.runDiscardBatchLegLocked(d.Legs)
}
