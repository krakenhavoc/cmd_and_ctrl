package game

import (
	"errors"
	"slices"

	"github.com/google/uuid"
)

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
//
// THE SECOND HALF OF THE TAIL: WHAT THE EVENT OWES ITS CALLER (#807).
// Everything above is what the ENGINE still owes a settled damage
// event. `damageTail.then` is the other half, and it is the same idea
// life_tail.go's lifeTail carries: the rest of the EFFECT that asked
// for the damage, run with the amount that actually landed.
//
// "This creature deals 1 damage to each opponent. You gain life equal
// to the damage dealt this way" (Creeping Bloodsucker) used to read
// each opponent's life total back on the line after dealing the
// damage. A damage event runs the CR 614 window, so it can pause on a
// CR 616 ordering prompt when two DIFFERENT damage replacements apply
// to one target — and then the read-back happens before the prompt is
// answered, that opponent counts as having taken nothing, and the gain
// is short by everything they took. Exactly #793's bug with
// DealDamageToPlayerForEffect in place of ChangePlayerLifeForEffect.
//
// So damage gets the same public continuation life got:
// DealDamageToPlayerThenForEffect / DealDamageToCreatureThenForEffect
// for one target, DealDamageEachThenForEffect for the batch, and
// runDamageTailLocked is the one place `then` runs. Card authors learn
// `...ThenForEffect` once and it means the same thing on both sides.

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

	// combatStep is the combat damage step the event was created in —
	// CombatStepFirstStrike, CombatStepRegular, or "" when no
	// first-strike pass ran (#187, ADR 0053 Decision 1). Copied onto
	// Event.CombatStep by emitDealDamageLocked. Riding the tail rather
	// than being read off the game when the damage lands is what keeps
	// it right across a pause: a CR 616 ordering prompt resumes with
	// the tag it was created with, and there is no Game field that a
	// finished pass could leave stale.
	combatStep string

	// actor is the player the emitted event is attributed to — the
	// controller of the damage source. uuid.Nil leaves Actor unset,
	// which is what the non-combat effect paths have always done.
	actor uuid.UUID

	// sourceLKI is the damage source's characteristics at
	// event-creation time, copied onto ReplacementEvent.SourceLKI by
	// damageThroughReplacementsLocked. CR 702.16e reads it; see the
	// field's doc on ReplacementEvent for why a DamageSource lookup
	// cannot answer the same question. #662.
	sourceLKI *Characteristic

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

	// then is the CALLER's half of the tail (#807): the rest of the
	// effect that asked for the damage, run with the amount that
	// ACTUALLY landed once the CR 614 window has settled it. The exact
	// sibling of lifeTail.then, and set by the same kind of entry
	// point — the catalog never builds a damageTail, it hands a `then`
	// to one of the ...ThenForEffect forms.
	//
	// `dealt` is the post-replacement amount, always non-negative on
	// the two rules paths, and 0 when nothing landed: fully prevented,
	// replaced away under CR 614.10, or the target gone between the
	// prompt and the answer. The continuation is told either way,
	// because a batch adding up "the damage dealt this way" must not
	// stall on the leg that dealt nothing.
	//
	// It takes the live *Game rather than capturing one, on the same
	// undo-safety contract confirmFrame, StackItem.Effect and
	// lifeTail.then follow — an undo restores this game's fields in
	// place, so a *Game argument is always the right game. It runs
	// with g.mu held, so it may start the next damage event or queue
	// the next prompt itself.
	//
	// Nil on every engine-internal path: combat damage, the manual
	// sandbox mark and every fire-and-forget catalog deal owe their
	// caller nothing.
	then func(g *Game, dealt int) error
}

