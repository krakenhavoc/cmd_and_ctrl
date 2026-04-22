package game

import (
	"errors"

	"github.com/google/uuid"
)

// pending_choice.go holds the S14+ generic "someone needs to make
// a pick" infrastructure. Replaces ad-hoc per-effect deferred-
// choice state with a unified queue that the client picker modal
// can drain one choice at a time.
//
// Shape: a choice names its Chooser (who picks), the source card
// that created it, what kind of pick it is, and — when the pick
// comes from somewhere other than the chooser's own zones — a
// FromPlayer reference. Thoughtseize is the canonical chooser !=
// discarder case: the caster is Chooser, the target is FromPlayer.
//
// The existing S13.4 DiscardPending map stays for cleanup-step
// max-hand-size discards (those are simpler: chooser == owner,
// and the cursor auto-resumes when the map drains). Effect-
// driven choices that go through PendingChoices keep resolving
// asynchronously — the spell routes to graveyard immediately,
// the pick is made later by a resolve_choice action.
//
// Introduced in S14 sub-PR 5 as infrastructure for Thoughtseize.

// PendingChoiceKind discriminates what kind of pick is owed.
// Extensible — each kind is interpreted by the server when
// ResolvePendingChoice fires and by the client when rendering the
// picker. Values:
//
//	"discard_from_hand" — chooser picks Count card IDs from
//	                      FromPlayer's hand; those cards move to
//	                      FromPlayer's graveyard.
//
// Future kinds (reserved names, not implemented yet):
//
//	"mode_pick"         — chooser picks a mode index for a modal
//	                      spell.
//	"mill_reveal"       — chooser picks which of the revealed top-N
//	                      library cards go where (Brainstorm's
//	                      put-two-back half).
type PendingChoiceKind string

const (
	PendingChoiceDiscardFromHand PendingChoiceKind = "discard_from_hand"

	// PendingChoiceMana — the chooser picks one color from a fixed
	// option set (Birds of Paradise's WUBRG, Arcane Signet's
	// commander-identity subset). On resolve the picked color drops
	// into the chooser's pool as one ManaToken sourced from the
	// permanent that fired the ability. Added in S15 sub-PR 2.
	PendingChoiceMana PendingChoiceKind = "mana_pick"

	// PendingChoiceReplacementOrder — CR 616 affected-player-
	// chooses-order prompt queued when ≥2 replacement effects
	// apply to the same event. ReplacementEffectIDs carries the
	// gathered IDs; the client renders a drag-reorder list with
	// label + source-card context. On resolve the chooser submits
	// the order as an `order []string` payload of the same IDs in
	// their chosen order; ResolveReplacementOrder validates the
	// permutation and resumes the pipeline. Added in S17 sub-PR 2.
	PendingChoiceReplacementOrder PendingChoiceKind = "replacement_order"
)

// PendingChoice is one outstanding "someone needs to pick" entry
// in the game's queue. Serialized to the wire via
// protocol.PendingChoiceView with the Options slice materialized
// at view time.
type PendingChoice struct {
	// ID uniquely identifies the choice for address-by-ID in the
	// resolve_choice action. Minted fresh on queue.
	ID uuid.UUID

	// Kind determines how the picks[] payload is interpreted on
	// resolve_choice.
	Kind PendingChoiceKind

	// Chooser is the player responsible for submitting picks. Only
	// they see the picker modal.
	Chooser uuid.UUID

	// FromPlayer owns the pool the picks come from. For
	// discard_from_hand, this is the player whose hand the picks
	// live in. Equals Chooser for self-discard (Mind Rot-style)
	// choices; differs for Thoughtseize (caster picks from
	// target's hand).
	FromPlayer uuid.UUID

	// Count is the number of items the chooser must pick. 1 for
	// Thoughtseize; 2 for Mind Rot if we ever migrate it here.
	Count int

	// Source is the card that queued this choice (Thoughtseize's
	// instance ID). Empty uuid when the choice isn't card-
	// originated (e.g. admin-driven test harness).
	Source uuid.UUID

	// Reason is a free-text label for the picker modal's header
	// ("Thoughtseize", "Vendilion Clique"). Kept on the server so
	// the wire carries it; no localisation yet.
	Reason string

	// ColorOptions is the legal-picks list for PendingChoiceMana.
	// Uppercase single-character entries (W/U/B/R/G/C). Unused for
	// other choice kinds. Populated server-side so the client
	// renders a color picker with exactly the right buttons
	// (commander-identity-filtered for Arcane Signet, the full five
	// for Birds of Paradise). Added in S15 sub-PR 2.
	ColorOptions []string

	// ReplacementEffectIDs is the ordered set of applicable
	// replacement-effect IDs the chooser must reorder for a
	// PendingChoiceReplacementOrder entry. The resolve_choice
	// payload returns the same IDs in the chosen order. Unused
	// for other choice kinds. Added in S17 sub-PR 2.
	ReplacementEffectIDs []ReplacementEffectID

	// replacementResume is the server-only continuation frame for
	// a PendingChoiceReplacementOrder entry: the in-flight
	// ReplacementEvent + the gathered applicable list. Not
	// serialised to the wire. Consumed by ResolveReplacementOrder
	// on submit. Added in S17 sub-PR 2.
	replacementResume *replacementResumeFrame
}

