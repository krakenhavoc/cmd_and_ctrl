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
//     the copied spell happened to have flashback. A copy that leaves
//     the stack WITHOUT resolving — countered, bounced, exiled — is
//     ended the same way at the shared exit, routeCardToZoneLocked
//     (#1340, CR 707.10a): see spellCopyLeavesStackLocked below.
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
//  4. A COPY OF A PERMANENT SPELL BECOMES A TOKEN. CR 608.3f: when
//     it would resolve, no permanent card enters — a token that is a
//     copy of the spell does, and the copy ceases to exist. That is
//     resolvePermanentSpellCopyLocked at the bottom of this file, and
//     it goes through #923's one token-creation path, so a doubler
//     doubles it (it really is a created token, CR 111.13), entry
//     replacements apply, and the entry can pause and resume. Until
//     #666 the engine refused instead, on the reasoning that a second
//     card-shaped object carrying the original's oracle ID on the
//     battlefield was worse than nothing; the token is what makes it
//     neither.
//
// Copies of ABILITIES (CR 707.10 covers those too — Strionic
// Resonator, Lithoform Engine, Rings of Brighthearth) live next door
// in ability_copy.go (#1223). They share this file's PROMPT: the
// CR 707.10c "you may choose new targets for the copy" is one
// question with one frame and one submit half, and only the last step
// — build a new card on the stack, or build a new item beside it —
// differs. See copyFrame below.

// CopySpellForEffect creates a copy of the spell `spellID` under
// `controller`'s control, per CR 707.10.
//
// `spellID` may be a spell on the stack (Reverberate's target) or the
// spell currently RESOLVING (#920) — "that player may copy this
// spell", the Chain cycle's own clause, where the copy is created by
// the resolving spell's effect and controlled by the player the card
// says creates it (CR 707.10b), not by the copied spell's controller.
// See stackSpellLocked for how the second one is still findable.
//
// When mayChooseNewTargets is set and the copied spell actually has
// a target clause with at least one legal target on the current
// board, the controller is prompted first and the copy is created
// from their answer. Otherwise the copy is created immediately with
// the original's targets — which is also what CR 707.10c says
// happens when a player declines to change them.
//
// `except` is the card's "except …" clause (CR 707.10a): Double
// Major's "except it isn't legendary if the spell is legendary". It
// edits the COPIABLE VALUES the copy is created with, the same
// PrintedValues an entering permanent's except clause edits (ADR
// 0043), so the modification is on the copy from the moment it
// exists and is still on the token it becomes. Nil for a plain copy.
//
// It is applied HERE, before anything is queued, rather than carried
// into the re-target prompt: the continuation crosses a snapshot and
// a closure cannot, so what the frame carries is the already-edited
// card.
//
// Errors: ErrCardNotFound when the spell is no longer on the stack
// (its controller may have had it countered in response to the copy
// effect, which is a normal outcome and not an engine fault —
// callers should treat it as "the copy effect did nothing"). That is
// the CR 608.2b outcome for a copy effect that TARGETS the spell. One
// that merely names it — storm, Doublecast's "copy that spell" —
// copies from last-known information instead (CR 608.2h) and calls
// CopyLastKnownSpellForEffect (stack_lki.go, #1255).
//
// Caller must hold g.mu. Added in S30 (#95); `except` in S45 (#666).
func (g *Game) CopySpellForEffect(spellID, controller uuid.UUID, mayChooseNewTargets bool, except func(v *PrintedValues)) error {
	src, item, ok := g.stackSpellLocked(spellID)
	if !ok {
		return ErrCardNotFound
	}
	g.copySpellFromLocked(src, item, controller, mayChooseNewTargets, except)
	return nil
}

