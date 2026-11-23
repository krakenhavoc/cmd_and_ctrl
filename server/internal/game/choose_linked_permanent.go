package game

import "github.com/google/uuid"

// choose_linked_permanent.go — "As this creature enters, choose another
// creature you control" whose answer a linked ability names later (CR 607.2e):
// Tyrannical Pitlord's "the chosen creature gets +3/+3 and has flying" and
// "when this creature leaves the battlefield, sacrifice the chosen creature".
//
// The choice is mandatory when there is a candidate (Min 1), unlike
// QueueChoosePermanentAsEntersForEffect's "you may choose". It is stored as
// the permanent's LinkedPermanent, an object with its CR 400.7 epoch, which
// PermanentInfo keeps as last-known information, so the leave trigger still
// names the creature once the source has gone (CR 608.2h).

// QueueChooseLinkedPermanentAsEntersForEffect asks `chooser` to pick one of
// `candidates` (battlefield permanents other than `source`) as `source`
// enters, and stores the answer as its LinkedPermanent. Returns the choice ID,
// or uuid.Nil when there was nothing to choose.
//
// Caller must hold g.mu (an AsEnters hook does).
func (g *Game) QueueChooseLinkedPermanentAsEntersForEffect(chooser, source uuid.UUID, question string, candidates []uuid.UUID) uuid.UUID {
	seen := make(map[uuid.UUID]bool, len(candidates))
	ids := make([]uuid.UUID, 0, len(candidates))
	epochs := make([]int, 0, len(candidates))
	for _, id := range candidates {
		c := findBattlefieldCard(g, id)
		if c == nil || id == source || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
		epochs = append(epochs, c.ObjectEpoch)
	}
	if len(ids) == 0 {
		return uuid.Nil
	}
	return g.QueueChooseCardsForEffect(ChooseCardsPrompt{
		Chooser:   chooser,
		Source:    source,
		Question:  question,
		Cards:     ids,
		Min:       1,
		Max:       1,
		LiveZone:  ZoneBattlefield,
		ThenKey:   chooseLinkedPermanentAsEntersThen,
		Carry:     append([]uuid.UUID(nil), ids...),
		CarryNums: epochs,
	})
}

// chooseLinkedPermanentAsEntersThen stores the pick on the permanent that
// asked as its LinkedPermanent, when the pick is still the object it was
// offered as. An on-disk key: never renamed, never reused.
var chooseLinkedPermanentAsEntersThen = RegisterChooseCardsThen("choose-cards/as-enters-linked-permanent", func(g *Game, r CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	picked := r.Picked[0]
	for i, id := range r.Carry {
		if id != picked || i >= len(r.Nums) {
			continue
		}
		c := findBattlefieldCard(g, picked)
		if c == nil || c.ObjectEpoch != r.Nums[i] {
			return nil
		}
		g.SetLinkedPermanentForEffect(r.Source, ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
		g.layerVersion.Add(1)
		return nil
	}
	return nil
})
