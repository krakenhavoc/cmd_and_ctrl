package game

import "github.com/google/uuid"

// retarget_change_all.go — RetargetChangeAll, "change the target or
// targets" over a spell or ability with more than one target (CR 115.7a,
// #6228, Psychic Battle):
//
//	"If an effect allows a player to 'change the target(s)' of a spell or
//	 ability, each target can be changed only to another legal target. If
//	 a target can't be changed to another legal target, the original
//	 target is unchanged, even if the original target is itself illegal
//	 by then. If all the targets aren't changed to other legal targets,
//	 none of them are changed."
//
// The chooser is asked slot by slot with the picks staged in
// PendingChoice.RetargetNext, as the related-target walk does (#4862);
// the stack item keeps its targets until the last slot is answered, and
// then they all change together (CR 115.7e judges only the final set).
// An optional change ("may change") may be declined at the first slot;
// after that, each slot must take a new target. A slot with no other
// legal target ends the walk and nothing changes.

// queueChangeAllRetargetLocked asks for the first real slot at or after
// `from`, or makes the change when every slot has its new target.
//
// Caller must hold g.mu.
func (g *Game) queueChangeAllRetargetLocked(offer RetargetOffer, from int, next []TargetRef) error {
	item := g.StackMeta[offer.ItemID]
	if item == nil {
		return nil
	}
	steps, src := g.itemAnnouncedClauses(item), g.retargetSourceLocked(item)
	first := true
	for i := 0; i < from && i < len(next); i++ {
		if next[i].Kind == TargetCard || next[i].Kind == TargetPlayer {
			first = false
		}
	}
	for i := from; i < len(next); i++ {
		ref := next[i]
		if ref.Kind != TargetCard && ref.Kind != TargetPlayer {
			continue
		}
		clause := retargetClauseFor(steps, ref)
		if clause == nil {
			// A free-form announcement: nothing can say what another
			// legal target would be, so this target can't be changed
			// and, CR 115.7a, neither can the others.
			return nil
		}
		alt := g.legalTargetsLocked(src, clause)
		// The later slots will change too, so only the picks already
		// made constrain this one (CR 115.7e judges the final set).
		alt = retargetAlternatives(alt, maskedAfter(next, i), i, clause)
		original := item.Targets[i].ID
		budget := changeAllSearchBudget
		fits := func(kind TargetRefKind, id uuid.UUID) bool {
			if id == original {
				// "Only to ANOTHER legal target."
				return false
			}
			trial := append([]TargetRef(nil), next...)
			trial[i] = TargetRef{Kind: kind, ID: id, Mode: ref.Mode, Slot: ref.Slot}
			return g.changeAllCanFinishLocked(src, steps, item.Targets, trial, i+1, &budget)
		}
		var cards, players []uuid.UUID
		for _, id := range alt.Cards {
			if fits(TargetCard, id) {
				cards = append(cards, id)
			}
		}
		for _, id := range alt.Players {
			if fits(TargetPlayer, id) {
				players = append(players, id)
			}
		}
		if len(cards)+len(players) == 0 {
			return nil
		}
		min := 1
		if first && offer.Optional {
			min = 0
		}
		reason := offer.Reason
		if reason == "" {
			reason = "Choose a new target"
			if clause.Label != "" {
				reason = "Choose the new target for " + clause.Label
			}
		}
		g.QueueChoiceForEffect(PendingChoice{
			Kind: PendingChoiceRetarget, Chooser: offer.Chooser, Count: 1,
			Source: offer.Source, Reason: reason, PickTargetPlayers: players,
			PickTargetCards: cards, PickTargetMin: min, PickTargetMax: 1,
			RetargetItem: offer.ItemID, RetargetPolicy: RetargetChangeAll,
			RetargetOptional: offer.Optional, RetargetSlot: i, RetargetReason: offer.Reason,
			RetargetNext: append([]TargetRef(nil), next...),
		})
		return nil
	}
	if err := g.retargetStackItemLocked(offer.ItemID, offer.Chooser, offer.Source, RetargetChangeAll, next); err != nil {
		// The board moved under the walk so that the full set is no
		// longer legal (CR 115.7e): nothing changes.
		if err == ErrIllegalTarget || err == ErrInvalidParam {
			return nil
		}
		return err
	}
	return nil
}

