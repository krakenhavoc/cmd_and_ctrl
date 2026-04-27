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

	// PendingChoiceOptionalReplacement — CR 614.10 yes/no prompt
	// for an Optional replacement effect (today: CR 903.9
	// commander-zone). Owner answers yes → Replace runs; no →
	// effect is marked applied without running, event proceeds.
	// On resolve the chooser submits `{apply: true|false}`;
	// ResolveOptionalReplacement re-enters the pipeline. Added
	// in S17 sub-PR 6.
	PendingChoiceOptionalReplacement PendingChoiceKind = "optional_replacement"

	// PendingChoiceDamageAssignment — CR 510.1c prompt queued
	// when a multi-blocker combat damages step needs the
	// attacker's controller to assign damage across the ordered
	// blocker list. The chooser is the attacker's controller.
	// The client renders a drag-to-reorder blocker list with a
	// damage input per blocker (and, if the attacker has trample,
	// a "damage to defending player" input for overflow).
	//
	// On resolve the chooser submits
	// `{assignments: [{blocker_id, amount}, ...], trample_to_player: N}`.
	// ResolveDamageAssignment validates at-least-lethal prefix,
	// total = attacker power, and trample-only-if-trample. Added
	// in S18 sub-PR 3.
	PendingChoiceDamageAssignment PendingChoiceKind = "damage_assignment"

	// PendingChoiceTriggerPrompt — CR 603.4 "you may" yes/no
	// prompt queued by the S19 trigger harvester when a matching
	// TriggeredAbility has a non-nil OptionalPrompt. The chooser is
	// the source's controller (or an override defined on
	// OptionalPrompt.Chooser). On resolve `{apply: true}` the
	// stashed Build closure runs against the captured event + LKI
	// and the resulting StackItem appends to PendingTriggers; on
	// `{apply: false}` the trigger drops without effect.
	//
	// Modal-choice triggers ("draw a card OR gain 3 life") are NOT
	// covered here — sub-PR 2 ships the optional yes/no path only.
	// Added in S19 sub-PR 2.
	PendingChoiceTriggerPrompt PendingChoiceKind = "trigger_prompt"
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

	// DamageAssignment is the client-facing payload for a
	// PendingChoiceDamageAssignment entry: the attacker's instance
	// ID, ordered blocker instance IDs (in declared order; the
	// client may re-order), attacker's effective power, and
	// trample-allowed flag. Wire-serialised via
	// PendingChoiceView.DamageAssignment. Added in S18 sub-PR 3.
	DamageAssignment *DamageAssignmentFrame

	// triggerResume is the server-only continuation frame for a
	// PendingChoiceTriggerPrompt entry: the captured event +
	// source-card value copy + LKI characteristics + the Build
	// closure to invoke on `apply: true`. Not serialised to the
	// wire. Consumed by ResolveTriggerPrompt on submit. Added in
	// S19 sub-PR 2.
	triggerResume *triggerResumeFrame
}

// triggerResumeFrame stashes the per-trigger continuation data the
// harvester captured at OptionalPrompt-queue time. The Build closure
// fires on `apply: true` against the value-copy source + LKI; on
// `apply: false` the frame is discarded. Added in S19 sub-PR 2.
type triggerResumeFrame struct {
	ev     Event
	source Card
	lki    Characteristic
	build  func(ev Event, source *Card, sourceLKI Characteristic, g *Game) *StackItem
}

