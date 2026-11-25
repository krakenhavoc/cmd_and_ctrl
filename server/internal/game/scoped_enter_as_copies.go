package game

import (
	"fmt"

	"github.com/google/uuid"
)

// scoped_enter_as_copies.go — "The next time one or more creatures or
// planeswalkers enter this turn, they enter as copies of the chosen
// creature" (Mystic Reflection, #6160).
//
// A CR 614.1c replacement effect a resolving spell creates (CR 611.2), held
// as a scoped record so it outlives the instant and survives a restore. It
// is the resolved, one-shot sibling of a Clone's own "enter as a copy"
// (copy_choice.go): no prompt, because the object to copy was chosen as the
// spell resolved, and not the permanent's own ability, so it reaches every
// creature or planeswalker entering, under anyone's control, token or card.
//
// THE CHOSEN OBJECT. Mod.Objects[0] is the creature chosen, pinned to the
// object it was (CR 400.7). Its copiable values (CR 707.2) are read as the
// permanents enter: from the live object while it is on the battlefield,
// and from its last-known information once it has left
// (PermanentForEffect, CR 608.2h), so a chosen creature that died in
// between is still what they copy.
//
// "THE NEXT TIME ONE OR MORE … ENTER". The record is spent by the first
// entry of a creature or planeswalker after it begins, and that entry is
// every permanent entering SIMULTANEOUSLY with it: the members of one
// simultaneous entry (entry_batch.go) all have their CR 614 windows walked
// before any of them lands, so the first window to apply stamps the whole
// entry onto the record (Objects[1:], with SpentBatch the event batch it
// happened in) and the other members' windows still match. Anything that
// enters later — a separate entry in the same resolution, or any entry
// once play has moved on — matches nothing. The spent record is dropped
// as the next event batch begins (dropSpentNextShieldsLocked).
//
// "Creatures or planeswalkers" is judged as each permanent would exist on
// the battlefield (CR 614.12): a Clone that has already been told what to
// enter as is that copy. A planeswalker that becomes a copy of a creature
// is that creature, so it has no loyalty to enter with (CR 306.5b reads the
// copy's characteristics). The mod kind is an on-disk identity: never
// renamed, never reused.

// ModNextEntersAsCopies is "the next time one or more creatures or
// planeswalkers enter this turn, they enter as copies of Objects[0]". A
// replacement, read by the replacement gather rather than the layer pass.
const ModNextEntersAsCopies ModKind = "nextEntersAsCopies"

// NextCreaturesOrPlaneswalkersEnterAsCopiesThisTurnForEffect registers
// Mystic Reflection's replacement: until end of turn, the next entry of one
// or more creatures or planeswalkers enters as copies of `chosen`, a
// permanent on the battlefield now. `controller` controls the replacement
// effect. Returns false, registering nothing, when `chosen` is not on the
// battlefield.
//
// Caller must hold g.mu (write) — every caller is a resolving effect.
func (g *Game) NextCreaturesOrPlaneswalkersEnterAsCopiesThisTurnForEffect(sourceID, controller, chosen uuid.UUID, label string) bool {
	c := findBattlefieldCard(g, chosen)
	if c == nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, controller,
		[]Mod{{Kind: ModNextEntersAsCopies, Objects: []ObjectRef{{ID: chosen, Epoch: c.ObjectEpoch}}}},
		g.UntilEndOfTurnDuration(), label)
}

// nextEntersAsCopiesModProblem is the kind's shape: the chosen object
// first, then the entry that spent it, which carries the batch it was
// spent in. "" when sound.
func nextEntersAsCopiesModProblem(m Mod) string {
	if m.Kind != ModNextEntersAsCopies {
		return ""
	}
	if len(m.Objects) == 0 || m.Objects[0].ID == uuid.Nil {
		return "a nextEntersAsCopies replacement names no object to copy"
	}
	if (len(m.Objects) > 1) != (m.SpentBatch != 0) {
		return fmt.Sprintf("a nextEntersAsCopies replacement's spending entry (%d objects) and batch (%d) disagree", len(m.Objects)-1, m.SpentBatch)
	}
	return ""
}