// copySpellFromLocked is everything CopySpellForEffect does once it has
// found the spell, shared with CopyLastKnownSpellForEffect (#1255),
// which differs only in WHERE it may find it. `src` is a value copy
// the caller owns; `item` is only read. Caller must hold g.mu.
func (g *Game) copySpellFromLocked(src Card, item *StackItem, controller uuid.UUID, mayChooseNewTargets bool, except func(v *PrintedValues)) {
	if except != nil {
		v := CopiableValuesOf(src)
		except(&v)
		// `src` is already a value copy off the stack, so this can
		// never reach the spell being copied.
		src.setPrintedValues(v)
	}
	spec := castTargetSpecForItem(CatalogKey(src), item)
	g.offerCopyTargetsLocked(src, item, controller, spec, mayChooseNewTargets)
}

// offerCopyTargetsLocked is CR 707.10c, shared by the spell copy
// above and the ability copy in ability_copy.go: open the "you may
// choose new targets for the copy" prompt, or create the copy now
// with the original's targets.
//
// `src` is the object the copy's legality is judged from — the copied
// SPELL for a spell copy (CR 707.10: the copy has the original's
// characteristics, so protection is tested against the original's
// colour and type), and the ability's SOURCE PERMANENT for an ability
// copy (CR 702.16b tests an ability's quality against the permanent
// it came from). `item` decides which copy gets built on the far
// side, off its Kind, which is the only thing the two paths do
// differently.
//
// Creating the copy NOW is the right answer in three cases and they
// are all the printed outcome rather than a shortcut: the card does
// not offer the choice, the original named no target anybody chose,
// and nothing on the board qualifies any more — CR 707.10c's choice
// is optional, and a copy that keeps an illegal target is countered
// by game rules on resolution, which is what the card does in paper.
// The alternative is a prompt with no answers, which since #791 is a
// table that cannot move.
//
// Caller must hold g.mu.
func (g *Game) offerCopyTargetsLocked(src Card, item *StackItem, controller uuid.UUID, spec *TargetSpec, mayChooseNewTargets bool) {
	if !mayChooseNewTargets || spec == nil || !itemHasChosenTarget(item) {
		g.createCopyLocked(src, item, controller, item.Targets)
		return
	}
	lt := g.legalTargetsLocked(SourceObject(controller, &src), spec)
	if len(lt.Players) == 0 && len(lt.Cards) == 0 {
		g.createCopyLocked(src, item, controller, item.Targets)
		return
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
		copyResume: &copyFrame{
			src:        src,
			item:       *item,
			controller: controller,
			spec:       spec,
		},
	})
}

// copyFrame is the continuation for the CR 707.10c "you may choose
// new targets" prompt, for a copy of a spell OR of an ability. It
// carries VALUE copies of the source and its stack item,
// deliberately: between the prompt and the answer the original can be
// countered, and the copy is unaffected by that (CR 707.10 — the
// copy's characteristics are locked in when it is created, and a copy
// of a countered spell still resolves).
//
// `item.Kind` is the discriminator, and it is the only one needed.
// `src` means slightly different things on the two paths and both
// readings are the object CR 702.16b tests targeting against: for a
// SPELL copy it is the copied spell itself, for an ABILITY copy the
// permanent the ability came from. Nothing between here and
// createCopyLocked has to know which.
type copyFrame struct {
	src        Card
	item       StackItem
	controller uuid.UUID
	spec       *TargetSpec
}

// createCopyLocked puts the copy on the stack: a new spell object for
// a spell, a new item beside the source permanent for an ability.
//
// The ONE branch in the copy path, and it is here rather than at each
// call site so that everything before it — the re-target offer, the
// prompt, the submit gate, the #809 refresh that re-reads a frozen
// legal set — is written once for both shapes.
//
// Caller must hold g.mu.
func (g *Game) createCopyLocked(src Card, item *StackItem, controller uuid.UUID, targets []TargetRef) {
	if item.Kind == StackItemSpell {
		g.createSpellCopyLocked(src, item, controller, targets)
		return
	}
	g.createAbilityCopyLocked(item, controller, targets)
}

