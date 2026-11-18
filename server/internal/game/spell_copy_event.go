package game

import "github.com/google/uuid"

// spell_copy_event.go — copying a spell is a replaceable event (#5918,
// CR 707.10, 614.1a).
//
// # The card that needs it
//
// Twinning Staff: "If you would copy a spell one or more times, instead
// copy it that many times plus an additional time. You may choose new
// targets for the additional copy." The replacement modifies one copy
// INSTRUCTION, before any copy exists, and it adds one copy per
// instruction: storm's "copy it for each other spell that was cast before
// it this turn" with a storm count of three is one instruction, so it
// becomes four copies, not six.
//
// # The one door
//
// Every spell copy the engine makes is made by one of the doors below,
// and each opens the window (RepEventCopySpell) once per instruction:
//
//   - CopySpellTimesForEffect: "copy it N times" (storm, replicate,
//     "copy that spell twice", effects.CopySpell with its Count).
//   - CopySpellForEffect, CopyLastKnownSpellForEffect and
//     CopyLastKnownSpellGainingKeywordsForEffect: one copy. Casualty and
//     conspire make one instruction per paid cost, because each instance
//     of the keyword triggers on its own (CR 702.153b, 702.78b).
//   - CopySpellRetargetedEachForEffect (and CopySpellRetargetedForEffect,
//     its one-creature form): "copy that spell for each other creature
//     … each copy targets a different one" (CR 707.10e) is one
//     instruction of that many copies.
//   - CopySpellKeepingTargetsForEffect: Chef's Kiss's copy.
//
// None of them is a copy of a card that is then cast (CR 707.12), which
// is a cast and not a copy of a spell, and none copies an ability.
//
// # What a replacement adds
//
// An additional copy is a copy of the same spell under the same player,
// made the way the instruction makes its copies (an "except" clause, a
// keyword it gains), except that its controller may choose new targets
// for it (CR 707.10c) whatever the instruction says about the others. A
// copy made "for" an object (CR 707.10e) names that object in the
// instruction, not the spell, so the additional one starts from the
// original spell's targets.
//
// # Why it never pauses
//
// The window settles inline (mustSettleNow, as a dice roll's does): the
// only replacements of this kind add copies, and adding commutes, so the
// CR 616.1 order the affected player would choose among several (two
// Staffs) cannot change the outcome. Each applies once (CR 614.5), so two
// Staffs add two copies. Nothing about the event outlives the call, so it
// adds nothing to a snapshot: each copy that asks for new targets is its
// own queued copyFrame, which a restore point already carries.

// SpellCopy is one instruction to copy a spell (CR 707.10).
type SpellCopy struct {
	// Spell is the spell to copy; Controller the player who copies it and
	// controls the copies (CR 707.10).
	Spell, Controller uuid.UUID
	// Times is how many copies the instruction makes. Zero is one.
	Times int
	// MayChooseNewTargets is "You may choose new targets for the copies"
	// (CR 707.10c).
	MayChooseNewTargets bool
	// FromLastKnown copies a spell that has left the stack from its
	// last-known information (CR 608.2h): an instruction that names the
	// spell without targeting it (storm). See CopyLastKnownSpellForEffect.
	FromLastKnown bool
	// Except is the instruction's "except …" clause (CR 707.9), applied
	// to each copy. Nil for a plain copy.
	Except func(v *PrintedValues)
}

// CopySpellTimesForEffect carries out one copy instruction: the window
// first, then each copy, in order, each with its own CR 707.10c offer.
//
// Errors: ErrCardNotFound when the spell is not on the stack and, for
// FromLastKnown, has no last-known record (see CopySpellForEffect).
//
// Caller must hold g.mu.
func (g *Game) CopySpellTimesForEffect(c SpellCopy) error {
	src, item, ok := g.stackSpellLocked(c.Spell)
	if !ok && c.FromLastKnown {
		src, item, ok = g.lastKnownSpellLocked(c.Spell)
	}
	if !ok {
		return ErrCardNotFound
	}
	n := c.Times
	if n <= 0 {
		n = 1
	}
	g.copySpellTimesLocked(src, item, c.Controller, n, c.MayChooseNewTargets, c.Except, nil)
	return nil
}