// replacementResumeFrame is the unexported per-prompt continuation
// stash. Holds the ReplacementEvent being processed + the gathered
// list so ResolveReplacementOrder can re-enter the apply-loop with
// the chosen order locked in. Added in S17 sub-PR 2.
type replacementResumeFrame struct {
	ev         *ReplacementEvent
	applicable []activeReplacement
}

// QueueChoiceForEffect appends a PendingChoice to the game's queue.
// Caller must hold g.mu. Returns the generated ID so the caller
// can reference the choice downstream if needed.
func (g *Game) QueueChoiceForEffect(choice PendingChoice) uuid.UUID {
	if choice.ID == uuid.Nil {
		choice.ID = uuid.New()
	}
	g.PendingChoices = append(g.PendingChoices, &choice)
	return choice.ID
}

// ResolvePendingChoice processes a resolve_choice action.
// Validates:
//   - the choice ID exists in the queue
//   - the chooserID matches the queue entry's Chooser
//   - len(picks) == entry.Count
//   - every pick is in the expected source zone (for
//     discard_from_hand, entry.FromPlayer's hand)
//
// On success, applies the kind-specific side effect (for
// discard_from_hand: move each pick to FromPlayer's graveyard,
// emit EventDiscardCard per pick), dequeues the entry, and
// returns nil.
//
// Caller must NOT hold g.mu — this method takes the write lock.
func (g *Game) ResolvePendingChoice(choiceID, chooserID uuid.UUID, picks []uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := -1
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	choice := g.PendingChoices[idx]
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if len(picks) != choice.Count {
		return ErrInvalidParam
	}
	switch choice.Kind {
	case PendingChoiceMana:
		// Mana picks are resolved by ResolveManaChoice — the payload
		// is a color string, not a list of card IDs. Callers who
		// route here with card-ID picks hit a guard rail.
		return ErrInvalidParam
	case PendingChoiceDiscardFromHand:
		from := g.playerByIDLocked(choice.FromPlayer)
		if from == nil {
			// FromPlayer left the game mid-choice. Drop the entry
			// silently — the choice is moot.
			g.dequeueChoiceLocked(idx)
			return nil
		}
		// Pre-validate every pick sits in from.Hand.
		for _, id := range picks {
			if !from.Hand.Contains(id) {
				return ErrCardNotFound
			}
		}
		for _, id := range picks {
			if _, err := MoveCard(from.Hand, from.Graveyard, id); err != nil {
				return err
			}
			g.markCardKnownInZoneLocked(from.Graveyard, id)
			g.EmitEvent(Event{
				Kind:    EventDiscardCard,
				Actor:   from.ID,
				CardID:  id,
				OldZone: ZoneHand,
				NewZone: ZoneGraveyard,
			})
		}
	default:
		return ErrInvalidParam
	}
	g.dequeueChoiceLocked(idx)
	return nil
}

// ResolveManaChoice processes a resolve_choice action for a
// PendingChoiceMana entry. The chooser picks one color from the
// entry's ColorOptions; the picked color drops into their pool as
// one ManaToken sourced from the permanent that fired the ability
// (choice.Source). Emits EventManaAdded.
//
// Distinct from ResolvePendingChoice because the payload shape is a
// color string, not a list of card IDs. The dispatcher decides
// which to call based on the action's payload.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveManaChoice(choiceID, chooserID uuid.UUID, color string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := -1
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	choice := g.PendingChoices[idx]
	if choice.Kind != PendingChoiceMana {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	// Validate color is in the pre-materialised option set.
	matched := false
	for _, o := range choice.ColorOptions {
		if o == color {
			matched = true
			break
		}
	}
	if !matched {
		return ErrInvalidParam
	}
	p := g.playerByIDLocked(chooserID)
	if p == nil {
		return ErrPlayerNotFound
	}
	p.ManaPool.AddMana(ManaToken{Color: color, Source: choice.Source})
	g.EmitEvent(Event{
		Kind:   EventManaAdded,
		Actor:  chooserID,
		Source: choice.Source,
	})
	g.dequeueChoiceLocked(idx)
	return nil
}