// DamageAssignmentFrame is the payload for a
// PendingChoiceDamageAssignment pending-choice entry. Both
// client-facing (serialised to the wire projection) and
// server-private (used by ResolveDamageAssignment to apply damage
// after validation). Added in S18 sub-PR 3.
type DamageAssignmentFrame struct {
	// AttackerID is the combat-damage source (the attacker).
	AttackerID uuid.UUID
	// BlockerIDs lists the creatures blocking this attacker, in
	// their declared order. The client re-orders via drag; the
	// server validates the re-ordered permutation on submit.
	BlockerIDs []uuid.UUID
	// AttackerPower is the effective power of the attacker at the
	// time of prompt queue. Total assigned damage must equal this
	// value; with trample, the leftover spills to
	// trample_to_player.
	AttackerPower int
	// AllowTrample is true when the attacker has the trample
	// keyword. The client renders a "to player" input only when
	// set; the server accepts trample_to_player > 0 only when set.
	AllowTrample bool
	// HasDeathtouch is true when the attacker has deathtouch (CR
	// 702.2c — 1 damage is lethal). The server uses this to relax
	// the at-least-lethal prefix rule: 1 damage satisfies the
	// threshold regardless of the blocker's remaining toughness.
	HasDeathtouch bool
	// FirstStrike is true when the assignment prompt was queued
	// from the first-strike substep. The resume path needs this
	// to avoid re-routing damage through the regular substep
	// hook.
	FirstStrike bool

	// SourceLifelink is the cached lifelink state of the attacker
	// at prompt-queue time. Captured here because the attacker may
	// have been destroyed by blocker damage (which resolves in the
	// same substep before the prompt fires) — looking it up again
	// at resume time would miss the keyword.
	SourceLifelink bool

	// SourceController is the attacker's controller at prompt-queue
	// time. Captured for the same reason as SourceLifelink (lifelink
	// credits this player even if the attacker is no longer on the
	// battlefield).
	SourceController uuid.UUID
}

// replacementResumeFrame is the unexported per-prompt continuation
// stash. Holds the ReplacementEvent being processed + the gathered
// list so ResolveReplacementOrder can re-enter the apply-loop with
// the chosen order locked in. Added in S17 sub-PR 2.
//
// For PendingChoiceOptionalReplacement (sub-PR 6), `applicable`
// carries a single entry — the optional effect the prompt is
// asking about. The resume path either fires that effect's
// Replace (on yes) or skips it (on no), then re-enters the apply-
// loop for CR 616.1 iteration.
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

// queueOptionalReplacementPromptLocked queues a CR 614.10 yes/no
// prompt for a single optional replacement effect. The chooser is
// the effect's Controller (for CR 903.9 commander-zone: the
// commander's owner). The resume path in ResolveOptionalReplacement
// either fires the Replace (on yes) or marks it applied and skips
// (on no), then re-enters the apply-loop.
//
// Caller must hold g.mu.
func (g *Game) queueOptionalReplacementPromptLocked(ev *ReplacementEvent, chosen activeReplacement) {
	var chooser uuid.UUID
	if chosen.effect.Controller != nil {
		chooser = chosen.effect.Controller(ev, g, chosen.source)
	}
	if chooser == uuid.Nil {
		chooser = affectedPlayerForEvent(ev, []activeReplacement{chosen}, g)
	}
	reason := chosen.effect.PromptQuestion
	if reason == "" {
		reason = chosen.effect.Label
	}
	choice := PendingChoice{
		Kind:                 PendingChoiceOptionalReplacement,
		Chooser:              chooser,
		Count:                1,
		Reason:               reason,
		ReplacementEffectIDs: []ReplacementEffectID{chosen.id},
		replacementResume: &replacementResumeFrame{
			ev:         ev,
			applicable: []activeReplacement{chosen},
		},
	}
	g.QueueChoiceForEffect(choice)
}