// stackSpellLocked returns the card and stack item for a spell on the
// stack — or for the spell currently RESOLVING, which is the same
// object one step later in its life (#920).
//
// The resolving branch is CR 707.10's "copy this spell", the Chain
// cycle's clause. A resolving spell is off the stack by every measure
// this function used to take: its meta is out of StackMeta and, once
// the copy decision pauses on a prompt, its card is out of the Stack
// zone and in a graveyard. In the RULES it is still there — a spell is
// put into its owner's graveyard as the final step of its own
// resolution (CR 608.2m) — so the copy is made from last-known
// information, which is what Game.resolving holds.
//
// It is narrow on purpose: resolvingSpellLocked answers only for the
// resolving item's own ID, so Reverberate pointed at a spell that was
// countered in response still gets ErrCardNotFound, which is the
// outcome its callers are written for.
//
// Caller must hold g.mu.
func (g *Game) stackSpellLocked(spellID uuid.UUID) (Card, *StackItem, bool) {
	if g.Stack != nil {
		for _, c := range g.Stack.Cards {
			if c.InstanceID != spellID {
				continue
			}
			item, ok := g.StackMeta[spellID]
			if !ok || item == nil {
				break
			}
			return c, item, true
		}
	}
	return g.resolvingSpellLocked(spellID)
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
	// Paid's MANA half is left at its zero value for the same reason,
	// and the zero value is a REAL answer rather than a gap (#761):
	// mana is not an object, so nothing was spent to cast the copy
	// (the Dawnglow Infusion ruling under CR 707.10). A copied Vexing
	// Bauble trigger counters the copy, a copied converge spell
	// converges for nothing, and both are correct.
	//
	// #664 is the one part of the record that DOES travel. CR 707.10
	// copies the choices made when the spell was cast, and "was it
	// kicked" is one of them — a copy of a kicked Rite of Replication
	// is kicked, the same way it keeps the modes and the X above.
	// Copied through paidWithOptionalCosts' output rather than the
	// wire list so a copy of a copy stays stable.
	if len(item.Paid.OptionalCosts) > 0 {
		meta.Paid.OptionalCosts = append([]int(nil), item.Paid.OptionalCosts...)
	}
	// ADR 0089 §2: and so is "was the gift promised, and to whom" —
	// the gift cost is paid by that choice (CR 702.174a), and a copy
	// of a promised spell gives its gift to the same opponent.
	meta.Paid.GiftOpponent = item.Paid.GiftOpponent
	g.StackMeta[copyCard.InstanceID] = meta
	g.recomputeSplitSecondLocked()

	// No EventCast: a copy is created, not cast. EventBecomesTarget
	// still fans out — the copy is a spell and the things it points
	// at have become the target of one (CR 115.7), which is what a
	// ward trigger or Monk Gyatso is watching for.
	g.emitBecameTargetLocked(controller, copyCard.InstanceID, copyCard.InstanceID, meta.Targets)
}

// resolveCopyTargetsLocked is the submit half of the CR 707.10c
// re-target prompt, reached from ResolvePickTargets when the choice
// carries a copyFrame. Validates the refs against the copied object's
// own clause — the same spec the original was announced under — then
// creates the copy on top of the stack.
//
// #1196: the check is the CR 115.7 one (retargetCheckLocked), not the
// announce gate. "You may choose new targets for the copy" is CR
// 707.10c, and CR 707.10c is CR 115.7c by reference — so a target the
// player LEFT ALONE may stay even if it has since become illegal,
// and the NUMBER of targets may not change. validateTargetsLocked
// had both backwards: it refused the first and allowed the second.
// What is still this file's own is the APPLICATION — a copy builds a
// new object rather than rewriting one, because its characteristics
// are a snapshot and the original may be gone.
//
// Caller must hold g.mu and must have located the choice at `idx`.
func (g *Game) resolveCopyTargetsLocked(idx int, cf *copyFrame, targets []TargetRef) error {
	for _, t := range targets {
		if t.Kind != TargetPlayer && t.Kind != TargetCard {
			return ErrInvalidParam
		}
	}
	// A SNAPSHOT, deliberately: the original spell can be countered
	// between the prompt and the answer, and the copy's
	// characteristics are the original's as they were (CR 707.10).
	// cf.src is the value copy the frame kept for exactly this.
	src := SourceSnapshot(cf.controller, SourceCharacteristics(&cf.src))
	steps := AnnouncedClauses(cf.spec, nil, nil)
	stamped := assignAnnouncedSlots(steps, targets)
	if err := g.retargetCheckLocked(src, steps, cf.item.Targets, stamped, RetargetChooseNew); err != nil {
		return err
	}
	targets = stamped
	g.dequeueChoiceLocked(idx)
	item := cf.item
	g.createCopyLocked(cf.src, &item, cf.controller, targets)
	g.runStateChecksLocked()
	return nil
}