// combatDamageTailLocked snapshots a live battlefield source into a
// combat damage tail. Used by the two combat paths that have a real
// source card in front of them; the DamageAssignmentFrame paths build
// their tail from the frame instead, because the frame already IS this
// snapshot, taken before the prompt was queued.
//
// step is the pass's Event.CombatStep value ("" when no first-strike
// pass ran). It is an argument, not something read off the game, so no
// value can outlive the pass that set it.
//
// Caller must hold g.mu.
func (g *Game) combatDamageTailLocked(kind damageTailKind, sourceID uuid.UUID, step string) *damageTail {
	t := &damageTail{kind: kind, combat: true, combatStep: step}
	t.sourceLKI = g.damageSourceLKILocked(sourceID)
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

// effectDamageTailLocked snapshots a permanent source, live or as it
// last existed, into a NON-COMBAT damage tail: the fight primitive, a pinger's activated
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
// A source that is not and never was a battlefield permanent — a
// spell, an emblem, uuid.Nil — has no permanent characteristics to
// read and carries neither keyword. That is what "a source with
// lifelink" means, and it is why Lightning Bolt still just deals 3.
// A permanent that HAS left is the #1396 paragraph below.
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
// #1396, CR 608.2h: a source that has LEFT the battlefield deals its
// damage with the keywords it had as it last existed there. "When this
// creature dies, it deals 2 damage to any target" from a creature
// wearing a Basilisk Collar is deathtouch damage, and a lifelinker
// killed in response to Warstorm Surge still gains its controller the
// life. Those keywords are read off Game.lastKnownPermanents, the
// per-object record #1379 keeps; see departedDamageSourceLocked for
// which object a source names. A spell on the stack, an emblem or a
// card that was never on the battlefield has no record and carries
// neither keyword, as before.
//
// obj names the source OBJECT (CR 400.7) when the caller knows it — a
// trigger about the creature that entered, via its ObjectSnapshot.
// With obj set, a live permanent of a different epoch is a new object
// and is never read: the record for obj is. Nil for every caller that
// has only an instance ID.
//
// Caller must hold g.mu.
func (g *Game) effectDamageTailLocked(kind damageTailKind, sourceID uuid.UUID, obj *ObjectRef) *damageTail {
	t := &damageTail{kind: kind}
	if sourceID == uuid.Nil {
		return t
	}
	// #662: the LKI starts from WHATEVER ZONE holds the source. A
	// Lightning Bolt is on the stack while it deals its damage and has
	// never been on the battlefield, and its colour is exactly what
	// CR 702.16e asks about. A departed permanent overrides it below
	// (#1417).
	t.sourceLKI = g.damageSourceLKILocked(sourceID)
	if src := findBattlefieldCard(g, sourceID); src != nil && (obj == nil || src.ObjectEpoch == obj.Epoch) {
		t.deathtouch = HasKeyword(src, "deathtouch")
		if HasKeyword(src, "lifelink") {
			t.lifelinkTo = src.Controller
		}
		return t
	}
	rec, ok := g.departedDamageSourceLocked(sourceID, obj)
	if !ok {
		return t
	}
	// #1417, CR 608.2h: the characteristics CR 702.16e protection and
	// every "damage from a <colour> source" matcher read are the
	// departed object's too, not the card's in the zone it went to. A
	// creature painted red that dies deals red damage; a printed-red
	// creature made colourless deals colourless damage. The same
	// record the keywords come from, so the two can never disagree
	// about which object dealt the damage — and a card that has moved
	// on (bounced and then cast, so a SPELL now) has no matching
	// record and keeps the current-zone read above.
	t.sourceLKI = lastKnownSourceCharacteristics(rec)
	// The record's Characteristic is Effective() as it last stood, so
	// its ability tokens are exactly what HasKeyword read off the live
	// card.
	t.deathtouch = slices.Contains(rec.Characteristic.Abilities, "deathtouch")
	if slices.Contains(rec.Characteristic.Abilities, "lifelink") {
		t.lifelinkTo = rec.Controller
	}
	return t
}

// departedDamageSourceLocked is the last-known record of the permanent
// a damage source names once that permanent has left the battlefield
// (#1396, CR 608.2h), or false when there is none to read.
//
// With obj, the answer is obj's own record and nothing else — the
// caller has said which object it means.
//
// With only an instance ID, the source is the LAST battlefield object
// the card was, and only while the card has not moved since it left:
// its current epoch is exactly one past the record's (CR 400.7 bumps
// the epoch once per zone change, in MoveCard). That is the dies
// trigger's case — the creature is in the graveyard it went to — and it
// excludes every card that has moved on: one that was bounced and then
// cast (on the stack now, a spell with no battlefield past), one that
// died and was then exiled from the graveyard, one that came back and
// left again (the newer record is the one that matches). A token that
// has ceased to exist (CR 704.5d) is in no zone at all and cannot come
// back (CR 111.8), so its last record is unambiguous.
//
// A card back on the battlefield never reaches here: the live read in
// effectDamageTailLocked answers first, with the new object.
//
// Caller must hold g.mu.
func (g *Game) departedDamageSourceLocked(sourceID uuid.UUID, obj *ObjectRef) (PermanentInfo, bool) {
	if obj != nil {
		return g.lastKnownPermanentLocked(*obj)
	}
	recs := g.lastKnownPermanents[sourceID]
	if len(recs) == 0 {
		return PermanentInfo{}, false
	}
	last := recs[len(recs)-1]
	if now := g.cardObjectEpochLocked(sourceID); now != -1 && now != last.Epoch+1 {
		return PermanentInfo{}, false
	}
	return last, true
}

// damageSourceLKILocked snapshots the damage source's
// characteristics from WHATEVER ZONE holds it right now — CR 608.2h's
// last-known information, taken at event-creation time so a CR 616
// pause, an SBA sweep or a resolution that puts the source in a
// graveyard cannot change the answer.
//
// Deliberately not findBattlefieldCard: a SPELL deals its damage from
// the stack, and "protection from red" is mostly about spells.
//
// Returns nil for uuid.Nil and for a source in no zone at all, which
// the matcher reads as "unknown source" and lets the damage through.
//
// Caller must hold g.mu.
func (g *Game) damageSourceLKILocked(sourceID uuid.UUID) *Characteristic {
	if sourceID == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(sourceID)
	if !ok {
		return nil
	}
	return SourceCharacteristics(&c)
}

// damageTailFromFrame builds a combat damage tail from a queued
// CR 510.1c damage-assignment prompt's frame. The frame was filled in
// when the prompt was queued, which is the whole point: by the time it
// is answered the attacker may have died to blocker damage dealt in the
// same substep, so a battlefield lookup would lose its deathtouch,
// lifelink and commander status.
//
// The combat step comes from frame.CombatStep, NOT frame.FirstStrike:
// a regular-pass prompt has FirstStrike false whether or not a
// first-strike pass ran, and a combat with no first strike anywhere
// must stay untagged (ADR 0053 Decision 1). A frame restored from a
// snapshot written before the field existed has "" and resumes
// untagged — the cue is lost, the board is right.
func damageTailFromFrame(kind damageTailKind, frame *DamageAssignmentFrame) *damageTail {
	t := &damageTail{
		kind:       kind,
		combat:     true,
		combatStep: frame.CombatStep,
		actor:      frame.SourceController,
		deathtouch: frame.HasDeathtouch,
		sourceLKI:  frame.SourceLKI,
	}
	if frame.SourceLifelink {
		t.lifelinkTo = frame.SourceController
	}
	if frame.SourceIsCommander {
		t.commanderSource = frame.AttackerID
	}
	return t
}

// damageThroughReplacementsLocked runs the CR 614 window on a damage
// event and lands it through the one tail. The shared body behind all
// six entry points — the two effect-facing deals, the three combat
// paths and the manual sandbox mark — and the exact sibling of
// changeLifeThroughReplacementsLocked.
//
// Reports paused=true when a CR 616 ordering prompt was queued (two
// DIFFERENT damage replacements applied to one event and the affected
// player has to order them). The event is not lost: it lands from
// applyResolvedReplacementEventLocked when the prompt is answered. A
// caller that owes its own caller an amount has nothing to report
// until then.
//
// Before #807 each of the six wrote this dance out itself, which is
// how a new terminal outcome (the caller's continuation) would have
// had to be added in six places and remembered in six more.
//
// Caller must hold g.mu.
func (g *Game) damageThroughReplacementsLocked(ev *ReplacementEvent) (paused bool, err error) {
	if ev == nil {
		return false, nil
	}
	// #662: the ONE place the source's last-known information moves
	// from the tail (where every entry point already snapshots it,
	// alongside deathtouch and lifelink) onto the event the CR 614
	// pipeline sees. CR 702.16e's built-in reads it there.
	if ev.SourceLKI == nil && ev.damageTail != nil {
		ev.SourceLKI = ev.damageTail.sourceLKI
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		return true, nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return false, err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		// CR 614.10 with a null replacement — a Fog, a prevention
		// shield that ate the whole event. No mutation and no
		// EventDealDamage, because a "whenever ~ is dealt damage"
		// trigger must not see damage that was prevented.
		//
		// The continuation still runs, with zero: "you gain life equal
		// to the damage dealt this way" gains nothing when the damage
		// was prevented, and a batch adding up several opponents would
		// otherwise wait forever for this one.
		return false, g.runDamageTailLocked(ev, 0)
	}
	return false, g.applyResolvedDamageLocked(out)
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
	var dealt int
	var err error
	switch t.kind {
	case damageTailManualMark:
		dealt, err = g.applyManualDamageMarkLocked(ev)
	case damageTailPlayer:
		dealt, err = g.applyResolvedDamageToPlayerLocked(ev, t)
	case damageTailPermanent:
		dealt, err = g.applyResolvedDamageToPermanentLocked(ev, t)
	}
	if err != nil {
		// The target is no longer there. That is still a TERMINAL
		// outcome of the event, so the continuation is told — with
		// zero, because nothing landed — and the original error is
		// what the caller (or the resume, which turns it into a
		// logged drop) sees. A continuation that fails on top of a
		// vanished target has nowhere better to go than the log.
		if tailErr := g.runDamageTailLocked(ev, 0); tailErr != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				ErrorMsg: "damage continuation failed: " + tailErr.Error(),
			})
		}
		return err
	}
	return g.runDamageTailLocked(ev, dealt)
}