// ResolveOptionalReplacement processes a resolve_choice action
// for a PendingChoiceOptionalReplacement entry. `apply` is the
// owner's yes/no decision: true → fire the stashed Replace; false
// → mark applied without firing. Either way, re-enters the apply-
// loop so CR 616.1 can pick up any newly-applicable effects.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveOptionalReplacement(choiceID, chooserID uuid.UUID, apply bool) error {
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
	if choice.Kind != PendingChoiceOptionalReplacement {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	frame := choice.replacementResume
	g.dequeueChoiceLocked(idx)
	if frame == nil || frame.ev == nil || len(frame.applicable) == 0 {
		return ErrInvalidParam
	}

	ev := frame.ev
	chosen := frame.applicable[0]
	if g.replacementsAppliedThisEvent == nil {
		g.replacementsAppliedThisEvent = make(map[ReplacementEventID]map[ReplacementEffectID]bool)
	}
	if _, ok := g.replacementsAppliedThisEvent[ev.ID]; !ok {
		g.replacementsAppliedThisEvent[ev.ID] = make(map[ReplacementEffectID]bool)
	}
	// Mark applied regardless of yes/no so the apply-loop doesn't
	// re-evaluate this effect again for this event (CR 614.10: the
	// decision is once per event).
	g.replacementsAppliedThisEvent[ev.ID][chosen.id] = true
	if apply && chosen.effect.Replace != nil {
		if err := chosen.effect.Replace(ev, g, chosen.source); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
		}
	}

	// Re-enter the apply-loop for any newly-applicable effects.
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		return nil
	}
	if err != nil {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		return nil
	}
	return g.applyResolvedReplacementEventLocked(out)
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

	// Apply ALL chosen effects in the submitted order (CR 616: the
	// affected player picks the order once; the engine fires them
	// in that order without re-prompting). Then re-enter the apply-
	// loop so CR 616.1 can pick up any newly-applicable effects
	// (effects that weren't applicable until one of these fired).
	//
	// Earlier drafts fired only ordered[0] and relied on the apply-
	// loop to re-queue a prompt for the remaining effects — that
	// mis-read 616.1 and forced the user to submit the same order
	// N times for N replacements. The right behavior is "one prompt
	// = one ordering decision, apply them all in sequence."
	applicableByID := make(map[ReplacementEffectID]activeReplacement, len(frame.applicable))
	for _, a := range frame.applicable {
		applicableByID[a.id] = a
	}
	ev := frame.ev
	if g.replacementsAppliedThisEvent == nil {
		g.replacementsAppliedThisEvent = make(map[ReplacementEventID]map[ReplacementEffectID]bool)
	}
	if _, ok := g.replacementsAppliedThisEvent[ev.ID]; !ok {
		g.replacementsAppliedThisEvent[ev.ID] = make(map[ReplacementEffectID]bool)
	}
	for _, id := range ordered {
		chosen, ok := applicableByID[id]
		if !ok {
			continue
		}
		if ev.Canceled {
			// A prior Cancel short-circuits the remaining chain.
			break
		}
		g.replacementsAppliedThisEvent[ev.ID][chosen.id] = true
		if chosen.effect.Replace != nil {
			if err := chosen.effect.Replace(ev, g, chosen.source); err != nil {
				g.EmitEvent(Event{Kind: EventEffectError, ErrorMsg: err.Error()})
			}
		}
	}

	// Resume the apply-loop to pick up any newly-applicable effects
	// (CR 616.1 — one of the applied replacements may have enabled
	// another that wasn't in the original prompt). Effects already
	// in replacementsAppliedThisEvent are skipped by gather.
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
	case RepEventMove:
		// S17 sub-PR 6: resume path for battlefield-leave moves
		// after the CR 903.9 commander-zone Optional prompt. The
		// pipeline settled on ev.NewZone (either the original
		// destination if owner said "no", or ZoneCommand if they
		// said "yes"). Run the physical move through the shared
		// executeBattlefieldLeaveLocked helper.
		if ev.OldZone != ZoneBattlefield {
			// Non-LTB moves (e.g. graveyard → battlefield for
			// reanimate) don't have a resume path yet. Sub-PR 6
			// only closes the battlefield-leave case.
			return nil
		}
		var owner *Player
		if card, ok := g.LookupCardForEffect(ev.CardID); ok {
			owner = g.playerByIDLocked(card.Owner)
		}
		return g.executeBattlefieldLeaveLocked(ev.CardID, ev.NewZone, ev.NewZoneOwner, owner)
	case RepEventStepTransition:
		// Step-transition resumes land with sub-PR 4+ (Stasis
		// skip-step). Sub-PR 6 doesn't add new prompt paths here.
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

// DamageAssignmentEntry is one {blocker, amount} pair from a
// damage-assignment resolve. Added in S18 sub-PR 3.
type DamageAssignmentEntry struct {
	BlockerID uuid.UUID
	Amount    int
}

