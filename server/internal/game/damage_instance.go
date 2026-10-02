package game

// damage_instance.go — ADR 0108 "Shared machinery: the damage
// instance" (owner decision 4, Delivery PR 0): which damage events
// happen AT THE SAME TIME.
//
// THE RULES.
//
//   - CR 615.8: a "next time" shield "prevent[s] the next instance of
//     damage from that source, regardless of how much damage that is.
//     Once an instance of damage from that source has been prevented,
//     any subsequent instances of damage that would be dealt by that
//     source are dealt normally."
//   - CR 615.5: a prevention effect's additional effect "takes place
//     immediately afterward."
//   - CR 615.13: a "whenever damage is prevented" ability triggers "each
//     time a prevention effect is applied to one or more simultaneous
//     damage events".
//   - CR 510.2: "all combat damage that's been assigned is dealt
//     simultaneously."
//   - CR 608.2c: a spell's instructions are followed in the order
//     written, so "it deals 2 damage to you. Then it deals 2 damage to
//     you" is two instructions, and two instances of damage.
//
// THE UNIT. The engine opens one damage event per (source, recipient),
// and ADR 0107 used the event batch (event_batch.go) as its stand-in for
// "at the same time". The batch is too wide: everything one resolution
// emits is one batch, so two separate instructions inside it were one
// instance and a next-damage shield prevented both. A DamageInstance is
// the narrower unit:
//
//   - each damage INSTRUCTION takes the next ID from Game.damageInstanceSeq
//     and stamps it on every damage event it opens
//     (ReplacementEvent.DamageInstance). An instruction is one
//     DealDamage…ForEffect call, the whole of a DealDamageEachThenForEffect
//     walk, or — for a printed instruction the catalog writes as a loop
//     ("deals 2 damage to each creature", a fight, "4 damage to the first
//     target and 3 to each other") — everything inside one
//     DamageInstanceForEffect scope;
//   - each COMBAT DAMAGE STEP is one instance (CR 510.2): the first combat
//     damage event of a step takes an ID, and every combat damage event of
//     the same step (the same event batch, since each combat damage step
//     opens its own, CR 510.4) shares it, including the events a CR 510.1c
//     assignment prompt opens when it is answered.
//
// The stamp rides the event, so a CR 616 ordering prompt resumes with the
// instance the event was opened in (cloneReplacementResume copies it with
// the event). It is TRANSIENT: never captured. A table paused on such a
// prompt is not a restore point anyway (ContinuationCensus counts the
// frame).
//
// THE READERS. "The next time" is spent per instance
// (Mod.SpentInstance, prevent_next_from_source.go), and a shield's
// follow-up is owed per (record, instance) (PreventionFollowUp.Instance).
// ADR 0108 §7's divide_shield, §8's one application per recipient or per
// source, §10's Phyrexian Unlife life check and the `Next` field of
// multiplyDamage and redirectDamage read the same stamp, off
// ReplacementEvent.DamageInstance.
//
// WHEN AN INSTANCE ENDS. When the instruction is over and none of its
// events is still waiting on a CR 616 prompt: a single call as its event
// reaches its terminal outcome (runDamageTailLocked), an Each walk at its
// base case, a scope as its function returns. Ending flushes the
// follow-ups owed for that instance (CR 615.5: "immediately afterward").
// A combat damage step has no explicit end; its follow-ups run at ADR
// 0107's flush points, which stay where they were: as a player would
// next receive priority (runStateChecksLocked, after all of the step's
// damage, CR 510.2) and as play moves on (beginEventBatchLocked). Those
// points also catch an instance whose end was skipped because one of its
// events was paused.
//
// WHAT REACHES A SNAPSHOT. Only the references: Mod.SpentInstance and
// PreventionFollowUp.Instance, both additive within v7. The counter is not
// a key: Clone copies it, and restore sets it to the largest instance a
// restored record names (maxNamedDamageInstance), as it does
// scopedEffectSeq — which is all uniqueness needs, because no restored
// event carries a stamp. The open scope and the combat step's instance
// are not captured either: a game restored in the middle of a combat
// damage step (a CR 510.1c assignment prompt is pure data, so that is a
// restore point) gives the rest of the step a new instance. That can
// only matter for a source whose damage in the step straddles the
// restore, and a source's combat damage is all opened at once — directly,
// or all of it from its one assignment prompt.

// DamageInstance identifies one instance of damage: every damage event
// opened by one damage instruction, or by one combat damage step. Zero is
// "no instance" — an event built by hand outside the engine's entry
// points — and readers treat it as the ADR 0107 event batch did.
type DamageInstance uint64

// nextDamageInstanceLocked takes the next instance ID.
//
// Caller must hold g.mu (write).
func (g *Game) nextDamageInstanceLocked() DamageInstance {
	g.damageInstanceSeq++
	return DamageInstance(g.damageInstanceSeq)
}

