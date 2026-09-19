package game

import "github.com/google/uuid"

// spell_copy.go — CR 707.10, copying a spell on the stack.
//
// "Copy target instant or sorcery spell. You may choose new targets
// for the copy." Reverberate, Twincast, Doublecast, Increasing
// Vengeance, Fork, and every storm card in the format run through
// this one primitive.
//
// Three facts about a copy drive the whole design, and each of them
// is a thing the engine would otherwise get wrong by default:
//
//  1. A COPY IS NOT A CAST. CR 707.10 creates it directly on the
//     stack; it was never announced, no cost was paid, and nothing
//     that watches EventCast may fire. That falls out of this file
//     emitting no EventCast — not out of a suppression check
//     somewhere downstream. (A cheaper implementation that reached
//     for CastSpell and skipped the cost would trigger every
//     Storm-Kiln Artist and Aetherflux Reservoir at the table.)
//
//  2. A COPY IS NOT A CARD. CR 707.10 again. When it finishes
//     resolving it ceases to exist rather than going to a graveyard.
//     The stack's ordinary instant/sorcery exit is
//     routeStackCardToGraveyardLocked, so resolveTopOfStackLocked
//     branches on StackItem.IsCopy — see ceaseToExistLocked below.
//     Without the branch, casting Twincast on your own Lightning
//     Bolt would leave two Bolts in your graveyard: countable by
//     Tarmogoyf, returnable by Regrowth, and flashback-castable if
//     the copied spell happened to have flashback.
//
//  3. THE TARGETS ARE RE-CHOSEN BEFORE IT LANDS. "You may choose new
//     targets for the copy" is resolved by the copy's controller as
//     the copy is created, not afterwards, so the prompt is queued
//     BEFORE anything reaches the stack and the copy is built in the
//     resume. That ordering also gives the copy the top of the stack
//     for free: an open pending choice suppresses every other move, so
//     nothing can be added or resolved while the prompt is waiting.
//
//     That suppression lives where it does for every other prompt kind
//     — legal.anyChoiceOpen for the bot, and the client's own gate —
//     rather than inside PassPriority, which has never carried the
//     check. A test that drives PassPriority directly can therefore
//     resolve the copied spell out from under an unanswered prompt and
//     build the copy onto an empty stack. It is a property of the raw
//     method shared with the scry, search and sacrifice prompts, not
//     something this file introduces; it is written down because for a
//     copy the symptom is a wrong ORDER rather than a stuck prompt,
//     which is the harder one to notice.
//
// What this deliberately does not do:
//
//   - Copies of PERMANENT spells. CR 608.3f makes a resolving copy
//     of a permanent spell a TOKEN, and this engine has no
//     token-from-stack-item path. Every card in the S30 batch says
//     "target instant or sorcery spell", so the restriction is
//     enforced by the cards' TargetSpec rather than by a check here;
//     a future card that copies a creature spell needs the token
//     rule before it ships (#666).
//   - Copies of ABILITIES (CR 707.10 covers those too — Lithoform
//     Engine, Strionic Resonator). Ability items carry a resolution
//     closure rather than a card, so they are a different copy
//     shape; nothing in S30 needs one.