// nextEntersAsCopiesAppliesLocked is the kind's test: a creature or
// planeswalker moving onto the battlefield, while the record is unspent or
// as one of the permanents of the entry that spent it, with copiable
// values still to be read for the chosen object.
//
// Caller must hold g.mu.
func nextEntersAsCopiesAppliesLocked(g *Game, m Mod, ev *ReplacementEvent) bool {
	if ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || ev.Canceled || len(m.Objects) == 0 {
		return false
	}
	if m.SpentBatch != 0 {
		if m.SpentBatch != g.eventBatch || !nextEntersAsCopiesSpentBy(m, ev.CardID) {
			return false
		}
	}
	chosen := m.Objects[0]
	if c := findBattlefieldCard(g, chosen.ID); c == nil || c.ObjectEpoch != chosen.Epoch {
		if _, ok := g.lastKnownPermanentLocked(chosen); !ok {
			return false
		}
	}
	if ev.EntersAsCopyOf != nil {
		// CR 614.12: as it would exist on the battlefield.
		v := ev.EntersAsCopyOf
		return v.HasCardType("Creature") || v.HasCardType("Planeswalker")
	}
	entering, ok := g.LookupCardForEffect(ev.CardID)
	return ok && (entering.IsCreature() || entering.IsPlaneswalker())
}

// nextEntersAsCopiesSpentBy reports whether `id` is one of the permanents
// of the entry that spent the record.
func nextEntersAsCopiesSpentBy(m Mod, id uuid.UUID) bool {
	for _, o := range m.Objects[1:] {
		if o.ID == id {
			return true
		}
	}
	return false
}

// applyNextEntersAsCopiesLocked is the kind's Replace: the permanent
// enters wearing the chosen object's copiable values (CR 707.2), and an
// unspent record is spent by the entry this permanent is part of.
//
// Caller must hold g.mu (write).
func (g *Game) applyNextEntersAsCopiesLocked(e ScopedEffect, mod int, m Mod, ev *ReplacementEvent) {
	chosen := m.Objects[0]
	info, ok := g.PermanentForEffect(chosen)
	if ok && info.Copiable != nil {
		values := info.Copiable.Clone()
		ev.EntersAsCopyOf = &values
		ev.copySourceID = chosen.ID
		ev.copyUntilEndOfTurn = false
	}
	if m.SpentBatch != 0 {
		return
	}
	entry := []ObjectRef{{ID: ev.CardID}}
	if ev.entryTail != nil && ev.entryTail.batch != nil {
		entry = entry[:0]
		for _, member := range ev.entryTail.batch.members {
			entry = append(entry, ObjectRef{ID: member.cardID})
		}
	}
	g.markNextEntersAsCopiesSpentLocked(e.Seq, mod, g.currentEventBatchLocked(), entry)
}

// markNextEntersAsCopiesSpentLocked records the entry that spent the
// record. COPY ON WRITE, as markNextShieldSpentLocked: a clone taken before
// the entry shares the slice and must keep the unspent record, so an undo
// rewinds it.
//
// Caller must hold g.mu (write).
func (g *Game) markNextEntersAsCopiesSpentLocked(seq int64, mod int, batch uint64, entry []ObjectRef) {
	i, ok := g.scopedEffectIndexBySeqLocked(seq)
	if !ok {
		return
	}
	next := cloneScopedEffects(g.ScopedEffects)
	rec := next[i]
	rec.Mods = cloneMods(rec.Mods)
	objs := append([]ObjectRef{rec.Mods[mod].Objects[0]}, entry...)
	rec.Mods[mod].Objects = objs
	rec.Mods[mod].SpentBatch = batch
	next[i] = rec
	g.ScopedEffects = next
}
