package game

import "github.com/google/uuid"

// delayed.go — S22: CR 603.7 delayed triggered abilities.
//
// "At the beginning of the next end step, return that card to the
// battlefield" is not a triggered ability on a permanent. It is a
// one-shot instruction created by a resolving spell or ability that
// sits outside every zone until its condition is met, fires exactly
// once, and then ceases to exist. `TriggeredAbility` cannot express
// it: that shape is declared statically on a catalog Spec and
// harvested by walking the battlefield, so it needs a source card in
// play — and the whole point of a delayed trigger is that the thing
// that created it is usually gone (an instant in the graveyard, a
// creature that left).
//
// So delayed triggers live on the Game as their own queue:
//
//	spell / ability resolves → ScheduleDelayedTriggerForEffect
//	  → g.DelayedTriggers  (dormant, survives Clone / RestoreFrom)
//	  → step entry matches DelayedTrigger.At
//	    → fireDelayedTriggersLocked builds a StackItem
//	    → PendingTriggers → StackMeta → resolve
//
// Two properties that shape the design:
//
//   - **It uses the stack.** A delayed trigger is a triggered ability
//     (CR 603.7b), so it goes on the stack via the same
//     PendingTriggers queue the harvester feeds and every player gets
//     a response window before it resolves. Applying the effect
//     straight from the step hook would deny that, exactly as
//     applying an ordinary trigger from `Build` would.
//
//   - **"Next" is free.** The queue is drained on step ENTRY only, so
//     an ability created *during* an end step cannot fire in that same
//     end step — the entry hook for it has already run. It waits for
//     the next one, which is precisely CR 603.7b without any
//     "created this step" bookkeeping. That timing is observable: a
//     blink cast in an opponent's end step returns the permanent a
//     whole turn later, not moments afterwards.
//
// Serialisation / undo: the queue is deep-copied by cloneLocked and
// swapped wholesale by RestoreFrom, and it is projected onto the wire
// as GameView.DelayedTriggers so a client (and the replay log built
// from those snapshots) can see what is still owed. The Effect func
// is shared rather than copied on Clone, on exactly the contract
// StackItem.Effect already documents: it takes the live *Game and its
// own item at fire time and captures neither, so an item restored
// into a different *Game resolves against that one.