// CopySpellForEffect creates a copy of the spell `spellID` under
// `controller`'s control, per CR 707.10.
//
// When mayChooseNewTargets is set and the copied spell actually has
// a target clause with at least one legal target on the current
// board, the controller is prompted first and the copy is created
// from their answer. Otherwise the copy is created immediately with
// the original's targets — which is also what CR 707.10c says
// happens when a player declines to change them.
//
// Errors: ErrCardNotFound when the spell is no longer on the stack
// (its controller may have had it countered in response to the copy
// effect, which is a normal outcome and not an engine fault —
// callers should treat it as "the copy effect did nothing").
//
// Caller must hold g.mu. Added in S30 (#95).
func (g *Game) CopySpellForEffect(spellID, controller uuid.UUID, mayChooseNewTargets bool) error {
	src, item, ok := g.stackSpellLocked(spellID)
	if !ok {
		return ErrCardNotFound
	}
	spec := castTargetSpecForItem(CatalogKey(src), item)
	if !mayChooseNewTargets || spec == nil || !itemHasChosenTarget(item) {
		g.createSpellCopyLocked(src, item, controller, item.Targets)
		return nil
	}
	// CR 707.10c — the choice is optional, and it is also
	// impossible when nothing on the board qualifies any more. In
	// that case the copy keeps the original's targets and is
	// countered by game rules on resolution, which is the printed
	// outcome rather than a wedged prompt.
	// The copy's source is the copied SPELL (CR 707.10 — the copy has
	// its characteristics), not the player making the copy, so
	// protection is tested against the original's colour and type.
	lt := g.legalTargetsLocked(SourceObject(controller, &src), spec)
	if len(lt.Players) == 0 && len(lt.Cards) == 0 {
		g.createSpellCopyLocked(src, item, controller, item.Targets)
		return nil
	}
	label := spec.Label
	if label == "" {
		label = "targets"
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:              PendingChoicePickTarget,
		Chooser:           controller,
		Count:             1,
		Source:            src.InstanceID,
		Reason:            "Choose new " + label + " for the copy (or re-pick the same)",
		PickTargetPlayers: lt.Players,
		PickTargetCards:   lt.Cards,
		PickTargetMin:     spec.Min,
		PickTargetMax:     spec.Max,
		copySpellResume: &copySpellFrame{
			src:        src,
			item:       *item,
			controller: controller,
			spec:       spec,
		},
	})
	return nil
}

// copySpellFrame is the continuation for the CR 707.10 "you may
// choose new targets" prompt. It carries VALUE copies of the source
// card and its stack item, deliberately: between the prompt and the
// answer the original spell can be countered, and the copy is
// unaffected by that (CR 707.10 — the copy's characteristics are
// locked in when it is created, and a copy of a countered spell
// still resolves).
type copySpellFrame struct {
	src        Card
	item       StackItem
	controller uuid.UUID
	spec       *TargetSpec
}

// stackSpellLocked returns the card and stack item for a spell
// currently on the stack. Caller must hold g.mu.
func (g *Game) stackSpellLocked(spellID uuid.UUID) (Card, *StackItem, bool) {
	if g.Stack == nil {
		return Card{}, nil, false
	}
	for _, c := range g.Stack.Cards {
		if c.InstanceID != spellID {
			continue
		}
		item, ok := g.StackMeta[spellID]
		if !ok || item == nil {
			return Card{}, nil, false
		}
		return c, item, true
	}
	return Card{}, nil, false
}

// itemHasChosenTarget reports whether the item names at least one
// real target. TargetSelf and TargetNone slots are not choices
// anybody made, so a spell carrying only those gets no re-target
// prompt.
func itemHasChosenTarget(item *StackItem) bool {
	for _, t := range item.Targets {
		if t.Kind == TargetCard || t.Kind == TargetPlayer {
			return true
		}
	}
	return false
}