// ResolveDamageAssignment processes a resolve_choice action for a
// PendingChoiceDamageAssignment entry. Validates:
//   - the choice ID exists in the queue
//   - the chooserID matches the attacker's controller
//   - ordered is a permutation of the blocker IDs from the frame
//   - sum(amounts) + trampleToPlayer == attacker power
//   - each ordered-prefix blocker is assigned at-least-lethal before
//     the next one receives any damage (CR 510.1c). Deathtouch
//     relaxes the threshold to 1.
//   - trampleToPlayer > 0 only when the attacker has trample
//
// On success, applies damage via the regular combat-damage path so
// replacement (Fog), lifelink (mark source's controller), and
// deathtouch (flag target) all fire. Added in S18 sub-PR 3.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveDamageAssignment(
	choiceID, chooserID uuid.UUID,
	ordered []DamageAssignmentEntry,
	trampleToPlayer int,
) error {
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
	if choice.Kind != PendingChoiceDamageAssignment {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	frame := choice.DamageAssignment
	if frame == nil {
		return ErrInvalidParam
	}
	if len(ordered) != len(frame.BlockerIDs) {
		return ErrInvalidParam
	}
	// Permutation check.
	want := make(map[uuid.UUID]int, len(frame.BlockerIDs))
	for _, id := range frame.BlockerIDs {
		want[id]++
	}
	for _, e := range ordered {
		if want[e.BlockerID] <= 0 {
			return ErrInvalidParam
		}
		want[e.BlockerID]--
		if e.Amount < 0 {
			return ErrInvalidParam
		}
	}
	// Trample gate — the client may send 0 even without trample,
	// which is fine; nonzero without trample is rejected.
	if trampleToPlayer < 0 {
		return ErrInvalidParam
	}
	if trampleToPlayer > 0 && !frame.AllowTrample {
		return ErrInvalidParam
	}
	// Sum check.
	total := trampleToPlayer
	for _, e := range ordered {
		total += e.Amount
	}
	if total != frame.AttackerPower {
		return ErrInvalidParam
	}
	// Prefix-lethal check: every blocker before the last non-zero
	// assignment must have received at-least-lethal damage. Lethal
	// threshold = max(1, blocker.CurrentToughness - blocker.DamageMarked).
	// Deathtouch collapses the threshold to 1.
	assigned := make(map[uuid.UUID]int, len(ordered))
	for i, e := range ordered {
		lethal := 1
		if !frame.HasDeathtouch {
			blk := findBattlefieldCard(g, e.BlockerID)
			if blk == nil {
				// Blocker left the battlefield mid-prompt. Treat as
				// lethal satisfied (no target).
				lethal = 0
			} else {
				remaining := blk.CurrentToughness() - blk.DamageMarked
				if remaining > 1 {
					lethal = remaining
				} else if remaining <= 0 {
					// Already lethal-marked; damage still applies but
					// threshold is satisfied.
					lethal = 0
				}
			}
		}
		// Earlier blockers must be at-least-lethal before this one
		// receives any damage (CR 510.1c "assigns damage in order").
		for j := 0; j < i; j++ {
			prior := ordered[j]
			priorLethal := 1
			if !frame.HasDeathtouch {
				pblk := findBattlefieldCard(g, prior.BlockerID)
				if pblk != nil {
					priorRemaining := pblk.CurrentToughness() - pblk.DamageMarked
					if priorRemaining > 1 {
						priorLethal = priorRemaining
					} else if priorRemaining <= 0 {
						priorLethal = 0
					}
				}
			}
			if prior.Amount < priorLethal {
				return ErrInvalidParam
			}
		}
		// If this blocker got less than lethal AND anything downstream
		// got non-zero damage OR trample spilled, reject.
		if e.Amount < lethal {
			for j := i + 1; j < len(ordered); j++ {
				if ordered[j].Amount > 0 {
					return ErrInvalidParam
				}
			}
			if trampleToPlayer > 0 {
				return ErrInvalidParam
			}
		}
		assigned[e.BlockerID] = e.Amount
	}
	g.dequeueChoiceLocked(idx)

	// Apply damage using the frame-cached source keywords: the
	// attacker may have been destroyed by blocker damage (which
	// resolved in the same substep before this prompt fires), so
	// looking up HasKeyword on the attacker at resume time would
	// miss deathtouch / lifelink. The frame captured those at
	// queue time.
	for _, e := range ordered {
		if e.Amount <= 0 {
			continue
		}
		g.markCombatDamageFromFrameLocked(e.BlockerID, e.Amount, frame)
	}
	if trampleToPlayer > 0 {
		// Determine the defending player: find the attacker if
		// still on battlefield; otherwise look at the prompt's
		// original intent — trample-to-player was only allowed
		// for a live attacker that had declared a target. Fall
		// back to any seat that isn't the controller (best-effort).
		var defenderID uuid.UUID
		if atkCard := findBattlefieldCard(g, frame.AttackerID); atkCard != nil {
			defenderID = atkCard.AttackingTarget
		}
		if defenderID == uuid.Nil {
			for _, s := range g.Seats {
				if s.ID != frame.SourceController {
					defenderID = s.ID
					break
				}
			}
		}
		if defenderID != uuid.Nil {
			g.markCombatDamageToPlayerFromFrameLocked(defenderID, trampleToPlayer, frame)
		}
	}
	// Blockers also deal their power back (simultaneous damage —
	// CR 510.1d). The attacker loop in assignAndDealCombatDamageLocked
	// already marked blocker→attacker damage before queuing the
	// prompt, so we don't re-fire it here.
	//
	// Run SBAs so deaths from this assignment land before the next
	// substep / step advance.
	g.runStateChecksLocked()
	return nil
}

// queueTriggerPromptLocked queues a CR 603.4 yes/no prompt for an
// optional triggered ability that just matched. Captures the event,
// a value copy of the source, and the LKI snapshot — all of which
// the resume path will pass back into the Build closure on `apply:
// true`. The chooser is the source's controller, unless the
// ability's OptionalPrompt overrides it for opponent-prompted
// triggers.
//
// Caller must hold g.mu. Added in S19 sub-PR 2.
func (g *Game) queueTriggerPromptLocked(
	ev Event,
	source Card,
	lki Characteristic,
	ability TriggeredAbility,
) {
	chooser := source.Controller
	if ability.OptionalPrompt != nil && ability.OptionalPrompt.Chooser != nil {
		if override := ability.OptionalPrompt.Chooser(ev, &source, g); override != uuid.Nil {
			chooser = override
		}
	}
	question := ""
	if ability.OptionalPrompt != nil {
		question = ability.OptionalPrompt.Question
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:    PendingChoiceTriggerPrompt,
		Chooser: chooser,
		Count:   1,
		Source:  source.InstanceID,
		Reason:  question,
		triggerResume: &triggerResumeFrame{
			ev:     ev,
			source: source,
			lki:    lki,
			build:  ability.Build,
		},
	})
}