// DelayedTrigger is one pending "at the beginning of the next
// <step>, do X" instruction. Created by ScheduleDelayedTriggerForEffect
// from a resolving effect; removed from the queue when it fires.
type DelayedTrigger struct {
	// ID identifies this trigger in the queue and on the wire.
	// Minted by ScheduleDelayedTriggerForEffect when left Nil.
	ID uuid.UUID

	// Controller is the player who controls the delayed ability —
	// the controller of the spell or ability that created it (CR
	// 603.7d). They get the resulting stack item, and it is their
	// APNAP slot the trigger drains in.
	Controller uuid.UUID

	// SourceCardID is the card whose text created the trigger. Kept
	// for attribution in the stack overlay and the event log; the
	// trigger fires whether or not that card still exists, which is
	// the normal case (the instant that made it is in a graveyard).
	SourceCardID uuid.UUID

	// Label is the stack-overlay copy, in the same shape ordinary
	// triggers use: "Waterbender's Restoration — return the exiled
	// creatures".
	Label string

	// At is the step whose BEGINNING fires this trigger. StepEnd for
	// the overwhelmingly common "at the beginning of the next end
	// step"; StepUpkeep for rebound-shaped effects.
	At Step

	// ControllerTurnOnly restricts the trigger to a step of the
	// CONTROLLER's own turn: "at the beginning of YOUR next main
	// phase" (Mana Drain) rather than "the next turn's upkeep"
	// (Arcane Denial, which any player's upkeep satisfies). When set,
	// a matching step on another player's turn leaves the trigger
	// queued instead of firing it.
	ControllerTurnOnly bool

	// CreatedTurn is the turn number the trigger was scheduled on.
	// Not used for firing (see the "next is free" note above) —
	// it's there for the wire view and for debugging a queue that
	// somehow never drains.
	CreatedTurn int

	// Cards is the trigger's payload: the card instances the effect
	// acts on, captured when the trigger was created. Stamped onto
	// the fired StackItem's Targets as TargetCard refs, so the
	// Effect reads them off the item exactly as a targeted trigger
	// reads its chosen target — and cloneStackItem deep-copies them
	// for free.
	//
	// The CR 608.2b existence re-check applies to those refs, which
	// is the behaviour a return-from-exile wants: a card that is no
	// longer where the trigger left it does nothing rather than
	// erroring.
	Cards []uuid.UUID

	// Effect is what the trigger does when its stack item resolves.
	// Same contract as StackItem.Effect: it receives the live game
	// and the item, and MUST NOT capture a *Game or a pointer into
	// a zone slice — undo restores a cloned game and the closure has
	// to resolve against that one. Read the payload from
	// item.Targets and the controller from item.Controller.
	//
	// Runs under g.mu held in write mode.
	Effect func(g *Game, item *StackItem) error

	// On is the EVENT condition, the #663 alternative to At: "when
	// you NEXT CAST an instant or sorcery spell this turn, copy that
	// spell" (Doublecast). The kinds are the same cheap pre-filter
	// TriggeredAbility.Watches is, read by the one hook in
	// triggerHarvester.OnEvent.
	//
	// Empty means "step-conditioned", which is every trigger this
	// file held before #663. A trigger may carry both At and On;
	// whichever condition is met first fires it, and it fires once.
	// See the 2026-09-18 amendment to ADR 0026, which reverses §1-2
	// for this case and says why.
	On []EventKind

	// AppliesTo narrows On to the event the card actually names — the
	// delayed sibling of TriggeredAbility.AppliesTo, with a
	// DelayedTrigger in the source's place because there is no source
	// card to hand it. Nil means every event of a watched kind
	// matches.
	//
	// Runs under g.mu held in write mode. MUST NOT call public
	// locking mutators. A closure, so it does not survive a snapshot
	// — see the census note on Effect.
	AppliesTo func(ev Event, dt *DelayedTrigger, g *Game) bool

	// Optional is the CR 603.5 "you may" on a fired trigger, asked
	// through the harvester's own prompt because the dispatch is the
	// harvester's. Nil for the mandatory case, which is every card on
	// this seam today.
	Optional *TriggerOptionalPrompt

	// Duration is how long this trigger is owed for — the SAME CR
	// 611.2 duration model the scoped statics use (ADR 0063,
	// duration.go), because "this turn" means the same thing to a
	// delayed trigger as it does to a Giant Growth and one switch
	// should decide it. Plain data, no closures, so it clones and
	// snapshots verbatim.
	//
	// Nil means "no duration", which is what a step-conditioned
	// trigger wants: "at the beginning of the NEXT end step"
	// scheduled during an end step has to outlive this turn.
	// ScheduleDelayedTriggerForEffect stamps UntilEndOfTurn on an
	// event-conditioned trigger that names none, because every
	// printed one says "this turn".
	Duration *Duration
}

// ScheduleDelayedTriggerForEffect registers a delayed triggered
// ability. Returns the trigger's ID, or uuid.Nil when the request is
// malformed (no Effect, or no step to fire at) — a malformed trigger
// is dropped rather than queued, so it can never wedge the queue.
//
// Caller must hold g.mu. Deliberately emits no event: the common
// caller is a resolving effect that has already emitted its own
// breadcrumbs, and one caller (Cosmic Intervention) runs inside the
// CR 614 replacement pipeline, where an EmitEvent would re-enter the
// trigger harvester in the middle of replacing an event.
func (g *Game) ScheduleDelayedTriggerForEffect(dt DelayedTrigger) uuid.UUID {
	if dt.Effect == nil || (dt.At == "" && len(dt.On) == 0) {
		return uuid.Nil
	}
	if dt.ID == uuid.Nil {
		dt.ID = uuid.New()
	}
	dt.CreatedTurn = g.Turn.Number
	// #663: "this turn" is the printed duration of every
	// event-conditioned delayed trigger there is, and CR 514.2 ends
	// it at cleanup whether or not it fired. A caller that means
	// something longer hands over its own Duration.
	if len(dt.On) > 0 && dt.Duration == nil {
		d := g.UntilEndOfTurnDuration()
		dt.Duration = &d
	}
	queued := dt
	if len(dt.Cards) > 0 {
		queued.Cards = append([]uuid.UUID(nil), dt.Cards...)
	}
	g.DelayedTriggers = append(g.DelayedTriggers, &queued)
	return queued.ID
}

