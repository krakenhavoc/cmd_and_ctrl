package game

import "github.com/google/uuid"

// Related targets are committed together: changing the anchor may require
// changing a later target too. The stack item keeps its original targets
// throughout the picker, including across an undo.
func (g *Game) queueRelatedRetargetLocked(offer RetargetOffer, from int, next []TargetRef) error {
	item := g.StackMeta[offer.ItemID]
	if item == nil {
		return nil
	}
	steps, src := g.itemAnnouncedClauses(item), g.retargetSourceLocked(item)
	for i := from; i < len(next); i++ {
		ref := next[i]
		stepIndex := -1
		for j, step := range steps {
			if step.Mode == ref.Mode && step.Slot == ref.Slot {
				stepIndex = j
				break
			}
		}
		if stepIndex < 0 {
			continue
		}
		clause := steps[stepIndex].Clause
		alt := g.legalTargetsLocked(src, &clause)
		alt = retargetAlternatives(alt, next, i, &clause)
		fits := func(kind TargetRefKind, id uuid.UUID) bool {
			trial := append([]TargetRef(nil), next...)
			trial[i] = TargetRef{Kind: kind, ID: id, Mode: ref.Mode, Slot: ref.Slot}
			return g.copyTargetsCanFinishLocked(src, steps, item.Targets, trial, stepIndex+1)
		}
		cards := alt.Cards[:0]
		for _, id := range alt.Cards {
			if fits(TargetCard, id) {
				cards = append(cards, id)
			}
		}
		alt.Cards = cards
		players := alt.Players[:0]
		for _, id := range alt.Players {
			if fits(TargetPlayer, id) {
				players = append(players, id)
			}
		}
		alt.Players = players
		canKeep := g.copyTargetsCanFinishLocked(src, steps, item.Targets, next, stepIndex+1)
		if len(cards)+len(players) == 0 {
			if !canKeep {
				return ErrIllegalTarget
			}
			continue
		}
		min := 1
		if canKeep {
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
	return g.RetargetStackItemForEffect(offer.ItemID, offer.Chooser, offer.Policy, next)
}
