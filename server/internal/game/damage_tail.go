package game

import "github.com/google/uuid"

// damage_tail.go is the single answer to "what still has to happen once
// the CR 614 replacement pipeline has settled a damage event?", and it
// exists for the same reason permanent_damage.go does: the answer used
// to be written out once per entry point, and one of those copies was
// wrong.
//
// Issue #694: six call sites build a RepEventDamage, run
// applyReplacementsLocked, and then finish the job inline. When two or
// more replacements apply, the pipeline returns errReplacementPending
// and every one of those six returns early — the damage lands later,
// from the CR 616 resume in applyResolvedReplacementEventLocked. That
// resume was a seventh copy of the tail, cloned from the oldest and
// simplest entry point (the manual MarkDamage sandbox verb): it marked
// DamageMarked on a battlefield card and returned ErrCardNotFound for
// anything else. So ordering two damage replacements silently dropped
// the player's life loss, the CR 120.3 planeswalker/battle split,
// deathtouch, lifelink and commander damage — a Lightning Bolt aimed at
// an opponent under Torbran + Angrath's Marauders changed nobody's life
// total, and the action came back as an error with the prompt already
// dequeued.
//
// The fix is the one #529 used for zone moves: the event carries its
// own tail, and there is exactly one implementation of it. Each entry
// point fills in a damageTail before it calls the pipeline, then calls
// applyResolvedDamageLocked with the settled event — and so does the
// resume. A paused damage event now ends where an unpaused one would.
//
// THE TAIL IS A SNAPSHOT, ON PURPOSE. Deathtouch, lifelink, the actor
// and "is this source a commander" are read off the source when the
// event is CREATED, not when the damage lands. Before #694 the two
// non-frame combat paths read them afterwards, which is fine when
// nothing pauses and wrong when something does: a CR 616 prompt in the
// middle of the combat damage step is answered after the blockers'
// damage has already been applied and swept by SBAs, so the attacker
// whose lifelink and deathtouch are being asked about may be in a
// graveyard by then. The two DamageAssignmentFrame paths already
// snapshotted for exactly this reason (see DamageAssignmentFrame); the
// tail generalises it to all of them.
//
// WHAT THE TAIL DELIBERATELY DOES NOT DO. It never runs state-based
// actions. The combat paths must not — all combat damage is dealt
// simultaneously and the sweep belongs after the whole step (CR 510.2)
// — and the effect paths get theirs from the resolution bookend around
// resolveTopOfStackLocked. The two callers that owe a sweep
// (markDamageWithKind, and the CR 616 resume, which is an action
// boundary like every other Resolve* handler) run it themselves.

// damageTailKind selects which shape of tail a settled damage event
// gets. It is about the TARGET and the entry point's contract, not
// about combat: whether the damage is combat damage is a separate
// flag, because both a player and a permanent can take either kind.
type damageTailKind uint8

const (
	// damageTailManualMark is the MarkDamage / MarkCombatDamage
	// sandbox verb: a signed delta written straight onto
	// Card.DamageMarked and clamped at zero, with no CR 120.3 split
	// and no combat riders. It is the sandbox's "fix the board by
	// hand" affordance, not a rules path, and it is the shape the
	// pre-#694 resume applied to every damage event.
	damageTailManualMark damageTailKind = iota

	// damageTailPermanent is damage to a battlefield permanent,
	// applied through the CR 120.3 split in
	// applyDamageToPermanentLocked.
	damageTailPermanent

	// damageTailPlayer is damage to a player: life loss, plus the
	// CR 903.10a commander-damage tally when the source is a
	// commander dealing combat damage.
	damageTailPlayer
)

// damageTail is the post-replacement half of a damage event — the part
// that runs once the CR 614 pipeline has settled the amount. Unexported
// engine plumbing: the catalog never sets or reads it, it just sees the
// ReplacementEvent fields.
//
// Everything on it except kind and combat is a snapshot of the damage
// SOURCE taken at event-creation time. See the file comment for why.
type damageTail struct {
	kind damageTailKind

	// combat marks this as combat damage (CR 510). Sets Combat on the
	// emitted EventDealDamage so "deals combat damage" triggers see
	// it, and gates the commander-damage tally.
	combat bool

	// actor is the player the emitted event is attributed to — the
	// controller of the damage source. uuid.Nil leaves Actor unset,
	// which is what the non-combat effect paths have always done.
	actor uuid.UUID

	// deathtouch marks the target as destroyed-at-the-next-sweep
	// (Card.MarkedLethalByDeathtouch, read by the CR 704.5h SBA).
	// CR 702.2b is the rule: "a creature with toughness greater than
	// 0 that's been dealt damage by a source with deathtouch since
	// the last time state-based actions were checked is destroyed" —
	// no mention of combat, which is why every entry point sets it
	// from the source since #711, not just the combat ones.
	// (CR 702.2c is the separate, combat-only half about what counts
	// as lethal when ASSIGNING combat damage.)
	deathtouch bool

	// lifelinkTo is the player CR 702.15b credits with life equal to
	// the damage dealt — the source's controller. uuid.Nil means the
	// source has no lifelink, or is not a permanent at all (a spell,
	// an emblem, an absent source) and so has no keywords to read.
	// Set on every path since #711, for the same reason as
	// deathtouch above.
	lifelinkTo uuid.UUID

	// commanderSource is the instance ID the CR 903.10a 21-damage
	// tally is keyed under, set only when the source is a commander.
	// uuid.Nil means "not a commander", so no tally. Keyed by instance
	// ID rather than looked up at landing time because a partner pair
	// runs two independent clocks (#77) and because the commander may
	// have died to blocker damage before a paused event resumes.
	commanderSource uuid.UUID
}