// dequeueChoiceLocked drops the choice at index idx, preserving
// slice order for the rest. Caller must hold g.mu.
func (g *Game) dequeueChoiceLocked(idx int) {
	if idx < 0 || idx >= len(g.PendingChoices) {
		return
	}
	g.PendingChoices = append(g.PendingChoices[:idx], g.PendingChoices[idx+1:]...)
	if len(g.PendingChoices) == 0 {
		g.PendingChoices = nil
	}
}

// queueReplacementOrderPromptLocked queues a CR 616 order-choose
// prompt for ev. The chooser is the affected-player (inferred from
// ev.Kind's target field). ReplacementEffectIDs are emitted in the
// order the engine gathered them; the client renders a drag-
// reorder list and returns the same IDs in the chosen order. The
// resume frame stashes ev + applicable so ResolveReplacementOrder
// can re-enter the apply-loop.
//
// Caller must hold g.mu.
func (g *Game) queueReplacementOrderPromptLocked(ev *ReplacementEvent, applicable []activeReplacement) {
	ids := make([]ReplacementEffectID, 0, len(applicable))
	for _, a := range applicable {
		ids = append(ids, a.id)
	}
	chooser := affectedPlayerForEvent(ev, applicable, g)
	choice := PendingChoice{
		Kind:                 PendingChoiceReplacementOrder,
		Chooser:              chooser,
		Count:                len(ids),
		Reason:               "Order replacement effects",
		ReplacementEffectIDs: ids,
		replacementResume: &replacementResumeFrame{
			ev:         ev,
			applicable: applicable,
		},
	}
	g.QueueChoiceForEffect(choice)
}

// affectedPlayerForEvent returns the chooser for a CR 616 prompt —
// the player whose event is being replaced (the affected player).
// Falls back to the first applicable replacement's Controller, then
// to uuid.Nil. Callers are expected to have ≥1 applicable entry.
func affectedPlayerForEvent(ev *ReplacementEvent, applicable []activeReplacement, g *Game) uuid.UUID {
	if ev == nil {
		return uuid.Nil
	}
	switch ev.Kind {
	case RepEventDraw:
		return ev.DrawPlayer
	case RepEventLife:
		return ev.LifePlayer
	case RepEventCounter:
		if card, ok := g.LookupCardForEffect(ev.CounterTarget); ok {
			return card.Controller
		}
	case RepEventDamage:
		if card, ok := g.LookupCardForEffect(ev.DamageTarget); ok {
			return card.Controller
		}
		return ev.DamageTarget
	case RepEventMove:
		if card, ok := g.LookupCardForEffect(ev.CardID); ok {
			return card.Controller
		}
	case RepEventStepTransition:
		if ev.StepTransitionSeat >= 0 && ev.StepTransitionSeat < len(g.Seats) {
			return g.Seats[ev.StepTransitionSeat].ID
		}
	}
	if len(applicable) > 0 && applicable[0].effect.Controller != nil {
		return applicable[0].effect.Controller(ev, g, applicable[0].source)
	}
	return uuid.Nil
}

// ResolveReplacementOrder processes a resolve_choice action for a
// PendingChoiceReplacementOrder entry. Validates:
//   - the choice ID exists in the queue
//   - the chooserID matches the entry's Chooser
//   - ordered is a permutation of the entry's ReplacementEffectIDs
//
// On success, re-enters the replacement apply-loop with the chosen
// order locked in for the current iteration. Subsequent iterations
// may queue another prompt (the chain unrolls asynchronously, one
// resolve_choice per branch-point). After the apply-loop settles,
// the pipeline function's resume helper re-invokes the underlying
// mutation with the (possibly mutated / canceled) event.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveReplacementOrder(choiceID, chooserID uuid.UUID, ordered []ReplacementEffectID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := -1
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	choice := g.PendingChoices[idx]
	if choice.Kind != PendingChoiceReplacementOrder {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	if len(ordered) != len(choice.ReplacementEffectIDs) {
		return ErrInvalidParam
	}
	// Permutation check: same multiset of IDs.
	want := make(map[ReplacementEffectID]int, len(choice.ReplacementEffectIDs))
	for _, id := range choice.ReplacementEffectIDs {
		want[id]++
	}
	for _, id := range ordered {
		if want[id] <= 0 {
			return ErrInvalidParam
		}
		want[id]--
	}
	frame := choice.replacementResume
	g.dequeueChoiceLocked(idx)
	if frame == nil || frame.ev == nil {
		return ErrInvalidParam
	}

	// Apply the chosen order's first pick, then re-enter the
	// apply-loop. Subsequent branches queue their own prompts as
	// needed.
	applicableByID := make(map[ReplacementEffectID]activeReplacement, len(frame.applicable))
	for _, a := range frame.applicable {
		applicableByID[a.id] = a
	}
	first := ordered[0]
	chosen, ok := applicableByID[first]
	if !ok {
		return ErrInvalidParam
	}
	ev := frame.ev
	if g.replacementsAppliedThisEvent == nil {
		g.replacementsAppliedThisEvent = make(map[ReplacementEventID]map[ReplacementEffectID]bool)
	}
	if _, ok := g.replacementsAppliedThisEvent[ev.ID]; !ok {
		g.replacementsAppliedThisEvent[ev.ID] = make(map[ReplacementEffectID]bool)
	}
	g.replacementsAppliedThisEvent[ev.ID][chosen.id] = true
	if chosen.effect.Replace != nil {
		if err := chosen.effect.Replace(ev, g, chosen.source); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
		}
	}

	// Resume the apply-loop to pick up any further replacements
	// (iterative CR 616.1 — one replacement might enable another).
	// If more than one applicable remains, this call queues
	// another prompt and returns early with errReplacementPending.
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// Another prompt queued; unroll asynchronously.
		return nil
	}
	if err != nil {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}

	// Apply-loop settled. Dispatch the underlying mutation per
	// ev.Kind using the (possibly mutated) event payload. Added in
	// S17 sub-PR 3 so Doubling Season + Hardened Scales actually
	// land counters after the CR 616 prompt resolves.
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		return nil
	}
	return g.applyResolvedReplacementEventLocked(out)
}