// damageInstructionLocked is the instance a damage instruction's events
// carry: the open scope's when one is open (owned false: the scope ends
// it), and otherwise a new one the instruction owns and must end.
//
// Caller must hold g.mu (write).
func (g *Game) damageInstructionLocked() (inst DamageInstance, owned bool) {
	if g.openDamageInstance != 0 {
		return g.openDamageInstance, false
	}
	return g.nextDamageInstanceLocked(), true
}

// stampEffectDamageInstanceLocked stamps a non-combat damage event as it
// is opened. A non-zero `inst` is an Each walk's, joined as it stands;
// zero makes the event its own instruction (damageInstructionLocked), and
// an instance it takes for itself ends with its terminal outcome
// (damageTail.endsInstance).
//
// Caller must hold g.mu (write).
func (g *Game) stampEffectDamageInstanceLocked(ev *ReplacementEvent, inst DamageInstance) {
	owned := false
	if inst == 0 {
		inst, owned = g.damageInstructionLocked()
	}
	ev.DamageInstance = inst
	if ev.damageTail != nil {
		ev.damageTail.endsInstance = owned
	}
}

// combatDamageInstanceLocked is the instance of the combat damage step
// being dealt (CR 510.2): the one taken by the step's first combat damage
// event, for every later one in the same event batch. Each combat damage
// step opens a batch of its own (advanceCursorLocked), so the first-strike
// and the regular step are two instances (CR 510.4).
//
// Caller must hold g.mu (write).
func (g *Game) combatDamageInstanceLocked() DamageInstance {
	batch := g.currentEventBatchLocked()
	if g.combatDamageInstance == 0 || g.combatDamageInstanceBatch != batch {
		g.combatDamageInstance = g.nextDamageInstanceLocked()
		g.combatDamageInstanceBatch = batch
	}
	return g.combatDamageInstance
}

// DamageInstanceForEffect runs fn as ONE damage instruction: every damage
// event the DealDamage… entry points open while it runs is one instance
// (CR 615.8). It is for a printed instruction the catalog writes as more
// than one call — "deals 2 damage to each creature and each player", a
// fight (CR 701.14a), "4 damage to the first target and 3 to each other",
// Fireball's divided damage — which is one instance in the rules.
//
// Two instructions must NOT share a scope: "it deals 2 damage to you.
// Then it deals 2 damage to you" is two instances (CR 608.2c), and a
// next-damage shield prevents only the first.
//
// Nested scopes are one instance: an inner scope (a shared helper that
// loops) inside an outer one (the card that calls two of them for one
// sentence) joins the outer's. A continuation that runs later — after a
// CR 616 prompt is answered — is outside the scope; an event that was
// opened inside it keeps its stamp.
//
// The instance ends as fn returns, which runs the follow-ups owed for it
// (CR 615.5), unless one of its events is still waiting on a CR 616
// prompt; then ADR 0107's flush points run them.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) DamageInstanceForEffect(fn func() error) error {
	if fn == nil {
		return nil
	}
	if g.openDamageInstance != 0 {
		return fn()
	}
	inst := g.nextDamageInstanceLocked()
	g.openDamageInstance = inst
	defer func() {
		g.openDamageInstance = 0
		g.endDamageInstanceLocked(inst)
	}()
	return fn()
}

// endDamageInstanceLocked is an instance ending: the follow-ups owed for
// it run now (CR 615.5: "immediately afterward"), once each with its
// total. Skipped while any event of the instance is waiting on a CR 616
// prompt — it has not settled, and running the follow-up now would run it
// a second time when the paused event adds to it. ADR 0107's flush points
// run it then.
//
// Caller must hold g.mu (write).
func (g *Game) endDamageInstanceLocked(inst DamageInstance) {
	if inst == 0 || g.damageInstancePausedLocked(inst) {
		return
	}
	g.flushPreventionFollowUpsForInstanceLocked(inst)
}

// damageInstancePausedLocked reports whether a damage event of `inst` is
// held by an open CR 616 prompt.
//
// Caller must hold g.mu.
func (g *Game) damageInstancePausedLocked(inst DamageInstance) bool {
	for _, c := range g.PendingChoices {
		if c == nil || c.replacementResume == nil || c.replacementResume.ev == nil {
			continue
		}
		if ev := c.replacementResume.ev; ev.Kind == RepEventDamage && ev.DamageInstance == inst {
			return true
		}
	}
	return false
}

// maxNamedDamageInstance is the largest instance a restored record names:
// a spent next-damage shield's SpentInstance, or an owed follow-up's
// Instance. Restore resumes the counter from it.
func maxNamedDamageInstance(records []ScopedEffect, followUps []PreventionFollowUp) uint64 {
	var n uint64
	for i := range records {
		for _, m := range records[i].Mods {
			if uint64(m.SpentInstance) > n {
				n = uint64(m.SpentInstance)
			}
		}
	}
	for _, f := range followUps {
		if uint64(f.Instance) > n {
			n = uint64(f.Instance)
		}
	}
	return n
}
