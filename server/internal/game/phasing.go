package game

import "github.com/google/uuid"

// phasing.go — CR 702.26, the permanent that is treated as though it
// does not exist. ADR 0084, #1199.
//
// # The shape, in one sentence
//
// A phased-out permanent is moved out of g.Battlefield.Cards and into
// g.PhasedOut.Cards, and NOTHING ELSE HAPPENS.
//
// That is the whole design, and both halves of the sentence carry
// weight.
//
// # Why it leaves the slice
//
// CR 702.26b: "Except for rules and effects that specifically mention
// phased-out permanents, a phased-out permanent is treated as though
// it does not exist. It can't affect or be affected by anything else
// in the game."
//
// "Anything else in the game" is a statement about every read of the
// battlefield, and this engine has 125 of them open-coded in
// internal/game, internal/legal, internal/protocol and internal/aiseat,
// another 189 in the card catalog, and 144 callers of eight
// near-identical single-card lookups. #1199 proposed one predicate at
// each walk. That is the right rule and the wrong mechanism at this
// scale: a walk that forgets to ask is a silent rules bug — a
// phased-out creature that still blocks, a Wrath that still kills it,
// a Lord that still counts it — and no test can notice the omission.
//
// A permanent that is not in the slice cannot be missed by a walk over
// the slice. Targeting, the layer pass, the trigger harvester, the SBA
// sweep, combat, the autotapper, the legal-move enumerator, the view
// builder and every bot heuristic are correct here with no edit and
// no knowledge of phasing at all. This is leave_game.go's argument for
// CR 800.4a ("leaving the game is not a zone change") and ADR 0036
// §12's argument against a reverse attachment index: a second place to
// ask is a second place to be wrong.
//
// # Why nothing else happens
//
// CR 702.26d: "The phasing event doesn't actually cause a permanent to
// change zones or control […] Zone-change triggers don't trigger when
// a permanent phases in or out. Tokens continue to exist on the
// battlefield while phased out. Counters and stickers remain on a
// permanent while it's phased out."
//
// So the two functions below call NEITHER MoveCard, NOR
// battlefieldExitLocked, NOR pruneChoicesAfterArrivalLocked, and emit
// NO EventZoneMove / EventLTB / EventETB. In particular
// Card.ObjectEpoch is untouched — the promise activation_tally.go has
// been making since S38, which is what keeps an exhaust ability spent
// across a phase cycle where a flicker refreshes it.
//
// MoveCard's battlefield-exit block is the negative image of this
// rule. Every field it clears — Tapped, NextUntapSkips, Counters,
// LostLastCounter, BattleX/Y, the combat targets, GoadedBy,
// DamageMarked, RegenerationShields, AttachedTo, AttachedAt,
// BaseController, NamedTribe, ChosenColor, ChosenPlayer, ChosenName,
// Provenance, ClassLevel, Solved, ProtectorPlayerID,
// EnteredBattlefieldAt, SummonedThisTurn and the face-down state — is
// a field CR 702.26d says to keep. Reading that block is the clearest
// statement of why phasing must not travel through it.
//
// # What is NOT built (ADR 0084 Decision 8)
//
// CR 702.26e and 702.26f are the continuous-effect corners: a
// phased-out permanent is excluded from the SET of objects a resolved
// continuous effect affects, and a "for as long as" duration that
// tracks it ends because it can no longer see it. The slice move gives
// the common case for free — while the permanent is out of the
// battlefield slice the layer pass cannot see it, so it contributes
// nothing and receives nothing. The corner is the RETURN: a "gain
// control until end of turn" or a "+3/+3 until end of turn" whose
// object phases out keeps its ScopedStatic registration and applies
// again if the permanent phases back in inside the duration. No
// catalogued card reaches it.

// ZonePhasedOut names the holding slice for phased-out permanents.
//
// IT IS A LABEL, NOT A ZONE IN THE CR 400 SENSE. Phasing is not a zone
// change (CR 702.26d) and a phased-out permanent is still on the
// battlefield as far as the rules are concerned. What the kind buys is
// that cloneZone, snapshotZone / restoreZone and the view's
// viewOfZone work on g.PhasedOut unchanged.
//
// findCardZoneLocked deliberately does not look here, so
// FindCardZoneForEffect answers nil for a phased-out permanent — the
// right answer to "where is this object" when the game is treating it
// as though there is no object. Nothing routes a card here through
// MoveCard, zoneFromRefLocked or executeZoneRouteLocked; the two
// functions in this file are the only writers.
const ZonePhasedOut ZoneKind = "phased_out"