// combatDamageTailLocked snapshots a live battlefield source into a
// combat damage tail. Used by the two combat paths that have a real
// source card in front of them; the DamageAssignmentFrame paths build
// their tail from the frame instead, because the frame already IS this
// snapshot, taken before the prompt was queued.
//
// Caller must hold g.mu.
func (g *Game) combatDamageTailLocked(kind damageTailKind, sourceID uuid.UUID) *damageTail {
	t := &damageTail{kind: kind, combat: true}
	src := findBattlefieldCard(g, sourceID)
	if src == nil {
		return t
	}
	t.actor = src.Controller
	t.deathtouch = HasKeyword(src, "deathtouch")
	if HasKeyword(src, "lifelink") {
		t.lifelinkTo = src.Controller
	}
	if src.IsCommander {
		t.commanderSource = src.InstanceID
	}
	return t
}

// effectDamageTailLocked snapshots a live battlefield source into a
// NON-COMBAT damage tail: the fight primitive, a pinger's activated
// ability, a "target creature you control deals damage equal to its
// power" spell. The sibling of combatDamageTailLocked, and it reads
// the same two keywords off the same place for the same reason.
//
// #711: CR 702.15b and CR 702.2b are statements about the SOURCE, not
// about the combat damage step — "damage dealt by a source with
// lifelink causes that source's controller ... to gain that much
// life", "a creature ... that's been dealt damage by a source with
// deathtouch ... is destroyed as a state-based action". Neither says
// combat. Both entry points used to pass a bare tail, so a Wurmcoil
// Engine's fight gained its controller nothing and a Prodigal
// Pyromancer wearing a Basilisk Collar pinged for one plain damage.
// The card files had been claiming otherwise for sprints (Bite Down,
// Soul's Fire and Chandra's Ignition all say the damage carries the
// creature's deathtouch and lifelink "as printed"); now they are right.
//
// A source that is not a battlefield permanent — a spell, an emblem, a
// token that has already left, uuid.Nil — has no characteristics to
// read and carries neither keyword. That is what "a source with
// lifelink" means, and it is why Lightning Bolt still just deals 3.
//
// Unlike combatDamageTailLocked this sets no actor and no commander
// source. The non-combat paths have never attributed their
// EventDealDamage to a player, and the CR 903.10a 21-damage tally is
// combat damage only; neither is what #711 is about.
//
// Keywords come from the layer engine's resolved characteristics, as
// they do for combat. Callers that grant or pump in the same breath
// (the fight primitive, every activated ability, stack resolution)
// have already recomputed by the time they get here.
//
// Caller must hold g.mu.
func (g *Game) effectDamageTailLocked(kind damageTailKind, sourceID uuid.UUID) *damageTail {
	t := &damageTail{kind: kind}
	if sourceID == uuid.Nil {
		return t
	}
	src := findBattlefieldCard(g, sourceID)
	if src == nil {
		return t
	}
	t.deathtouch = HasKeyword(src, "deathtouch")
	if HasKeyword(src, "lifelink") {
		t.lifelinkTo = src.Controller
	}
	return t
}

// damageTailFromFrame builds a combat damage tail from a queued
// CR 510.1c damage-assignment prompt's frame. The frame was filled in
// when the prompt was queued, which is the whole point: by the time it
// is answered the attacker may have died to blocker damage dealt in the
// same substep, so a battlefield lookup would lose its deathtouch,
// lifelink and commander status.
func damageTailFromFrame(kind damageTailKind, frame *DamageAssignmentFrame) *damageTail {
	t := &damageTail{
		kind:       kind,
		combat:     true,
		actor:      frame.SourceController,
		deathtouch: frame.HasDeathtouch,
	}
	if frame.SourceLifelink {
		t.lifelinkTo = frame.SourceController
	}
	if frame.SourceIsCommander {
		t.commanderSource = frame.AttackerID
	}
	return t
}

