package game

import (
	"slices"

	"github.com/google/uuid"
)

// combat_damage_hold.go — #3665: a combat damage step that waits on a
// damage-assignment prompt deals none of its damage until every one of
// its prompts is answered.
//
// THE RULE. CR 510.1 assigns all of the step's combat damage first —
// the attackers' and the blockers', including the divisions a CR
// 510.1c / 510.1d prompt asks for — and CR 510.2 then deals it all at
// once. State-based actions are checked only after that (CR 704.3), so
// no creature in the step can die before another one's damage has been
// assigned.
//
// THE BUG. assignAndDealCombatDamageLocked used to mark every damage it
// could decide on its own (unblocked attackers, single blocks, the
// blockers' damage back) and queue the prompts alongside. The step then
// ran the state-based actions: an attacker blocked by two 1/1s died to
// their damage, and its "dies" triggers fired, while its own controller
// was still being asked how to divide its damage — and the blockers it
// went on to kill died in a second, separate sweep.
//
// THE FIX. When the step queues any prompt, the damage it decided
// without one is HELD here (heldCombatDamage.marks) instead of dealt,
// and each answer is held with it as it comes in (answers). The answer
// that closes the last prompt deals the lot as one event batch, inside
// one damage stage and one state-trigger hold, exactly as a step with
// no prompt deals it; the state-based actions run once, after. A step
// with no prompt is unchanged: it deals its marks in the same order it
// always did, before it returns.
//
// WHY THE BOARD CAN BE TRUSTED LATER. A damage-assignment prompt blocks
// the table (choice_gate.go), so nothing is cast, activated or advanced
// between the step and its last answer: the sources and recipients the
// marks name are the ones on the battlefield when the step assigned
// them. The one thing that can change the board is a player leaving
// the game (CR 800.4a). Their prompts are dropped, their objects go,
// and the next state check (stateChecksLocked) deals what is held: a
// mark whose recipient has gone deals nothing, which is CR 510.1b's own
// rule for an attack target that left, and a mark whose source has gone
// is dealt with its last-known information.
//
// SNAPSHOTS. The hold is plain data, but the persisted snapshot does
// not carry it: a table holding one counts a continuation in
// ContinuationCensus.ChoiceResumeFrames, so it is not a restore point,
// and a restart rewinds to the last clean boundary before the step.
// Clone carries it, so an undo across an answer re-holds it.
//
// A restore point written before this change can hold a prompt with no
// hold behind it: its blockers' damage was dealt when the prompt was
// queued. ResolveDamageAssignment deals such an answer at once, as it
// always did.

// heldCombatDamage is one combat damage step's damage, waiting on its
// damage-assignment prompts. Copy on write: it is replaced, never
// changed in place, so Clone may share it.
type heldCombatDamage struct {
	// step is the step's Event.CombatStep value, which tags every mark.
	step string
	// marks is the damage the step decided without a prompt, in the
	// order the step decided it.
	marks []heldCombatMark
	// answers is each answered prompt's division, in answer order.
	answers []heldDamageAssignment
}

// heldCombatMark is one combat damage event the step decided by itself.
type heldCombatMark struct {
	source uuid.UUID
	target uuid.UUID
	amount int
	// toAttackTarget deals it as an attacker's damage to what it
	// attacked (dealCombatDamageToAttackTargetLocked: a player, a
	// planeswalker or a battle); otherwise it is marked on a permanent.
	toAttackTarget bool
}

// heldDamageAssignment is one answered damage-assignment prompt.
type heldDamageAssignment struct {
	frame   *DamageAssignmentFrame
	ordered []DamageAssignmentEntry
	trample int
}

// dealCombatMarkLocked deals one decided combat damage event through the
// path the step has always used for it.
//
// Caller must hold g.mu.
func (g *Game) dealCombatMarkLocked(m heldCombatMark, step string) {
	if m.toAttackTarget {
		g.dealCombatDamageToAttackTargetLocked(m.target, m.source, m.amount, step)
		return
	}
	g.markCombatDamageOnCardLocked(m.target, m.amount, m.source, step)
}

// damageAssignmentPendingLocked reports whether any damage-assignment
// prompt is still open.
//
// Caller must hold g.mu.
func (g *Game) damageAssignmentPendingLocked() bool {
	return slices.ContainsFunc(g.PendingChoices, func(c *PendingChoice) bool {
		return c != nil && c.Kind == PendingChoiceDamageAssignment
	})
}

// holdDamageAssignmentAnswerLocked adds an answered prompt to the held
// step. Copy on write.
//
// Caller must hold g.mu.
func (g *Game) holdDamageAssignmentAnswerLocked(frame *DamageAssignmentFrame, ordered []DamageAssignmentEntry, trample int) {
	h := g.heldCombatDamage
	if h == nil {
		return
	}
	next := &heldCombatDamage{
		step:  h.step,
		marks: h.marks,
		answers: append(slices.Clip(h.answers), heldDamageAssignment{
			frame:   frame,
			ordered: slices.Clone(ordered),
			trample: trample,
		}),
	}
	g.heldCombatDamage = next
}

// heldCombatDamageToLocked is the damage the held step has already
// assigned to the permanent `id` — what CR 702.19b means by "damage from
// other creatures that's being assigned during the same combat damage
// step", which a trampler's lethal damage takes into account. Zero when
// nothing is held.
//
// Caller must hold g.mu.
func (g *Game) heldCombatDamageToLocked(id uuid.UUID) int {
	h := g.heldCombatDamage
	if h == nil || id == uuid.Nil {
		return 0
	}
	n := 0
	for _, m := range h.marks {
		if m.target == id && !m.toAttackTarget {
			n += m.amount
		}
	}
	for _, a := range h.answers {
		for _, e := range a.ordered {
			if e.BlockerID == id {
				n += e.Amount
			}
		}
	}
	return n
}

// HeldCombatDamageToForEffect is the exported read of
// heldCombatDamageToLocked, for the legal-move enumerator's lethal
// damage. Caller must hold g.mu (read or write).
func (g *Game) HeldCombatDamageToForEffect(id uuid.UUID) int {
	return g.heldCombatDamageToLocked(id)
}

// settleHeldCombatDamageLocked deals the held step once no
// damage-assignment prompt is left open: every decided mark, then every
// answered division, as one batch of simultaneous damage (CR 510.2).
// The caller runs the state-based actions after. Reports whether it
// dealt anything.
//
// Caller must hold g.mu.
func (g *Game) settleHeldCombatDamageLocked() bool {
	h := g.heldCombatDamage
	if h == nil || g.damageAssignmentPendingLocked() {
		return false
	}
	g.heldCombatDamage = nil
	if g.State != StateActive {
		return false
	}
	// The same two sections the step opens around its own deals: one
	// CR 603.8 check after all of it (ADR 0107 §1), and one CR 615.7
	// group while a charged shield is live (ADR 0108 §7).
	defer g.holdStateTriggersLocked()()
	if g.anyChargedShieldLocked() {
		if prev, staged := g.openDamageStageLocked(g.combatDamageInstanceLocked(), true, false); staged {
			defer g.closeDamageStageLocked(prev)
		}
	}
	for _, m := range h.marks {
		g.dealCombatMarkLocked(m, h.step)
	}
	for _, a := range h.answers {
		g.applyDamageAssignmentLocked(a.frame, a.ordered, a.trample)
	}
	return true
}
