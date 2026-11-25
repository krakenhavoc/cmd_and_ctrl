package game

// retarget_change_all.go — #6228: "change the target or targets" of an
// item with more than one target (CR 115.7a, Psychic Battle). "Each
// target can be changed only to another legal target. If a target can't
// be changed to another legal target, the original target is unchanged
// … If all the targets aren't changed to other legal targets, none of
// them are changed." CR 115.7e: only the final set is evaluated.
//
// The walk is the staged one related targets use
// (queueRelatedRetargetLocked): each slot is asked in turn against the
// list as staged so far (PendingChoice.RetargetNext), and the item's
// targets are rewritten once, when the last slot is answered.

// queueChangeAllRetargetLocked asks for slot `from`'s new target, or
// commits `next` when every slot has been answered. A slot with no other
// legal target ends the walk with nothing changed (CR 115.7a). A slot of
// an optional change may be declined, which changes nothing either
// (ResolveRetarget).
//
// Caller must hold g.mu.
func (g *Game) queueChangeAllRetargetLocked(offer RetargetOffer, from int, next []TargetRef) error {
	item := g.StackMeta[offer.ItemID]
	if item == nil {
		return nil
	}
	steps, src := g.itemAnnouncedClauses(item), g.retargetSourceLocked(item)
	for i := from; i < len(next); i++ {
		ref := next[i]
		if ref.Kind != TargetCard && ref.Kind != TargetPlayer {
			continue
		}
		clause := retargetClauseFor(steps, ref)
		if clause == nil {
			// A free-form announcement: nothing says what another
			// legal target would be, so this slot can't be changed.
			return nil
		}
		alt := g.legalTargetsLocked(src, clause)
		alt = retargetAlternatives(alt, next, i, clause)
		fits := func(t TargetRef) bool {
			trial := append([]TargetRef(nil), next...)
			trial[i] = t
			return g.targetsAfterCanFinishLocked(src, steps, item.Targets, trial, i)
		}
		var cards, players = alt.Cards[:0:0], alt.Players[:0:0]
		for _, id := range alt.Cards {
			if fits(TargetRef{Kind: TargetCard, ID: id, Mode: ref.Mode, Slot: ref.Slot}) {
				cards = append(cards, id)
			}
		}
		for _, id := range alt.Players {
			if fits(TargetRef{Kind: TargetPlayer, ID: id, Mode: ref.Mode, Slot: ref.Slot}) {
				players = append(players, id)
			}
		}
		if len(cards)+len(players) == 0 {
			// CR 115.7a: this one can't be changed, so none is.
			return nil
		}
		min := 1
		if offer.Optional {
			min = 0
		}
		reason := offer.Reason
		if reason == "" {
			reason = "Choose the new target for " + clause.Label
		}
		g.QueueChoiceForEffect(PendingChoice{
			Kind: PendingChoiceRetarget, Chooser: offer.Chooser, Count: 1,
			Source: offer.Source, Reason: reason, PickTargetPlayers: players,
			PickTargetCards: cards, PickTargetMin: min, PickTargetMax: 1,
			RetargetItem: offer.ItemID, RetargetPolicy: offer.Policy,
			RetargetOptional: offer.Optional, RetargetSlot: i, RetargetReason: offer.Reason,
			RetargetNext: append([]TargetRef(nil), next...),
		})
		return nil
	}
	return g.retargetStackItemLocked(offer.ItemID, offer.Chooser, offer.Policy, next, offer.Source)
}