// ResolveTriggerPrompt processes the controller's yes/no answer for
// a PendingChoiceTriggerPrompt entry. On `apply: true` the stashed
// Build closure runs against the captured event + LKI; the resulting
// StackItem (if non-nil) appends to PendingTriggers via the same
// queueHarvestedTriggerLocked the mandatory path uses. On `apply:
// false` the entry is dropped silently — the trigger is treated as
// having never been declared.
//
// Caller must NOT hold g.mu — this method takes the write lock.
// Added in S19 sub-PR 2.
func (g *Game) ResolveTriggerPrompt(choiceID, chooserID uuid.UUID, apply bool) error {
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
	if choice.Kind != PendingChoiceTriggerPrompt {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	frame := choice.triggerResume
	g.dequeueChoiceLocked(idx)
	if !apply || frame == nil || frame.build == nil {
		return nil
	}
	source := frame.source
	item := frame.build(frame.ev, &source, frame.lki, g)
	if item == nil {
		return nil
	}
	g.queueHarvestedTriggerLocked(item)
	return nil
}

// PendingChoiceKindFor returns the kind of the queue entry with the
// given ID, or empty + false when no such entry exists. Used by the
// resolve_choice action dispatcher to route a yes/no payload to the
// right resolve method (S17 PendingChoiceOptionalReplacement vs S19
// PendingChoiceTriggerPrompt — both consume `{apply: bool}` so the
// dispatcher can't disambiguate from the payload shape alone).
//
// Caller must NOT hold g.mu — takes the read lock. Added in S19
// sub-PR 2.
func (g *Game) PendingChoiceKindFor(choiceID uuid.UUID) (PendingChoiceKind, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			return c.Kind, true
		}
	}
	return "", false
}

// findBattlefieldCard returns a pointer to the battlefield card
// with the given instance ID, or nil. Caller must hold g.mu.
func findBattlefieldCard(g *Game, id uuid.UUID) *Card {
	if g.Battlefield == nil {
		return nil
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}