// resolvePermanentSpellCopyLocked is CR 608.3f: "if a copy of a
// spell is in a zone other than the stack, it ceases to exist … if
// the copy is of a permanent spell, it becomes a token as it
// resolves."
//
// Three things fall out of doing it this way rather than by pushing
// the copy onto the battlefield:
//
//   - No card-shaped object carrying the original's oracle ID ever
//     reaches the battlefield. A copy is not a card (CR 707.10), so a
//     bounce spell cannot turn it into a card in somebody's hand, and
//     the CR 704.5d existence check removes it from any zone but the
//     battlefield (token_existence.go) — both for free, off
//     `Card.IsToken`, because the template carries the Token
//     supertype.
//   - It really is a CREATED token (CR 111.13), so it goes through
//     #923's one creation path: the CR 701.7b window opens, Doubling
//     Season doubles it, Academy Manufactor could rewrite it, and
//     every token then takes the ordinary battlefield entry —
//     enters-tapped, enters-with-counters, `fireETBHookLocked` and a
//     resumable pause included.
//   - The copy's own modifications ride along. The token is a copy of
//     the SPELL, and the spell on the stack already carries whatever
//     its "except" clause said (Double Major's "it isn't legendary"),
//     because CopySpellForEffect wrote those values into it when the
//     copy was created.
//
// The copy leaves the stack FIRST. The creation can pause on a CR 616
// ordering prompt, and a copy still sitting on the stack whose
// StackMeta the resolver has already deleted is an object nothing can
// resolve.
//
// Caller must hold g.mu.
func (g *Game) resolvePermanentSpellCopyLocked(top Card, item *StackItem) error {
	g.ceaseToExistLocked(top.InstanceID)
	tmpl := tokenCopyOfSpell(top)
	// CR 707.10b / CR 400.7d: the optional additional costs the copy
	// "paid" travel onto the permanent it becomes, exactly as they do
	// on the ordinary permanent-resolution branch — a Double Major on
	// a KICKED Wolfbriar Elemental makes a token that counts the
	// kicks. This is NOT a copiable value and deliberately not on
	// PrintedValues: it is cast-time state stamped at entry, and
	// #988's createSpellCopyLocked is what put it on the copy's item
	// in the first place. mintTokenLocked does not clear it, so the
	// template is the right place to carry it.
	if len(item.Paid.OptionalCosts) > 0 {
		tmpl.Provenance.OptionalCosts = append([]int(nil), item.Paid.OptionalCosts...)
	}
	tmpl.Provenance.GiftOpponent = item.Paid.GiftOpponent
	return g.CreateTokensThenForEffect(TokenCreation{
		Controller: item.Controller,
		Groups:     []TokenGroup{{Template: tmpl, Count: 1}},
		Source:     item.SourceCardID,
	}, nil)
}