// fireDelayedTriggersLocked drains every queued trigger whose `At`
// matches `step` onto PendingTriggers, in scheduling order, and
// removes them from the queue.
//
// Called from runStepEntryHooksLocked on entry to each step, which
// is what gives "the NEXT end step" its meaning for free — a trigger
// scheduled during the current end step is queued after this hook has
// already run for that step, so it waits for the following one.
//
// The queue is rewritten BEFORE any item is enqueued: enqueuing emits
// EventTrigger, which runs the listener chain (including the trigger
// harvester), and a listener that looked at a queue still holding
// already-fired entries would see them twice.
//
// Caller must hold g.mu.
func (g *Game) fireDelayedTriggersLocked(step Step) {
	if len(g.DelayedTriggers) == 0 || step == "" {
		return
	}
	var keep, fire []*DelayedTrigger
	for _, dt := range g.DelayedTriggers {
		if dt == nil {
			continue
		}
		if dt.At == step && (!dt.ControllerTurnOnly || g.activePlayerIDLocked() == dt.Controller) {
			fire = append(fire, dt)
			continue
		}
		keep = append(keep, dt)
	}
	if len(fire) == 0 {
		return
	}
	g.DelayedTriggers = keep
	for _, dt := range fire {
		g.queueHarvestedTriggerLocked(dt.stackItem())
	}
}

// stackItem builds the StackItem this delayed trigger puts on the
// stack when it fires. The payload rides as TargetCard refs so the
// Effect reads it off the item (clone-safe) rather than out of a
// captured variable.
func (dt *DelayedTrigger) stackItem() *StackItem {
	item := &StackItem{
		ID:           uuid.New(),
		Kind:         StackItemTriggered,
		Controller:   dt.Controller,
		Owner:        dt.Controller,
		SourceCardID: dt.SourceCardID,
		Label:        dt.Label,
		Effect:       dt.Effect,
	}
	for _, cardID := range dt.Cards {
		if cardID == uuid.Nil {
			continue
		}
		item.Targets = append(item.Targets, TargetRef{Kind: TargetCard, ID: cardID})
	}
	return item
}

// cloneDelayedTrigger deep-copies one queued trigger. Cards is
// reallocated; Effect is shared, on the same contract cloneStackItem
// documents for StackItem.Effect. Returns nil for nil input.
func cloneDelayedTrigger(dt *DelayedTrigger) *DelayedTrigger {
	if dt == nil {
		return nil
	}
	out := &DelayedTrigger{
		ID:                 dt.ID,
		Controller:         dt.Controller,
		SourceCardID:       dt.SourceCardID,
		Label:              dt.Label,
		At:                 dt.At,
		ControllerTurnOnly: dt.ControllerTurnOnly,
		CreatedTurn:        dt.CreatedTurn,
		Effect:             dt.Effect,
		// #663: the event condition. AppliesTo and Optional are
		// shared, not copied, on exactly the contract Effect above
		// carries — they read the live *Game handed to them and
		// capture neither it nor a pointer into a zone slice.
		AppliesTo: dt.AppliesTo,
		Optional:  dt.Optional,
	}
	if dt.Duration != nil {
		d := *dt.Duration
		out.Duration = &d
	}
	if len(dt.On) > 0 {
		out.On = append([]EventKind(nil), dt.On...)
	}
	if len(dt.Cards) > 0 {
		out.Cards = append([]uuid.UUID(nil), dt.Cards...)
	}
	return out
}

// activePlayerIDLocked is the ID of the seat whose turn it is, or
// uuid.Nil before the game has an active seat. Caller must hold g.mu.
func (g *Game) activePlayerIDLocked() uuid.UUID {
	if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) || g.Seats[g.Turn.ActiveSeat] == nil {
		return uuid.Nil
	}
	return g.Seats[g.Turn.ActiveSeat].ID
}

// --- #663: the event condition ---------------------------------------
//
// "When you next cast an instant or sorcery spell this turn, copy that
// spell" is a delayed triggered ability whose condition is an EVENT
// rather than a step. ADR 0026 §1-2 decided against that and the
// 2026-09-18 amendment reverses it for this one case; the amendment
// carries the reasoning, and what follows is the whole implementation:
// one hook, one match, one dispatch.
//
// Everything else about a delayed trigger is unchanged. It is the same
// queue, the same clone, the same snapshot, the same "fires once and
// ceases to exist". What is new is where "once" is decided — the first
// matching event instead of the first matching step entry.