// createSpellCopyLocked puts the copy on the stack. Caller must hold
// g.mu.
//
// The copy is a new object with a fresh InstanceID carrying the
// original's COPIABLE VALUES (CR 707.2) — printed name, type line,
// mana cost, oracle ID and faces — which is exactly the value copy
// of the Card minus its identity and zone bookkeeping. The oracle ID
// riding along is what makes the copy resolve: the catalog's
// OnResolve is dispatched by oracle ID, so the copy runs the same
// effect the original will.
//
// Owner is set to the copy's controller rather than the original
// owner. A copy has no owner in the rules (it is not a card) and the
// field is only read by the graveyard routing this file's
// ceaseToExistLocked bypasses — but pointing it at the controller
// means that if a future path ever does route it, it does not
// silently hand a card to a player who never had one.
func (g *Game) createSpellCopyLocked(src Card, item *StackItem, controller uuid.UUID, targets []TargetRef) {
	copyCard := src
	copyCard.InstanceID = uuid.New()
	g.noteCreatedSourceLocked(copyCard.InstanceID)
	copyCard.Owner = controller
	copyCard.Controller = controller
	copyCard.KnownBy = nil
	copyCard.Counters = nil
	copyCard.LostLastCounter = false
	copyCard.AttachedTo = TargetRef{}
	copyCard.effective = nil
	g.Stack.PushTop(copyCard)
	// A spell on the stack is public information, copy or not.
	g.markCardKnownInZoneLocked(g.Stack, copyCard.InstanceID)

	meta := &StackItem{
		ID:           copyCard.InstanceID,
		Kind:         StackItemSpell,
		Controller:   controller,
		Owner:        controller,
		SourceCardID: copyCard.InstanceID,
		Targets:      append([]TargetRef(nil), targets...),
		Modes:        append([]int(nil), item.Modes...),
		XValue:       item.XValue,
		Distribution: cloneDistributionLocked(item.Distribution),
		AltCost:      item.AltCost,
		SplitSecond:  item.SplitSecond,
		IsCopy:       true,
		Seq:          g.nextStackSeqLocked(),
	}
	// CastFromZone is deliberately left empty. The copy was not cast
	// from anywhere (CR 707.10), so Wash Away's "target spell that
	// wasn't cast from its owner's hand" reads it as exactly that.
	//
	// Paid is left at its zero value for the same reason, and the
	// zero value is a REAL answer rather than a gap (#761): mana is
	// not an object, so nothing was spent to cast the copy (the
	// Dawnglow Infusion ruling under CR 707.10). A copied Vexing
	// Bauble trigger counters the copy, a copied converge spell
	// converges for nothing, and both are correct.
	g.StackMeta[copyCard.InstanceID] = meta
	g.recomputeSplitSecondLocked()

	// No EventCast: a copy is created, not cast. EventBecomesTarget
	// still fans out — the copy is a spell and the things it points
	// at have become the target of one (CR 115.7), which is what a
	// ward trigger or Monk Gyatso is watching for.
	g.emitBecameTargetLocked(controller, copyCard.InstanceID, copyCard.InstanceID, meta.Targets)
}

// resolveCopySpellTargetsLocked is the submit half of the CR 707.10
// re-target prompt, reached from ResolvePickTargets when the choice
// carries a copySpellFrame. Validates the refs against the copied
// spell's own clause — the same spec the original was announced
// under — then creates the copy on top of the stack.
//
// Caller must hold g.mu and must have located the choice at `idx`.
func (g *Game) resolveCopySpellTargetsLocked(idx int, cf *copySpellFrame, targets []TargetRef) error {
	for _, t := range targets {
		if t.Kind != TargetPlayer && t.Kind != TargetCard {
			return ErrInvalidParam
		}
	}
	// A SNAPSHOT, deliberately: the original spell can be countered
	// between the prompt and the answer, and the copy's
	// characteristics are the original's as they were (CR 707.10).
	// cf.src is the value copy the frame kept for exactly this.
	if err := g.validateTargetsLocked(SourceSnapshot(cf.controller, SourceCharacteristics(&cf.src)), cf.spec, targets); err != nil {
		return err
	}
	g.dequeueChoiceLocked(idx)
	item := cf.item
	g.createSpellCopyLocked(cf.src, &item, cf.controller, targets)
	g.runStateChecksLocked()
	return nil
}

// ceaseToExistLocked removes a resolved or countered COPY from the
// stack without sending it anywhere (CR 707.10 — a copy is not a
// card, so it has no owner's graveyard to go to).
//
// Emits an EventZoneMove with an empty NewZone so the client's stack
// view drops the item and the event log shows what happened; nothing
// in the engine watches for a move to nowhere, which is the point.
//
// Caller must hold g.mu.
func (g *Game) ceaseToExistLocked(cardID uuid.UUID) {
	if g.Stack == nil {
		return
	}
	if _, err := g.Stack.Remove(cardID); err != nil {
		// Already gone — a copy that was countered and cleaned up by
		// some other path. Nothing to do, and not an error: "it no
		// longer exists" is the postcondition either way.
		return
	}
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		CardID:  cardID,
		OldZone: ZoneStack,
	})
}