// KeywordPhasing is CR 702.26's token, in the canonicalKeywords table
// (keywords.go) because this file is the consumer that honours it —
// the phase-OUT half of CR 502.1's turn-based action. The deck
// importer stamps it from Scryfall's keywords array like every other
// canonical token, so a printed-phasing permanent needs no catalog
// entry at all.
const KeywordPhasing = "phasing"

// phaseOutOptions carries the two riders a "phases out until …" card
// prints. The zero value is ordinary phasing: out now, in during the
// controller's next untap step (CR 502.1).
type phaseOutOptions struct {
	// LockedBy names an object whose presence on the battlefield is
	// the phase-out's duration — Oubliette's "phases out until this
	// enchantment leaves the battlefield", Out of Time's "phase out
	// until this enchantment leaves the battlefield".
	//
	// A permanent phased out this way does NOT phase in at its
	// controller's untap step. It phases in the moment the named
	// object stops being on the battlefield, which is
	// sweepPhaseInLocksLocked below. That is what the cards do: Out of
	// Time's whole point is that every creature comes back at once
	// when the last time counter is removed, mid-turn, not on
	// somebody's next untap step.
	//
	// A uuid rather than an ADR 0063 Duration because presence is the
	// condition and there is nothing else in the vocabulary to say.
	// A full Duration on every Card would have been ten fields for
	// one printed phrase.
	LockedBy uuid.UUID

	// TapOnPhaseIn is Oubliette's "Tap that creature as it phases in
	// this way". A rider on ONE phase-in, consumed as it fires.
	//
	// A bool on the card rather than a delayed trigger because the
	// phase-in it rides on is a turn-based action or a duration
	// ending — neither uses the stack (CR 702.26a), so there is
	// nothing for a trigger to sit on.
	TapOnPhaseIn bool
}