// applyResolvedReplacementEventLocked runs the underlying
// mutation for a fully-settled ReplacementEvent — called from the
// CR 616 resume path (ResolveReplacementOrder) after the
// replacement apply-loop finishes with no pending prompts. The
// event's payload may have been mutated by replacements (e.g.
// Doubling Season doubled CounterDelta; Library of Leng rewrote
// NewZone). Pipeline functions' initial (non-paused) path inlines
// the same mutation; the resume path uses this central dispatcher.
//
// Caller must hold g.mu.
func (g *Game) applyResolvedReplacementEventLocked(ev *ReplacementEvent) error {
	switch ev.Kind {
	case RepEventCounter:
		return g.applyCounterLocked(ev.CounterTarget, ev.CounterName, ev.CounterDelta)
	case RepEventDraw:
		return g.actuallyDrawCardLocked(ev.DrawPlayer)
	case RepEventLife:
		p := g.playerByIDLocked(ev.LifePlayer)
		if p == nil {
			return ErrPlayerNotFound
		}
		p.ChangeLife(ev.LifeDelta)
		g.EmitEvent(Event{Kind: EventChangeLife, Target: ev.LifePlayer, Amount: ev.LifeDelta})
		return nil
	case RepEventDamage:
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ev.DamageTarget {
				g.Battlefield.Cards[i].DamageMarked += ev.DamageAmount
				if g.Battlefield.Cards[i].DamageMarked < 0 {
					g.Battlefield.Cards[i].DamageMarked = 0
				}
				if ev.DamageAmount > 0 {
					g.EmitEvent(Event{
						Kind:   EventDealDamage,
						Source: ev.DamageSource,
						Target: ev.DamageTarget,
						Amount: ev.DamageAmount,
					})
				}
				g.runStateChecksLocked()
				return nil
			}
		}
		return ErrCardNotFound
	case RepEventMove, RepEventStepTransition:
		// Move + step-transition resumes land with sub-PR 4+
		// (Kismet enters-tapped routing, Stasis skip-step). For
		// sub-PR 3 no catalog card queues a multi-replacement
		// prompt on these kinds, so this is a no-op TODO.
		return nil
	}
	return nil
}

// QueueDiscardFromRevealedHand is the Thoughtseize entry point.
// Reveals the target's hand to the chooser (sticky via S13.5
// KnownBy), then queues a discard_from_hand PendingChoice. The
// spell can return nil from its OnResolve immediately — the
// discard fires asynchronously when the chooser submits their
// pick via resolve_choice.
//
// Caller must hold g.mu.
func (g *Game) QueueDiscardFromRevealedHand(
	chooser, fromPlayer, source uuid.UUID,
	count int,
	reason string,
) uuid.UUID {
	// Reveal to the chooser specifically so their client-side KnownBy
	// lets redactCardForViewer keep the identity. The reveal is
	// sticky — cards the chooser saw stay revealed after the choice
	// resolves, same as any other reveal effect.
	if p := g.playerByIDLocked(fromPlayer); p != nil {
		for i := range p.Hand.Cards {
			p.Hand.Cards[i].AddKnower(chooser)
		}
	}
	// Cap count to available hand size so the chooser isn't stuck
	// on an impossible count (CR 701.8c "as many as you can").
	if p := g.playerByIDLocked(fromPlayer); p != nil && p.Hand.Size() < count {
		count = p.Hand.Size()
	}
	if count <= 0 {
		return uuid.Nil
	}
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:       PendingChoiceDiscardFromHand,
		Chooser:    chooser,
		FromPlayer: fromPlayer,
		Count:      count,
		Source:     source,
		Reason:     reason,
	})
}