// applyResolvedDamageLocked performs the underlying mutation for a
// damage ReplacementEvent whose replacement pipeline has settled. This
// is the ONLY place damage lands: every entry point calls it on the
// unpaused path and the CR 616 resume calls it on the paused one, so
// the two cannot drift.
//
// The caller is responsible for state-based actions — see the file
// comment.
//
// Returns ErrCardNotFound / ErrPlayerNotFound when the target is no
// longer there to be damaged. The inline callers that have always
// surfaced that keep surfacing it; the resume treats it as "the target
// left, the damage simply does not happen" rather than failing an
// action whose prompt is already dequeued.
//
// Caller must hold g.mu.
func (g *Game) applyResolvedDamageLocked(ev *ReplacementEvent) error {
	if ev == nil {
		return nil
	}
	// A damage event with no tail can only come from outside this
	// file's six entry points. None exists today; the manual mark is
	// the conservative default because it is what the pre-#694 resume
	// did for every event that reached it.
	t := ev.damageTail
	if t == nil {
		t = &damageTail{kind: damageTailManualMark}
	}
	switch t.kind {
	case damageTailManualMark:
		return g.applyManualDamageMarkLocked(ev)
	case damageTailPlayer:
		return g.applyResolvedDamageToPlayerLocked(ev, t)
	case damageTailPermanent:
		return g.applyResolvedDamageToPermanentLocked(ev, t)
	}
	return nil
}

// applyManualDamageMarkLocked is the sandbox MarkDamage verb's tail: a
// signed delta on DamageMarked, clamped at zero, with no CR 120.3
// split and no combat riders. Deliberately unchanged from what it has
// always been — it is the "a player is fixing the board by hand" path,
// and a negative delta (undo a mark) is a legitimate use of it.
//
// Caller must hold g.mu.
func (g *Game) applyManualDamageMarkLocked(ev *ReplacementEvent) error {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID != ev.DamageTarget {
			continue
		}
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
		return nil
	}
	return ErrCardNotFound
}

// applyResolvedDamageToPlayerLocked lands settled damage on a player:
// life loss, the CR 903.10a commander tally, the EventDealDamage the
// "deals damage to a player" triggers watch, and CR 702.15 lifelink.
//
// Caller must hold g.mu.
func (g *Game) applyResolvedDamageToPlayerLocked(ev *ReplacementEvent, t *damageTail) error {
	if ev.DamageAmount <= 0 {
		return nil
	}
	p := g.playerByIDLocked(ev.DamageTarget)
	if p == nil {
		return ErrPlayerNotFound
	}
	// The event is emitted before the life change on the non-combat
	// path so the log reads "damage dealt → life changed"; the combat
	// path has always changed life first. Both orders are kept as they
	// were, because a listener on EventDealDamage can read life totals
	// and the two paths' existing tests pin what it sees.
	if t.combat {
		p.ChangeLife(-ev.DamageAmount)
		// CR 903.10a: combat damage from a commander accrues toward
		// the 21-damage loss SBA.
		if t.commanderSource != uuid.Nil {
			p.RecordCommanderDamage(t.commanderSource, ev.DamageAmount)
		}
		g.emitDealDamageLocked(ev, t)
	} else {
		g.emitDealDamageLocked(ev, t)
		p.ChangeLife(-ev.DamageAmount)
	}
	g.creditLifelinkLocked(t, ev.DamageSource, ev.DamageAmount)
	return nil
}

// applyResolvedDamageToPermanentLocked lands settled damage on a
// battlefield permanent through the CR 120.3 split (marked on a
// creature, loyalty off a planeswalker, defense off a battle), then
// emits the event and credits lifelink.
//
// Caller must hold g.mu.
func (g *Game) applyResolvedDamageToPermanentLocked(ev *ReplacementEvent, t *damageTail) error {
	if ev.DamageAmount <= 0 {
		// Fully prevented. The permanent is untouched and no
		// EventDealDamage fires, which is what "prevented" means — a
		// "whenever ~ is dealt damage" trigger must not see it.
		return nil
	}
	if !g.applyDamageToPermanentLocked(ev.DamageTarget, ev.DamageAmount, t.deathtouch) {
		return ErrCardNotFound
	}
	g.emitDealDamageLocked(ev, t)
	g.creditLifelinkLocked(t, ev.DamageSource, ev.DamageAmount)
	return nil
}

// emitDealDamageLocked emits the one EventDealDamage a settled damage
// event produces, with the Actor and Combat fields the tail carries so
// "deals combat damage" triggers key off the same shape whether or not
// the event paused.
//
// Caller must hold g.mu.
func (g *Game) emitDealDamageLocked(ev *ReplacementEvent, t *damageTail) {
	g.EmitEvent(Event{
		Kind:   EventDealDamage,
		Actor:  t.actor,
		Source: ev.DamageSource,
		Target: ev.DamageTarget,
		Amount: ev.DamageAmount,
		Combat: t.combat,
	})
}

// creditLifelinkLocked credits the tail's lifelink beneficiary with
// life equal to the damage dealt (CR 702.15). No-op when the source
// had no lifelink, or when this entry point has never applied it.
//
// Runs through ChangeLife rather than the replacement pipeline: life
// gain from lifelink is not itself treated as a replaceable event at
// this scope.
//
// Caller must hold g.mu.
func (g *Game) creditLifelinkLocked(t *damageTail, sourceID uuid.UUID, amount int) {
	if amount <= 0 || t.lifelinkTo == uuid.Nil {
		return
	}
	p := g.playerByIDLocked(t.lifelinkTo)
	if p == nil {
		return
	}
	p.ChangeLife(amount)
	g.EmitEvent(Event{
		Kind:   EventChangeLife,
		Target: t.lifelinkTo,
		Source: sourceID,
		Amount: amount,
	})
}
