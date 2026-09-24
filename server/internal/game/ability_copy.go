package game

import "github.com/google/uuid"

// ability_copy.go — CR 707.10, copying an ACTIVATED or TRIGGERED
// ability that is already on the stack (#1223).
//
// "Copy target activated or triggered ability you control. You may
// choose new targets for the copy" (Lithoform Engine), "copy target
// triggered ability you control" (Strionic Resonator), and "whenever
// you activate an ability, if it isn't a mana ability, you may pay
// {2}. If you do, copy that ability" (Rings of Brighthearth).
//
// It is the same rule as a spell copy and it is deliberately not a
// second implementation of one. What CR 707.10 says about an ability
// is what it says about a spell — the copy is a new object with the
// original's characteristics and the choices made when it was put on
// the stack — so the PROMPT, the gate and the re-target frame are
// spell_copy.go's, and this file is the four places the two differ:
//
//  1. THERE IS NO CARD. A spell is a card on the stack, and
//     createSpellCopyLocked builds a second card beside it. An
//     ability is a StackItem alone: its source permanent stays on the
//     battlefield, unaware, and the copy is one more entry in
//     StackMeta. Nothing is pushed into a zone, so nothing has to be
//     removed from one later — which is also why ceaseToExistLocked
//     has no ability twin. CR 608.2n already ends every ability item
//     on resolution, copy or not.
//
//  2. IT WAS NOT ACTIVATED AND IT DID NOT TRIGGER (CR 707.10a). The
//     copy is CREATED. Nothing announced it, nothing paid for it and
//     nothing triggered it, so no EventActivateAbility and no
//     EventTrigger are emitted here — and, exactly as with the spell
//     copy's missing EventCast, that is a fact about what this file
//     DOES NOT DO rather than a suppression flag somebody downstream
//     reads. Rings of Brighthearth copying its own copy forever, and
//     "whenever an opponent activates an ability" firing twice for
//     one activation, are the two bugs the omission prevents.
//
//  3. A MANA ABILITY CANNOT BE COPIED. CR 605.3b: a mana ability
//     resolves immediately and never uses the stack, so there is no
//     item to name. That falls out of the lookup finding nothing
//     (ErrCardNotFound) rather than out of a check, which is the
//     stronger version of the rule: Rings of Brighthearth does not
//     even trigger, because a mana ability announces
//     EventManaAbilityActivated and the card watches the other kind.
//
//  4. THE TRIGGERING EVENT TRAVELS. CR 707.10 copies "any choices
//     made when the ability triggered", and the Strionic Resonator
//     rulings are explicit that the copy remembers the same event: a
//     copied "whenever this deals combat damage to a player, draw
//     that many cards" draws the same number. That is StackItem.Trigger
//     (#1223, trigger_event.go), and it is why the two halves of this
//     issue are one issue — before it, the event lived in the Effect
//     closure, where a copy reached it only by accident of sharing a
//     func pointer and a clause could not reach it at all.

// CopyAbilityForEffect creates a copy of the activated or triggered
// ability item `itemID` under `controller`'s control, per CR 707.10.
//
// `itemID` is a STACK ITEM id, not a card id, and the difference is
// the whole reason Event.StackItemID exists: an ability's source
// permanent stays on the battlefield and may have several of its
// abilities on the stack at once, so naming the card cannot say which
// one. Strionic Resonator gets the id from its target clause
// (AbilityOnStack, targets.go); Rings of Brighthearth gets it off the
// activation event.
//
// When `mayChooseNewTargets` is set and the ability actually has a
// target clause with at least one legal target on the current board,
// the controller is prompted first and the copy is created from their
// answer (CR 707.10c). Otherwise the copy is created immediately with
// the original's targets, which is also what CR 707.10c says happens
// when a player declines to change them.
//
// Errors: ErrCardNotFound when no such ability is on the stack — its
// controller may have had it countered in response, and a mana
// ability was never there at all (CR 605.3b). Both are normal
// outcomes and callers should treat them as "the copy effect did
// nothing". ErrInvalidParam when `itemID` names a SPELL, which is
// CopySpellForEffect's job and a different shape (a spell copy needs
// the card).
//
// Caller must hold g.mu.
func (g *Game) CopyAbilityForEffect(itemID, controller uuid.UUID, mayChooseNewTargets bool) error {
	item, ok := g.stackAbilityLocked(itemID)
	if !ok {
		return ErrCardNotFound
	}
	if item.Kind == StackItemSpell {
		return ErrInvalidParam
	}
	// The object the copy's targeting legality is judged from: the
	// SOURCE PERMANENT, because an ability's characteristics for
	// CR 702.16b are its source's (docs/decisions/0072-protection.md
	// §2), and a value copy of it because the permanent can leave
	// between the prompt and the answer — the copy still resolves
	// (CR 608.2), and it is judged on what the source was.
	src := g.abilitySourceCardLocked(item)
	g.offerCopyTargetsLocked(src, item, controller, item.targetSpec, mayChooseNewTargets)
	return nil
}