// runDamageTailLocked runs a settled damage event's continuation
// exactly once, with the amount that actually landed. The continuation
// is cleared before it runs, so one that re-enters the pipeline on the
// same event value cannot run itself twice.
//
// Every TERMINAL outcome of a damage event goes through here — landed,
// fully prevented, replaced away, target gone — because a caller adding
// up "the damage dealt this way" has to be told even when the answer is
// zero, or it waits forever. A pause is not a terminal outcome: the
// resume reaches this function later, through applyResolvedDamageLocked.
//
// The continuation is cleared on the EVENT's own damageTail pointer, a
// copy of which every undo snapshot has of its own
// (cloneReplacementResume) — so undoing the answer to a CR 616 prompt
// and answering it again runs the continuation again, exactly once.
//
// Caller must hold g.mu.
func (g *Game) runDamageTailLocked(ev *ReplacementEvent, dealt int) error {
	if ev == nil || ev.damageTail == nil || ev.damageTail.then == nil {
		return nil
	}
	then := ev.damageTail.then
	ev.damageTail.then = nil
	return then(g, dealt)
}

// applyManualDamageMarkLocked is the sandbox MarkDamage verb's tail: a
// signed delta on DamageMarked, clamped at zero, with no CR 120.3
// split and no combat riders. Deliberately unchanged from what it has
// always been — it is the "a player is fixing the board by hand" path,
// and a negative delta (undo a mark) is a legitimate use of it.
//
// Reports the damage DEALT, which is the delta only when it is
// positive: taking a mark back off a creature is not dealing it
// negative damage. The verb carries no continuation today, so this is
// the answer to a question nobody asks; giving it the honest one costs
// nothing and keeps the three tails' contract identical.
//
// Caller must hold g.mu.
func (g *Game) applyManualDamageMarkLocked(ev *ReplacementEvent) (int, error) {
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
			return ev.DamageAmount, nil
		}
		return 0, nil
	}
	return 0, ErrCardNotFound
}

