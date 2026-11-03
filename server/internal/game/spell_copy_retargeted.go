package game

import "github.com/google/uuid"

// spell_copy_retargeted.go — "that player copies that spell for each other
// creature they control that the spell could target. Each copy targets a
// different one of those creatures" (Mirrorwing Dragon, #4834).
//
// The copy's targets are not chosen by anybody: every target of the
// original that is the object `from` is the object `to` in the copy, and
// every other target is kept (CR 707.10: the copy copies the choices made
// for the original). A copy is made only for a `to` the spell could
// target: the whole new list must pass the announce gate, judged for the
// copy's controller and the spell's characteristics, so a creature with
// hexproof from that player, protection from the spell's colour or one
// the clause's predicate excludes is ignored, and a spell with several
// targets needs `to` to be legal for each of them (the ruling).
//
// The spell is the one on the stack or, once it has left, as it last
// existed there (CR 608.2h): the copy effect names the spell, it does not
// target it, as storm's does not (CopyLastKnownSpellForEffect).

// retargetedSpellCopyLocked finds the spell and builds the copy's target
// list, reporting false when there is no spell to copy, it can't be
// copied, nothing of it is `from`, or the new list is not legal.
//
// Caller must hold g.mu.
func (g *Game) retargetedSpellCopyLocked(spellID, controller, from, to uuid.UUID) (Card, *StackItem, []TargetRef, bool) {
	src, item, ok := g.stackSpellLocked(spellID)
	if !ok {
		src, item, ok = g.lastKnownSpellLocked(spellID)
	}
	if !ok || item == nil || item.Kind != StackItemSpell || spellPrintsCantBeCopied(src) {
		return Card{}, nil, nil, false
	}
	next := make([]TargetRef, len(item.Targets))
	moved := false
	for i, t := range item.Targets {
		if t.Kind == TargetCard && t.ID == from {
			t.ID = to
			moved = true
		}
		next[i] = t
	}
	if !moved {
		return Card{}, nil, nil, false
	}
	ts := SourceSnapshot(controller, SourceCharacteristics(&src))
	if err := g.validateAnnouncedTargetsWithLocked(ts, g.itemAnnouncedClauses(item), next, nil, item.Targets); err != nil {
		return Card{}, nil, nil, false
	}
	return src, item, next, true
}

// SpellCouldTargetInsteadForEffect reports whether a copy of the spell
// `spellID`, controlled by `controller`, could have every target that is
// `from` be `to` instead: Mirrorwing Dragon's "each other creature they
// control that the spell could target".
//
// Caller must hold g.mu.
func (g *Game) SpellCouldTargetInsteadForEffect(spellID, controller, from, to uuid.UUID) bool {
	_, _, _, ok := g.retargetedSpellCopyLocked(spellID, controller, from, to)
	return ok
}

// CopySpellRetargetedForEffect puts a copy of the spell `spellID` on the
// stack under `controller`, its every target that is `from` being `to`
// instead, and reports whether it did. Nothing is asked. A division (CR
// 601.2d) travels with its slot, as a retarget's does (CR 115.7c).
//
// Caller must hold g.mu.
func (g *Game) CopySpellRetargetedForEffect(spellID, controller, from, to uuid.UUID) bool {
	src, item, next, ok := g.retargetedSpellCopyLocked(spellID, controller, from, to)
	if !ok {
		return false
	}
	g.createSpellCopyLocked(src, item, controller, next)
	top, err := g.Stack.Top()
	if err != nil {
		return false
	}
	if meta := g.StackMeta[top.InstanceID]; meta != nil && meta.IsCopy {
		meta.Distribution = remapDistributionLocked(meta.Distribution, item.Targets, next)
	}
	return true
}
