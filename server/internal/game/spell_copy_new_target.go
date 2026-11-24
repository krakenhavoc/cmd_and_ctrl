package game

import "github.com/google/uuid"

// spell_copy_new_target.go — "copy that spell, and the copy targets the
// chosen player or permanent" (Zevlor, Elturel Exile, #6094): CR 707.10e,
// an effect that copies a spell and SPECIFIES a new target for the copy.
//
// It is spell_copy_retargeted.go's "every target that is `from` is `to`"
// with the one difference 707.10e prints: EVERY target of the copy is the
// specified player or object, whatever the original's were, and that new
// target may be a player where the original's was an object, or the other
// way round. The copy is created only when the specified one is a legal
// target for each instance of the word "target" (the whole new list passes
// the announce gate, judged for the copy's controller and the spell's
// characteristics); otherwise it isn't created. Nobody is asked.
//
// The spell is the one on the stack or, once it has left, as it last
// existed there (CR 608.2h): the copy effect names the spell, it does not
// target it.

// spellCopyWithNewTargetLocked builds the copy's target list, reporting
// false when there is no spell to copy, it can't be copied, it has no
// targets, or the new list is not legal. Caller must hold g.mu.
func (g *Game) spellCopyWithNewTargetLocked(spellID, controller uuid.UUID, to TargetRef) (Card, *StackItem, []TargetRef, bool) {
	src, item, ok := g.stackSpellLocked(spellID)
	if !ok {
		src, item, ok = g.lastKnownSpellLocked(spellID)
	}
	if !ok || item == nil || item.Kind != StackItemSpell || spellPrintsCantBeCopied(src) || len(item.Targets) == 0 {
		return Card{}, nil, nil, false
	}
	if to.Kind != TargetCard && to.Kind != TargetPlayer {
		return Card{}, nil, nil, false
	}
	next := make([]TargetRef, len(item.Targets))
	for i, t := range item.Targets {
		t.Kind, t.ID, t.Epoch = to.Kind, to.ID, to.Epoch
		next[i] = t
	}
	ts := SourceSnapshot(controller, SourceCharacteristics(&src))
	ts.spell = true
	if err := g.validateAnnouncedTargetsWithLocked(ts, g.itemAnnouncedClauses(item), next, nil, item.Targets); err != nil {
		return Card{}, nil, nil, false
	}
	return src, item, next, true
}

// SpellSoleTargetForEffect is StackItemSoleTargetForEffect for a spell that
// may have left the stack: the one player or object the spell targeted, on
// the stack or as it last existed there (CR 608.2h), and true when it
// targeted exactly one different player or object (CR 115.9c). Caller must
// hold g.mu.
func (g *Game) SpellSoleTargetForEffect(spellID uuid.UUID) (TargetRef, bool) {
	if t, ok := g.StackItemSoleTargetForEffect(spellID); ok {
		return t, true
	}
	if g.StackMeta[spellID] != nil {
		return TargetRef{}, false
	}
	_, item, ok := g.lastKnownSpellLocked(spellID)
	if !ok || item == nil {
		return TargetRef{}, false
	}
	var sole TargetRef
	found := false
	for _, t := range item.Targets {
		if t.Kind != TargetCard && t.Kind != TargetPlayer {
			continue
		}
		if !found {
			sole, found = t, true
			continue
		}
		if t.Kind != sole.Kind || t.ID != sole.ID {
			return TargetRef{}, false
		}
	}
	return sole, found
}

// CopySpellWithNewTargetForEffect puts a copy of the spell `spellID` on the
// stack under `controller` with every target being `to` (CR 707.10e), and
// reports whether it did: a `to` that is not a legal target for each
// instance of the word "target" creates no copy. It is one copy
// instruction, so a replacement of the instruction applies to it once: an
// additional copy "you may choose new targets for" (Twinning Staff) is a
// copy of the spell with its original targets, offered new ones (CR
// 707.10c), as CopySpellRetargetedEachForEffect makes it.
//
// Caller must hold g.mu.
func (g *Game) CopySpellWithNewTargetForEffect(spellID, controller uuid.UUID, to TargetRef) bool {
	src, item, next, ok := g.spellCopyWithNewTargetLocked(spellID, controller, to)
	if !ok {
		return false
	}
	total, additional := g.replaceSpellCopyLocked(spellID, controller, 1)
	for i := 0; i < total; i++ {
		if i >= total-additional {
			g.copySpellFromLocked(src, item, controller, true, nil, nil)
			continue
		}
		g.putRetargetedSpellCopyLocked(src, item, controller, next)
	}
	return total > 0
}