// phaseOutSetLocked expands a list of named permanents into the full
// set that phases out with them, and says which of them phase out
// INDIRECTLY.
//
// CR 702.26g: "When a permanent phases out, any Auras, Equipment, or
// Fortifications attached to that permanent phase out at the same
// time. This alternate way of phasing out is known as phasing out
// 'indirectly'." It is transitive — an Equipment on the creature, and
// an Aura on that Equipment — so the walk is a worklist rather than
// one call to AttachmentsOf. ADR 0036 keeps one direction only and no
// reverse index, which is why each level costs a scan.
//
// CR 702.26h: "If an object would simultaneously phase out directly
// and indirectly, it just phases out indirectly." So the indirect mark
// is set whenever an object is REACHED through an attachment, even if
// it was also named — which is why the marking happens here, once for
// the whole batch, rather than per card at the call site.
//
// Returns the set in a stable order (named permanents in the order
// given, then each level of attachments in battlefield order) and the
// indirect marks. Caller must hold g.mu.
func (g *Game) phaseOutSetLocked(ids []uuid.UUID) (order []uuid.UUID, indirect map[uuid.UUID]bool) {
	if g.Battlefield == nil || len(ids) == 0 {
		return nil, nil
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	indirect = make(map[uuid.UUID]bool)
	var frontier []uuid.UUID
	for _, id := range ids {
		if id == uuid.Nil || seen[id] || findCardOnBattlefield(g, id) < 0 {
			continue
		}
		seen[id] = true
		order = append(order, id)
		frontier = append(frontier, id)
	}
	// Breadth-first over the attachment tree. `seen` is both the
	// visited set and the result set, so a cycle (which the attach
	// rules forbid but the data model permits) terminates.
	for len(frontier) > 0 {
		var next []uuid.UUID
		for _, host := range frontier {
			for _, att := range g.AttachmentsOf(host) {
				// Set BEFORE the seen check: CR 702.26h applies to a
				// permanent that was named as well as attached.
				indirect[att] = true
				if seen[att] {
					continue
				}
				seen[att] = true
				order = append(order, att)
				next = append(next, att)
			}
		}
		frontier = next
	}
	return order, indirect
}

// phaseOutLocked is the ONE way a permanent phases out.
//
// The set is chosen and expanded up front and the whole batch moves
// before anything is emitted, for the two reasons that agree:
// CR 702.26a's "This all happens simultaneously", and the reason
// untapStepSetLocked chooses its set up front — a move emits, an emit
// runs listeners, and a listener is not something to run while holding
// a *Card pointer into g.Battlefield.Cards.
//
// Returns nothing, unlike its twin below: every caller is an
// instruction rather than a question ("those permanents phase out"),
// and a permanent that was named but is no longer on the battlefield
// is silently skipped — CR 608.2b's per-slot existence re-check, which
// is not news to anybody. phaseInLocked has a return because its
// caller (sweepPhaseInLocksLocked) owes the SBA pass a "something
// happened" answer.
//
// Caller must hold g.mu in write mode.
func (g *Game) phaseOutLocked(source uuid.UUID, ids []uuid.UUID, opts phaseOutOptions) {
	if g.Battlefield == nil || g.PhasedOut == nil {
		return
	}
	order, indirect := g.phaseOutSetLocked(ids)
	if len(order) == 0 {
		return
	}
	moved := make([]uuid.UUID, 0, len(order))
	for _, id := range order {
		idx := findCardOnBattlefield(g, id)
		if idx < 0 {
			continue
		}
		c := g.Battlefield.Cards[idx]
		// CR 702.26b's own last sentence: "A permanent that phases out
		// is removed from combat. (See rule 506.4.)"
		c.AttackingTarget = uuid.Nil
		c.BlockingTarget = uuid.Nil
		g.forgetCombatRecordLocked(id)
		// CR 702.26a: it phases in during the untap step of the player
		// who controlled it WHEN IT PHASED OUT, which is not
		// necessarily Controller by then — CR 702.26f lets a
		// control-changing continuous effect expire while it is out.
		c.PhasedOutBy = c.Controller
		c.PhasedOutIndirect = indirect[id]
		c.PhaseInLockedBy = opts.LockedBy
		c.TapOnPhaseIn = opts.TapOnPhaseIn
		// The layer cache is a battlefield-only cache and this object
		// is leaving the battlefield SLICE. Nilled here rather than in
		// the listener for the reason the transform arm of
		// layerVersionBump gives: EmitEvent dispatches synchronously
		// and an earlier listener would otherwise read the stale
		// characteristic off the card it was just told about.
		//
		// NOT clearEffectiveCacheLocked, which also calls
		// restorePrintedSelf: a Clone that phases out is still a copy
		// (CR 707.2 copiable values are unaffected by status).
		c.effective = nil
		g.Battlefield.Cards = append(g.Battlefield.Cards[:idx], g.Battlefield.Cards[idx+1:]...)
		g.PhasedOut.Cards = append(g.PhasedOut.Cards, c)
		moved = append(moved, id)
	}
	for _, id := range moved {
		g.EmitEvent(Event{
			Kind:   EventPhaseOut,
			Actor:  g.phasedOutControllerLocked(id),
			Source: source,
			CardID: id,
		})
	}
}

// phaseInLocked returns named phased-out permanents to the battlefield
// slice. Same batching contract as phaseOutLocked: move everything,
// then emit.
//
// Caller must hold g.mu in write mode.
func (g *Game) phaseInLocked(ids []uuid.UUID) []uuid.UUID {
	if g.PhasedOut == nil || g.Battlefield == nil || len(ids) == 0 {
		return nil
	}
	moved := make([]uuid.UUID, 0, len(ids))
	actors := make(map[uuid.UUID]uuid.UUID, len(ids))
	for _, id := range ids {
		idx := -1
		for i := range g.PhasedOut.Cards {
			if g.PhasedOut.Cards[i].InstanceID == id {
				idx = i
				break
			}
		}
		if idx < 0 {
			continue
		}
		c := g.PhasedOut.Cards[idx]
		c.PhasedOutBy = uuid.Nil
		c.PhasedOutIndirect = false
		c.PhaseInLockedBy = uuid.Nil
		if c.TapOnPhaseIn {
			// Oubliette. Written directly rather than through
			// untapPermanentLocked's twin: this is not the CR 701.21a
			// keyword action performed ON a permanent, it is the state
			// the permanent phases in WITH, so nothing watching for
			// "becomes tapped" should hear it.
			c.Tapped = true
			c.TapOnPhaseIn = false
		}
		c.effective = nil
		g.PhasedOut.Cards = append(g.PhasedOut.Cards[:idx], g.PhasedOut.Cards[idx+1:]...)
		g.Battlefield.Cards = append(g.Battlefield.Cards, c)
		actors[id] = c.Controller
		moved = append(moved, id)
	}
	for _, id := range moved {
		g.EmitEvent(Event{
			Kind:   EventPhaseIn,
			Actor:  actors[id],
			CardID: id,
		})
	}
	return moved
}

// phasedOutControllerLocked reads the controller of a card that has
// just been moved into g.PhasedOut. Caller must hold g.mu.
func (g *Game) phasedOutControllerLocked(id uuid.UUID) uuid.UUID {
	if g.PhasedOut == nil {
		return uuid.Nil
	}
	for i := range g.PhasedOut.Cards {
		if g.PhasedOut.Cards[i].InstanceID == id {
			return g.PhasedOut.Cards[i].Controller
		}
	}
	return uuid.Nil
}

// forgetCombatRecordLocked drops the combat DECLARATIONS the engine
// holds against one object, for CR 506.4's "removed from combat".
//
// Deliberately not forgetPerObjectTurnStateLocked, which would also
// drop LoyaltyActivatedThisTurn: CR 702.26d says an effect that checks
// a phased-in permanent's history must not treat the phasing as having
// changed anything, and CR 606.3's "only one loyalty ability, only
// once each turn" is exactly such a history. A planeswalker that
// phased out after using its ability has still used it.
//
// Caller must hold g.mu.
func (g *Game) forgetCombatRecordLocked(cardID uuid.UUID) {
	delete(g.announcedAttacks, cardID)
	if len(g.announcedAttacks) == 0 {
		g.announcedAttacks = nil
	}
	delete(g.announcedBlocks, cardID)
	if len(g.announcedBlocks) == 0 {
		g.announcedBlocks = nil
	}
	delete(g.blockedAttackers, cardID)
	if len(g.blockedAttackers) == 0 {
		g.blockedAttackers = nil
	}
	delete(g.firstStrikeStepParticipants, cardID)
	if len(g.firstStrikeStepParticipants) == 0 {
		g.firstStrikeStepParticipants = nil
	}
}

// performPhasingLocked is CR 502.1, the FIRST of the untap step's
// three turn-based actions:
//
//	"First, all phased-in permanents with phasing that the active
//	 player controls phase out, and all phased-out permanents that the
//	 active player controlled when they phased out phase in. This all
//	 happens simultaneously. This turn-based action doesn't use the
//	 stack."
//
// It runs at the top of performUntapStepLocked, ahead of the CR 302.6
// summoning-sickness clear and well ahead of untapStepSetLocked
// (CR 502.3) — a permanent phasing in has to be in the untap set and
// one phasing out has to be out of it.
//
// BOTH SETS ARE CHOSEN BEFORE EITHER IS APPLIED, which is what
// "simultaneously" costs: without it, an Aura phasing in onto a
// creature that is about to phase out would be dragged straight back
// out again in the same step.
//
// CR 702.26m — "If an effect causes a player to skip their untap step,
// the phasing event simply doesn't occur that turn" — falls out:
// performUntapStepLocked is not called for a skipped step, so neither
// is this.
//
// Caller must hold g.mu in write mode.
func (g *Game) performPhasingLocked(activePlayer uuid.UUID) {
	if activePlayer == uuid.Nil {
		return
	}
	var in []uuid.UUID
	if g.PhasedOut != nil {
		for i := range g.PhasedOut.Cards {
			c := &g.PhasedOut.Cards[i]
			if c.PhasedOutBy != activePlayer {
				continue
			}
			// CR 702.26g: an attachment that phased out indirectly
			// "won't phase in by itself, but instead phases in along
			// with the permanent it's attached to". Collected below
			// with its host rather than here.
			if c.PhasedOutIndirect {
				continue
			}
			// The "phases out until ~ leaves the battlefield" family
			// is not on the untap step's clock at all; its phase-in is
			// sweepPhaseInLocksLocked.
			if c.PhaseInLockedBy != uuid.Nil {
				continue
			}
			in = append(in, c.InstanceID)
		}
		in = g.withIndirectlyPhasedAttachmentsLocked(in)
	}
	var out []uuid.UUID
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller != activePlayer {
				continue
			}
			// Effective(), so a granted phasing (Shimmer's "each land
			// of the chosen type has phasing", an Aura that grants it)
			// works exactly like a printed one and a permanent that
			// has lost all abilities stops phasing.
			if !HasKeyword(c, KeywordPhasing) {
				continue
			}
			out = append(out, c.InstanceID)
		}
	}
	g.phaseInLocked(in)
	g.phaseOutLocked(uuid.Nil, out, phaseOutOptions{})
}