// applyResolvedDamageToPlayerLocked lands settled damage on a player:
// life loss, the CR 903.10a commander tally, the EventDealDamage the
// "deals damage to a player" triggers watch, and CR 702.15 lifelink.
//
// The life loss is the one part a player whose LIFE TOTAL CAN'T CHANGE
// (CR 119.8, #1200, ADR 0085) skips — the damage is still dealt, which
// is what CR 120.3a's "causes that player to lose that much life"
// means once something says the life can't be lost.
//
// Reports the damage actually dealt, for the caller's continuation.
//
// Caller must hold g.mu.
func (g *Game) applyResolvedDamageToPlayerLocked(ev *ReplacementEvent, t *damageTail) (int, error) {
	if ev.DamageAmount <= 0 {
		// Fully prevented. Nothing landed and no EventDealDamage
		// fires, so the continuation is told zero.
		return 0, nil
	}
	p := g.playerByIDLocked(ev.DamageTarget)
	if p == nil {
		return 0, ErrPlayerNotFound
	}
	if p.Eliminated {
		// #808, CR 800.4a: an eliminated seat stays in g.Seats, so the
		// nil check alone let a player who conceded between a CR 616
		// prompt and its answer take the damage.
		return 0, ErrPlayerEliminated
	}
	// The event is emitted before the life change on the non-combat
	// path so the log reads "damage dealt → life changed"; the combat
	// path has always changed life first. #1117's layer invalidation
	// rides the WRITE on both, never the event, precisely because the
	// two orders differ — see invalidateLayersForLifeChangeLocked. Both orders are kept as they
	// were, because a listener on EventDealDamage can read life totals
	// and the two paths' existing tests pin what it sees.
	//
	// #1200, CR 119.8 + CR 120.3a (ADR 0085 Decision 5): the damage is
	// still DEALT to a player whose life total can't change — the
	// event fires, the CR 903.10a tally accrues, lifelink credits, a
	// "whenever ~ is dealt damage" trigger sees it — and only the life
	// loss CR 120.3a would have caused does not happen. Platinum
	// Emperion's printed ruling, and the reason this is a branch here
	// rather than an arm of the life built-in: damage does not fire
	// the CR 614 life window at all, so the rule has two consumers.
	loseLife := !g.playerLifeTotalCantChangeLocked(p)
	if t.combat {
		if loseLife {
			p.ChangeLife(-ev.DamageAmount)
			g.invalidateLayersForLifeChangeLocked()
		}
		// CR 903.10a: combat damage from a commander accrues toward
		// the 21-damage loss SBA.
		if t.commanderSource != uuid.Nil {
			p.RecordCommanderDamage(t.commanderSource, ev.DamageAmount)
		}
		g.emitDealDamageLocked(ev, t)
	} else {
		g.emitDealDamageLocked(ev, t)
		if loseLife {
			p.ChangeLife(-ev.DamageAmount)
			g.invalidateLayersForLifeChangeLocked()
		}
	}
	g.creditLifelinkLocked(t, ev.DamageSource, ev.DamageAmount)
	return ev.DamageAmount, nil
}