// resolveChangeAllRetargetLocked answers one slot of the walk. No pick
// declines the whole change, which only the first slot of an optional
// change allows (PickTargetMin 0).
//
// Caller must hold g.mu.
func (g *Game) resolveChangeAllRetargetLocked(idx int, choice *PendingChoice, offer RetargetOffer, item *StackItem, targets []TargetRef) error {
	slot := choice.RetargetSlot
	next := append([]TargetRef(nil), choice.RetargetNext...)
	if slot < 0 || slot >= len(next) || len(next) != len(item.Targets) || len(targets) > 1 {
		return ErrInvalidParam
	}
	if len(targets) == 0 {
		if choice.PickTargetMin > 0 {
			return ErrInvalidParam
		}
		g.dequeueChoiceLocked(idx)
		g.runStateChecksLocked()
		return nil
	}
	t := targets[0]
	if !pickOffered(choice, t) {
		return ErrIllegalTarget
	}
	t.Mode, t.Slot = next[slot].Mode, next[slot].Slot
	next[slot] = t
	g.dequeueChoiceLocked(idx)
	if err := g.queueChangeAllRetargetLocked(offer, slot+1, next); err != nil {
		return err
	}
	g.runStateChecksLocked()
	return nil
}

// pickOffered reports whether the answer is one of the prompt's options.
func pickOffered(choice *PendingChoice, t TargetRef) bool {
	switch t.Kind {
	case TargetCard:
		for _, id := range choice.PickTargetCards {
			if id == t.ID {
				return true
			}
		}
	case TargetPlayer:
		for _, id := range choice.PickTargetPlayers {
			if id == t.ID {
				return true
			}
		}
	}
	return false
}

// changeAllSearchBudget bounds the completion search below: a spell with
// many targets on a crowded board would otherwise try every assignment.
// Running out answers "yes", which is safe: the final check at the end
// of the walk still refuses an illegal set, and then nothing changes.
const changeAllSearchBudget = 4096

// maskedAfter is `next` with every real target after `pos` blanked, so
// retargetAlternatives judges a pick only against the ones made before
// it.
func maskedAfter(next []TargetRef, pos int) []TargetRef {
	out := append([]TargetRef(nil), next...)
	for j := pos + 1; j < len(out); j++ {
		if out[j].Kind == TargetCard || out[j].Kind == TargetPlayer {
			out[j] = TargetRef{Kind: TargetNone, Mode: out[j].Mode, Slot: out[j].Slot}
		}
	}
	return out
}

// changeAllCanFinishLocked reports whether the real slots from `from` on
// can each take another legal target so that the whole set passes the
// CR 115.7a check. A depth-first search, bounded by *budget.
//
// Caller must hold g.mu.
func (g *Game) changeAllCanFinishLocked(src TargetSource, steps []AnnouncedClause, original, trial []TargetRef, from int, budget *int) bool {
	if *budget <= 0 {
		return true
	}
	for j := from; j < len(trial); j++ {
		ref := trial[j]
		if ref.Kind != TargetCard && ref.Kind != TargetPlayer {
			continue
		}
		clause := retargetClauseFor(steps, ref)
		if clause == nil {
			return false
		}
		alt := retargetAlternatives(g.legalTargetsLocked(src, clause), maskedAfter(trial, j), j, clause)
		try := func(kind TargetRefKind, id uuid.UUID) bool {
			if id == original[j].ID {
				return false
			}
			next := append([]TargetRef(nil), trial...)
			next[j] = TargetRef{Kind: kind, ID: id, Mode: ref.Mode, Slot: ref.Slot}
			return g.changeAllCanFinishLocked(src, steps, original, next, j+1, budget)
		}
		for _, id := range alt.Cards {
			if try(TargetCard, id) {
				return true
			}
		}
		for _, id := range alt.Players {
			if try(TargetPlayer, id) {
				return true
			}
		}
		return false
	}
	*budget--
	return g.retargetCheckLocked(src, steps, original, trial, RetargetChangeAll) == nil
}