// stackAbilityLocked returns the ABILITY item for `itemID` — on the
// stack, or the one currently RESOLVING.
//
// The resolving branch is the ability twin of stackSpellLocked's
// (#920) and exists for the same reason: resolveTopAbilityLocked
// deletes the item from StackMeta before it runs the Effect, so an
// effect that copies the ability it is part of would find nothing.
// No printed card reaches it today — every ability-copy card in the
// catalog copies an ability that is still on the stack UNDER the one
// doing the copying — and the branch is here because the parked slot
// already has the answer (resolving_item.go's beginResolvingLocked
// says so in as many words) and because a reader who finds the hole
// later has no way to tell "not written" from "deliberately absent".
//
// Narrow on purpose, like the spell one: it answers only for the
// resolving item's OWN id, so a Strionic Resonator pointed at an
// ability that was countered in response still gets nothing.
//
// Caller must hold g.mu.
func (g *Game) stackAbilityLocked(itemID uuid.UUID) (*StackItem, bool) {
	if item := g.StackMeta[itemID]; item != nil {
		return item, true
	}
	if r := g.resolving; r != nil && r.item != nil && !r.hasCard && r.item.ID == itemID {
		return r.item, true
	}
	return nil, false
}

// abilitySourceCardLocked is a value copy of the permanent an ability
// came from, or a bare Card carrying just the controller when it can
// no longer be found.
//
// A bare Card is not a hole: an ability whose source has left still
// resolves (CR 608.2), and the copy's re-target prompt judges its
// picks against a source with no qualities — which is the same
// declared limitation stackItemSourceLocked already carries for the
// CR 608.2b re-check, written down in ADR 0072 §2.
//
// Caller must hold g.mu.
func (g *Game) abilitySourceCardLocked(item *StackItem) Card {
	if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
		return c
	}
	return Card{InstanceID: item.SourceCardID, Controller: item.Controller}
}