// withIndirectlyPhasedAttachmentsLocked grows a phase-in set by the
// attachments that phased out with each of its members (CR 702.26g),
// transitively. The mirror of phaseOutSetLocked, over g.PhasedOut.
//
// Only INDIRECT attachments are collected: one that phased out on its
// own (Clever Concealment naming an Aura, say) is on its own clock and
// is already in the set if it belongs there.
//
// Caller must hold g.mu.
func (g *Game) withIndirectlyPhasedAttachmentsLocked(ids []uuid.UUID) []uuid.UUID {
	if g.PhasedOut == nil || len(ids) == 0 {
		return ids
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	frontier := append([]uuid.UUID(nil), out...)
	for len(frontier) > 0 {
		var next []uuid.UUID
		for _, host := range frontier {
			for i := range g.PhasedOut.Cards {
				c := &g.PhasedOut.Cards[i]
				if !c.PhasedOutIndirect || seen[c.InstanceID] || !c.IsAttachedTo(host) {
					continue
				}
				seen[c.InstanceID] = true
				out = append(out, c.InstanceID)
				next = append(next, c.InstanceID)
			}
		}
		frontier = next
	}
	return out
}

// sweepPhaseInLocksLocked phases in every permanent whose "until ~
// leaves the battlefield" duration has ended — Oubliette destroyed,
// Out of Time's last time counter removed.
//
// Called from the top of stateBasedActionsLocked, before the layer
// recompute and before attachmentSBALocked, so that an Aura phasing in
// onto a host that died while it was away is swept into its owner's
// graveyard in the same settling (CR 702.26i, CR 704.5m).
//
// It is NOT a state-based action: it is a CR 611.2b "for as long as"
// duration ending, and the SBA pass is simply the one place in the
// engine that runs after every action with the board settled. The
// distinction matters for CR 704.3, which is about applying SBAs to
// phased-out permanents — that stays true by construction, since a
// phased-out permanent is not in the battlefield slice the SBA pass
// walks.
//
// Returns whether anything phased in. Caller must hold g.mu.
func (g *Game) sweepPhaseInLocksLocked() bool {
	if g.PhasedOut == nil || len(g.PhasedOut.Cards) == 0 {
		return false
	}
	var ready []uuid.UUID
	for i := range g.PhasedOut.Cards {
		c := &g.PhasedOut.Cards[i]
		if c.PhaseInLockedBy == uuid.Nil {
			continue
		}
		if findCardOnBattlefield(g, c.PhaseInLockedBy) >= 0 {
			continue
		}
		ready = append(ready, c.InstanceID)
	}
	if len(ready) == 0 {
		return false
	}
	// The host's own attachments came out with it and come back with
	// it (CR 702.26g), whether or not they carry the lock themselves.
	return len(g.phaseInLocked(g.withIndirectlyPhasedAttachmentsLocked(ready))) > 0
}

// PhasedOutCardsForEffect is the read surface for the holding slice —
// the "rules and effects that specifically mention phased-out
// permanents" half of CR 702.26b. By value, like
// BattlefieldCardsForEffect.
//
// Caller must hold g.mu.
func (g *Game) PhasedOutCardsForEffect() []Card {
	if g.PhasedOut == nil {
		return nil
	}
	out := make([]Card, len(g.PhasedOut.Cards))
	copy(out, g.PhasedOut.Cards)
	return out
}

// IsPhasedOutForEffect reports whether the named object is currently
// phased out. Caller must hold g.mu.
func (g *Game) IsPhasedOutForEffect(cardID uuid.UUID) bool {
	if g.PhasedOut == nil {
		return false
	}
	for i := range g.PhasedOut.Cards {
		if g.PhasedOut.Cards[i].InstanceID == cardID {
			return true
		}
	}
	return false
}

// PhaseOutForEffect phases out the named permanents and everything
// attached to them (CR 702.26g), as a one-shot instruction — Teferi's
// Protection's "all permanents you control phase out", Vodalian
// Illusionist's "target creature phases out".
//
// A named permanent that is not on the battlefield is skipped, which
// is CR 608.2b's per-slot existence re-check for free. Caller must
// hold g.mu.
func (g *Game) PhaseOutForEffect(source uuid.UUID, ids ...uuid.UUID) error {
	g.phaseOutLocked(source, ids, phaseOutOptions{})
	return nil
}

// PhaseOutUntilLeavesForEffect is the "phases out until ~ leaves the
// battlefield" form (Oubliette, Out of Time). `until` is the object
// whose presence is the duration; `tapOnPhaseIn` is Oubliette's "Tap
// that creature as it phases in this way".
//
// Caller must hold g.mu.
func (g *Game) PhaseOutUntilLeavesForEffect(source, until uuid.UUID, tapOnPhaseIn bool, ids ...uuid.UUID) error {
	if until == uuid.Nil {
		return ErrInvalidParam
	}
	g.phaseOutLocked(source, ids, phaseOutOptions{LockedBy: until, TapOnPhaseIn: tapOnPhaseIn})
	return nil
}