// tokenCopyOfSpell builds the token template a resolving copy of a
// permanent spell becomes: the spell's copiable values (CR 707.2),
// plus the Token supertype.
//
// The face is settled first, for the same reason the ordinary
// permanent-resolution branch settles it (ADR 0034): an MDFC or a
// transform card resolves as the face that was cast, and an adventure
// resolves front-up. Copying Layout and Faces across is CR 707.10g —
// a copy of a double-faced permanent spell is double-faced too.
//
// The card-carried ability slices come across for the same reason
// applyCopy carries them: they are the only source of truth when the
// thing being copied is itself a token.
func tokenCopyOfSpell(src Card) Card {
	src.SetFace(faceOnResolve(src.Layout, src.ActiveFace))
	v := CopiableValuesOf(src)
	v.MakeToken()
	var tok Card
	tok.setPrintedValues(v)
	tok.ManaAbilities = append([]ManaAbilityShape(nil), src.ManaAbilities...)
	tok.ActivatedAbilities = append([]ActivatedAbilityShape(nil), src.ActivatedAbilities...)
	return tok
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

// stackCopyLocked reports whether the object `cardID` on the stack is
// a COPY of a spell (StackItem.IsCopy) — whether it is still waiting on
// the stack (StackMeta) or is the item currently resolving, whose meta
// the resolver has already taken out of StackMeta while the card is
// still standing on the stack (#920's slot). The second case is a copy
// whose own effect moves it ("shuffle this spell into its owner's
// library"): it leaves the stack by the same route a counterspell
// uses, and it is no more a card for having done it itself.
//
// Caller must hold g.mu.
func (g *Game) stackCopyLocked(cardID uuid.UUID) bool {
	if item, ok := g.StackMeta[cardID]; ok && item != nil {
		return item.IsCopy
	}
	_, item, ok := g.resolvingSpellLocked(cardID)
	return ok && item.IsCopy
}

// spellCopyLeavesStackLocked is the exit route's answer for a COPY of a
// spell (#1340): countered, returned to hand, exiled, tucked, or moved
// by hand in the sandbox, it goes nowhere. CR 707.10a: "If a copy of a
// spell is in a zone other than the stack, it ceases to exist"; CR
// 704.5e says the same as a state-based action. The engine never lets
// the copy reach the other zone at all, rather than landing it there
// and sweeping it a beat later, because nothing may observe it in
// between — a copy is not a card, so no "put into a graveyard from
// anywhere" trigger, no Tarmogoyf, no Regrowth and no flashback may
// ever see it.
//
// What it owes is the route's EVENT, minus the landing:
//
//   - a counterspell still COUNTERED it, so a Countered route emits
//     EventCounterSpell exactly as it does for a card (CR 701.6a — the
//     counter happened; only the destination is missing), and every
//     "whenever a spell is countered" watcher sees it;
//   - every other route emits the ceasing-to-exist shape
//     ceaseToExistLocked and prepareCopyCeasesLocked already use: an
//     EventZoneMove out of the stack with NO NewZone. Never a move to a
//     graveyard, a hand or exile — the copy was never there.
//
// The stack record goes with it, as it does on every stack exit
// (#1318). The last-known record has already been taken by the caller
// (routeCardToZoneLocked), so a copy that is countered can itself still
// be copied from last-known information by an effect that names it
// (CR 608.2h).
//
// Caller must hold g.mu.
func (g *Game) spellCopyLeavesStackLocked(cardID uuid.UUID, r zoneRoute) {
	if g.Stack == nil {
		return
	}
	if _, err := g.Stack.Remove(cardID); err != nil {
		return
	}
	if _, ok := g.StackMeta[cardID]; ok {
		delete(g.StackMeta, cardID)
		g.recomputeSplitSecondLocked()
	}
	var out Event
	if r.Countered {
		out = Event{Kind: EventCounterSpell, Target: cardID, CardID: cardID}
	} else {
		out = Event{
			Kind:    EventZoneMove,
			Actor:   r.Actor,
			Source:  r.Source,
			CardID:  cardID,
			OldZone: ZoneStack,
		}
	}
	r.Cause.stampCause(&out)
	g.EmitEvent(out)
	// The two prunes every landed exit runs: a prompt that still offers
	// the copy as a candidate, or asks about another move of it, is
	// asking about an object that no longer exists.
	g.pruneCardSetChoicesLocked()
	g.pruneStaleZoneChangeChoicesLocked()
}