// createAbilityCopyLocked puts the copy of an ability on the stack.
// Caller must hold g.mu.
//
// What travels is CR 707.10's "the characteristics, plus the choices
// made when it was put on the stack": the kind, the source, the
// label, the resolution behaviour (Effect, and the specs the CR
// 608.2b re-check reads), the targets, the modes, X, the division,
// the reflexive payload, and the triggering event.
//
// What does NOT travel, and why each one is a decision rather than an
// omission:
//
//   - THE MANA. Nothing was spent to create a copy (CR 707.10, the
//     Dawnglow Infusion ruling), so Paid's mana half is left at its
//     zero value, and that zero is a real answer rather than a gap
//     (#761): NoManaSpent is true of a copy and should be. The rest
//     of the record — counters removed or added, life paid, the
//     optional costs chosen — DOES travel, because those are what the
//     resolution reads back as facts about the announcement ("for
//     each counter removed this way"), the same argument #664 makes
//     for carrying OptionalCosts onto a spell copy. A copy that
//     zeroed them would silently resolve those clauses for nothing,
//     which is a worse answer than the rule's.
//   - DoubledBy. CR 603.2d attribution names the permanent that
//     added an extra INSTANCE of a trigger; a copy was added by a
//     copy effect, which is a different sentence, and leaving the
//     field clear keeps the stack overlay honest about which is
//     which.
//   - Ordered. A copy is put straight onto the stack rather than
//     through the CR 603.3b APNAP queue, so the flag has nothing to
//     say about it.
//   - HoldPriority, CastFromZone, AltCost, Foretold, FaceDown,
//     AltCostExiles, SplitSecond. Every one of these is a fact about
//     a CAST, and none of them is meaningful on an ability item in
//     the first place.
func (g *Game) createAbilityCopyLocked(item *StackItem, controller uuid.UUID, targets []TargetRef) {
	if g.StackMeta == nil {
		g.StackMeta = make(map[uuid.UUID]*StackItem)
	}
	copyID := uuid.New()
	meta := &StackItem{
		ID:           copyID,
		Kind:         item.Kind,
		Controller:   controller,
		Owner:        controller,
		SourceCardID: item.SourceCardID,
		SourceEpoch:  item.SourceEpoch,
		Label:        item.Label,
		Targets:      append([]TargetRef(nil), targets...),
		Payload:      append([]TargetRef(nil), item.Payload...),
		Trigger:      cloneTriggerContext(item.Trigger),
		Modes:        append([]int(nil), item.Modes...),
		XValue:       item.XValue,
		Distribution: cloneDistributionLocked(item.Distribution),
		Paid:         copiedPaidCost(item.Paid),
		Effect:       item.Effect,
		targetSpec:   item.targetSpec,
		modeSpec:     item.modeSpec,
		IsCopy:       true,
		Seq:          g.nextStackSeqLocked(),
	}
	g.StackMeta[copyID] = meta

	// CR 707.10a, said by omission: no EventActivateAbility and no
	// EventTrigger. The copy was created, not activated and not
	// triggered, so "whenever you activate an ability", "whenever an
	// opponent activates an ability" and the CR 726 loop breaker's
	// player-decision notch all correctly see nothing here.
	//
	// EventBecomesTarget still fans out — the copy is an ability and
	// the things it points at have become the target of one
	// (CR 115.7), which is what a ward trigger or Monk Gyatso is
	// watching for, and what the ordinary activation path emits at
	// the same point in its own announcement.
	g.emitBecameTargetLocked(controller, meta.SourceCardID, meta.ID, meta.Targets)
}

// copiedPaidCost is the announcement record a CR 707.10 copy carries:
// everything the resolution reads back as a fact about the original's
// announcement, and none of the mana.
//
// See createAbilityCopyLocked's doc for the argument. OnPaper is
// deliberately left false alongside the empty Mana slice: "the engine
// waived a charge" is a statement about a payment that was made, and
// no payment was made here — the copy's record is KNOWN and known to
// be nothing, which is exactly what #761 built the distinction for.
func copiedPaidCost(p PaidCost) PaidCost {
	return PaidCost{
		CountersRemoved: p.CountersRemoved,
		CountersAdded:   p.CountersAdded,
		LifePaid:        p.LifePaid,
		OptionalCosts:   append([]int(nil), p.OptionalCosts...),
		// #759: a copy of a station ability reads the SAME tapped
		// creature the original does — what to tap was chosen when
		// the original was put on the stack, which is what CR 707.10
		// says a copy carries. Its own slice, so the exit hook's
		// rewrite of one record never reaches the other through a
		// shared backing array; both are rewritten, independently.
		TappedOthers: append([]PaidTap(nil), p.TappedOthers...),
	}
}

// AbilityItemOnStackForEffect reports whether `itemID` names an
// activated or triggered ability currently on the stack, and who
// controls it.
//
// The read a copy card's target clause and its bot enumeration both
// need, and the one that cannot be spelled as a card lookup: the id
// belongs to a StackItem, and the card it names is a permanent
// standing on the battlefield doing nothing.
//
// Caller must hold g.mu.
func (g *Game) AbilityItemOnStackForEffect(itemID uuid.UUID) (kind StackItemKind, controller uuid.UUID, ok bool) {
	item := g.StackMeta[itemID]
	if item == nil || item.Kind == StackItemSpell {
		return "", uuid.Nil, false
	}
	return item.Kind, item.Controller, true
}