// applyResolvedDamageToPermanentLocked lands settled damage on a
// battlefield permanent through the CR 120.3 split (marked on a
// creature, loyalty off a planeswalker, defense off a battle), then
// emits the event and credits lifelink.
//
// Reports the damage actually dealt, for the caller's continuation.
//
// Caller must hold g.mu.
func (g *Game) applyResolvedDamageToPermanentLocked(ev *ReplacementEvent, t *damageTail) (int, error) {
	if ev.DamageAmount <= 0 {
		// Fully prevented. The permanent is untouched and no
		// EventDealDamage fires, which is what "prevented" means — a
		// "whenever ~ is dealt damage" trigger must not see it.
		return 0, nil
	}
	if !g.applyDamageToPermanentLocked(ev.DamageTarget, ev.DamageAmount, t.deathtouch) {
		return 0, ErrCardNotFound
	}
	g.emitDealDamageLocked(ev, t)
	g.creditLifelinkLocked(t, ev.DamageSource, ev.DamageAmount)
	return ev.DamageAmount, nil
}

// emitDealDamageLocked emits the one EventDealDamage a settled damage
// event produces, with the Actor, Combat and CombatStep fields the tail
// carries so "deals combat damage" triggers — and the public log's
// combat_step tag — key off the same shape whether or not the event
// paused. This is the only place Event.CombatStep is written.
//
// Caller must hold g.mu.
func (g *Game) emitDealDamageLocked(ev *ReplacementEvent, t *damageTail) {
	g.EmitEvent(Event{
		Kind:       EventDealDamage,
		Actor:      t.actor,
		Source:     ev.DamageSource,
		Target:     ev.DamageTarget,
		Amount:     ev.DamageAmount,
		Combat:     t.combat,
		CombatStep: t.combatStep,
	})
}

// creditLifelinkLocked credits the tail's lifelink beneficiary with
// life equal to the damage dealt (CR 702.15a). No-op when the source
// had no lifelink, or when this entry point has never applied it.
//
// #482: this is life GAIN — CR 702.15b, "damage dealt by a source with
// lifelink causes that source's controller to gain that much life" —
// so it runs the CR 614 life window like every other life change
// instead of writing the total directly. Rhox Faithmender doubles its
// own printed lifelink because of this one line.
//
// The credit stays a SEPARATE event from the damage. The damage
// replacements settled the amount already (CR 120.3); the life gain is
// then replaced on its own terms, which is why a life doubler doubles
// the gain without touching the damage.
//
// An error here means the beneficiary is no longer seated. There is
// nothing to gain and it is never a reason to fail the damage.
//
// Caller must hold g.mu.
func (g *Game) creditLifelinkLocked(t *damageTail, sourceID uuid.UUID, amount int) {
	if amount <= 0 || t.lifelinkTo == uuid.Nil {
		return
	}
	_ = g.ChangePlayerLifeForEffect(sourceID, t.lifelinkTo, amount)
}