// fireEventDelayedTriggersLocked is the hook, called once per event
// from triggerHarvester.OnEvent after the zone walks. Every trigger
// whose event condition this event meets fires and is removed; a turn
// holding two Doublecasts fires both on the same cast, which is two
// copies and is what the cards print.
//
// The queue is rewritten BEFORE the first dispatch, for the reason
// fireDelayedTriggersLocked gives: dispatching emits EventTrigger,
// which re-enters the listener chain, and a listener reading a queue
// that still held already-fired entries would see them twice.
//
// Caller must hold g.mu in write mode.
func (g *Game) fireEventDelayedTriggersLocked(ev Event) {
	if len(g.DelayedTriggers) == 0 {
		return
	}
	var keep, fire []*DelayedTrigger
	for _, dt := range g.DelayedTriggers {
		if dt == nil {
			continue
		}
		if dt.matchesEventLocked(ev, g) {
			fire = append(fire, dt)
			continue
		}
		keep = append(keep, dt)
	}
	if len(fire) == 0 {
		return
	}
	g.DelayedTriggers = keep
	for _, dt := range fire {
		g.dispatchEventDelayedTriggerLocked(ev, dt)
	}
}

// matchesEventLocked is the event condition: a watched kind, then the
// card's own predicate. A trigger with no On never matches, which is
// what keeps every step-conditioned trigger in the queue untouched.
//
// Caller must hold g.mu.
func (dt *DelayedTrigger) matchesEventLocked(ev Event, g *Game) bool {
	if len(dt.On) == 0 || !triggerWatches(dt.On, ev.Kind) {
		return false
	}
	return dt.AppliesTo == nil || dt.AppliesTo(ev, dt, g)
}

// dispatchEventDelayedTriggerLocked hands a fired trigger to the
// HARVESTER's dispatch — the same dispatchTriggerLocked that
// harvestFromZone calls and that QueueReflexiveTriggerForEffect has
// called since ADR 0026's 2026-09-17 amendment.
//
// That reuse is the design, not a convenience: it is what makes the CR
// 603.5 "you may", the CR 603.3d target drop, the CR 608.2b re-check
// and the APNAP drain behave identically for a "when you next cast"
// and for an ETB, with one implementation of each rather than two.
//
// The triggering event's object rides as StackItem.Payload — "copy
// THAT spell" names the spell the event named, and the Effect reads it
// off the item rather than closing over it, which is what keeps the
// closure clone-safe (ADR 0026 §4).
//
// Caller must hold g.mu in write mode.
func (g *Game) dispatchEventDelayedTriggerLocked(ev Event, dt *DelayedTrigger) {
	source, lki := g.triggerSourceLocked(dt.SourceCardID, dt.Controller)
	label, effect := dt.Label, dt.Effect
	cards := append([]uuid.UUID(nil), dt.Cards...)
	ability := TriggeredAbility{
		// Watches / AppliesTo stay empty for the same reason a
		// reflexive trigger leaves them empty: the dispatch is handed
		// the match rather than asked to find one.
		Key:            label,
		OptionalPrompt: dt.Optional,
		Build: func(ev Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			item := NewTriggeredItem(source, label, effect)
			for _, cardID := range cards {
				if cardID != uuid.Nil {
					item.Targets = append(item.Targets, TargetRef{Kind: TargetCard, ID: cardID})
				}
			}
			if ev.CardID != uuid.Nil {
				item.Payload = append(item.Payload, TargetRef{Kind: TargetCard, ID: ev.CardID})
			}
			return item
		},
	}
	g.dispatchTriggerLocked(ev, source, lki, ability)
}

// clearExpiredDelayedTriggersLocked drops every delayed trigger whose
// duration has run out, fired or not (CR 514.2). Called from
// sweepTurnEndLocked beside the scoped-static and turn-scoped
// replacement sweeps — the same "this turn is over" pass — and it
// asks the same question they do: durationExpiredLocked is the ONE
// place in the engine that decides when a duration is over, and a
// second answer here would be a second rule.
//
// A trigger with no Duration is untouched: that is every
// step-conditioned trigger, and "at the beginning of the next end
// step" has to outlive the turn it was scheduled in.
//
// Allocates a fresh slice rather than compacting in place, because the
// backing array is shared with every undo snapshot Clone has taken —
// the same trap sweepScopedStaticsLocked documents. The sweep is
// idempotent; the cleanup hook runs it again after a discard pause
// drains.
//
// Caller must hold g.mu.
func (g *Game) clearExpiredDelayedTriggersLocked(endOfTurn bool) {
	if len(g.DelayedTriggers) == 0 {
		return
	}
	kept := make([]*DelayedTrigger, 0, len(g.DelayedTriggers))
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.Duration != nil && g.durationExpiredLocked(*dt.Duration, endOfTurn) {
			continue
		}
		kept = append(kept, dt)
	}
	if len(kept) == len(g.DelayedTriggers) {
		return
	}
	if len(kept) == 0 {
		kept = nil
	}
	g.DelayedTriggers = kept
}