// copySpellTimesLocked is the instruction once the spell is found: an
// uncopyable spell makes nothing and opens no window; otherwise the
// window settles the number, and the copies are made in order, the
// additional ones last and each offered new targets. Caller must hold
// g.mu.
func (g *Game) copySpellTimesLocked(src Card, item *StackItem, controller uuid.UUID, n int, mayChooseNewTargets bool, except func(v *PrintedValues), gains *copyGrant) {
	if spellPrintsCantBeCopied(src) {
		return
	}
	total, additional := g.replaceSpellCopyLocked(src.InstanceID, controller, n)
	for i := 0; i < total; i++ {
		g.copySpellFromLocked(src, item, controller, mayChooseNewTargets || i >= total-additional, except, gains)
	}
}

// replaceSpellCopyLocked opens the RepEventCopySpell window for an
// instruction to copy `spellID` `n` times under `controller`, and returns
// how many copies to make and how many of the last of them were added
// with "you may choose new targets". Caller must hold g.mu (write).
func (g *Game) replaceSpellCopyLocked(spellID, controller uuid.UUID, n int) (total, additional int) {
	if n <= 0 {
		return 0, 0
	}
	ev := &ReplacementEvent{
		Kind: RepEventCopySpell, Actor: controller,
		CopySpellID: spellID, CopyPlayer: controller, CopyCount: n,
		mustSettleNow: true,
	}
	out, err := g.applyReplacementsLocked(ev)
	g.clearReplacementEventLocked(ev.ID)
	if err != nil {
		// The iteration cap: the event as it stands.
		out = ev
	}
	if out == nil || out.Canceled || out.CopyCount <= 0 {
		return 0, 0
	}
	additional = min(max(out.copyAdditional, 0), out.CopyCount)
	return out.CopyCount, additional
}

// CopyOneMoreWithNewTargets is a replacement's "instead copy it that many
// times plus an additional time. You may choose new targets for the
// additional copy" (Twinning Staff, #5918) applied to a copy
// instruction's window. Each such replacement applies once per
// instruction (CR 614.5).
func (ev *ReplacementEvent) CopyOneMoreWithNewTargets() {
	if ev.Kind != RepEventCopySpell || ev.CopyCount <= 0 {
		return
	}
	ev.CopyCount++
	ev.copyAdditional++
}

// CopySpellRetargetedEachForEffect is "that player copies that spell for
// each other creature they control that the spell could target. Each copy
// targets a different one of those creatures" (Precursor Golem, Mirrorwing
// Dragon, Zada): ONE copy instruction (CR 707.10e) of a copy for each `to`
// the spell could target instead of `from`, in the order given. It reports
// how many copies it made for those objects. A replacement's additional
// copy (Twinning Staff) is a copy of the spell with its original targets,
// offered new ones (CR 707.10c).
//
// Caller must hold g.mu.
func (g *Game) CopySpellRetargetedEachForEffect(spellID, controller, from uuid.UUID, tos []uuid.UUID) int {
	type retargeted struct {
		src  Card
		item *StackItem
		next []TargetRef
	}
	var copies []retargeted
	for _, to := range tos {
		if src, item, next, ok := g.retargetedSpellCopyLocked(spellID, controller, from, to); ok {
			copies = append(copies, retargeted{src, item, next})
		}
	}
	if len(copies) == 0 {
		return 0
	}
	total, _ := g.replaceSpellCopyLocked(spellID, controller, len(copies))
	made := 0
	for _, c := range copies {
		if made >= total {
			break
		}
		g.putRetargetedSpellCopyLocked(c.src, c.item, controller, c.next)
		made++
	}
	first := copies[0]
	for i := made; i < total; i++ {
		g.copySpellFromLocked(first.src, first.item, controller, true, nil, nil)
	}
	return made
}

// putRetargetedSpellCopyLocked puts one copy of the spell on the stack
// with the targets `next`, its division travelling with its slot (CR
// 115.7c). Caller must hold g.mu.
func (g *Game) putRetargetedSpellCopyLocked(src Card, item *StackItem, controller uuid.UUID, next []TargetRef) {
	g.createSpellCopyLocked(src, item, controller, next)
	top, err := g.Stack.Top()
	if err != nil {
		return
	}
	if meta := g.StackMeta[top.InstanceID]; meta != nil && meta.IsCopy {
		meta.Distribution = remapDistributionLocked(meta.Distribution, item.Targets, next)
	}
}
