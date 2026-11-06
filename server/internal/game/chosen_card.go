package game

import "github.com/google/uuid"

// chosen_card.go — "Choose a creature card exiled with Koh" and "the last
// chosen card" (Koh, the Face Stealer; ADR 0155's 2026-11-05 amendment).
// The answer is an OBJECT, stored on the permanent whose ability chose it
// as Card.ChosenCard: plain data, carried by the snapshot, read by the
// permanent's borrow on every layer pass.
//
// Unlike ChosenPermanent this is not an as-enters choice. It is made each
// time the ability resolves, and each choice replaces the last ("the last
// chosen card"). A choice the chooser never makes (they left the game)
// leaves the last one standing.

// QueueChooseExiledCardForEffect asks `chooser` to choose exactly one of
// `candidates` (cards in exile) for the permanent `source`, and stores the
// answer as its ChosenCard. One candidate is chosen at once, with no
// prompt: there is nothing to decide. Returns false when there was
// nothing to choose, which leaves the last choice standing.
//
// A candidate is stored as the OBJECT it was when it was offered: one that
// left exile before the answer (a new object, CR 400.7) is not stored.
//
// Caller must hold g.mu in write mode (a resolution frame).
func (g *Game) QueueChooseExiledCardForEffect(chooser, source uuid.UUID, question string, candidates []*Card) bool {
	ids := make([]uuid.UUID, 0, len(candidates))
	epochs := make([]int, 0, len(candidates))
	seen := make(map[uuid.UUID]bool, len(candidates))
	for _, c := range candidates {
		if c == nil || seen[c.InstanceID] {
			continue
		}
		seen[c.InstanceID] = true
		ids = append(ids, c.InstanceID)
		epochs = append(epochs, c.ObjectEpoch)
	}
	switch len(ids) {
	case 0:
		return false
	case 1:
		g.SetChosenCardForEffect(source, ObjectRef{ID: ids[0], Epoch: epochs[0]})
		return true
	}
	g.QueueChooseCardsForEffect(ChooseCardsPrompt{
		Chooser:   chooser,
		Source:    source,
		Question:  question,
		Cards:     ids,
		Min:       1,
		Max:       1,
		LiveZone:  ZoneExile,
		ThenKey:   chooseExiledCardThen,
		Carry:     append([]uuid.UUID(nil), ids...),
		CarryNums: epochs,
	})
	return true
}

// chooseExiledCardThen stores the chosen card on the permanent that asked
// (the prompt's Source) when the card is still in exile as the object it
// was offered as. An on-disk key: never renamed, never reused.
var chooseExiledCardThen = RegisterChooseCardsThen("choose-cards/last-chosen-exiled-card", func(g *Game, r CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	picked := r.Picked[0]
	for i, id := range r.Carry {
		if id != picked || i >= len(r.Nums) {
			continue
		}
		if c := g.exiledObjectLocked(ObjectRef{ID: picked, Epoch: r.Nums[i]}); c != nil {
			g.SetChosenCardForEffect(r.Source, ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
		}
		return nil
	}
	return nil
})

// SetChosenCardForEffect stores `chosen` on the battlefield permanent `id`
// as its ChosenCard, replacing the last, and drops the cached layer
// resolution: the answer is a borrow's lender. A permanent that has left
// stores nothing.
//
// Caller must hold g.mu.
func (g *Game) SetChosenCardForEffect(id uuid.UUID, chosen ObjectRef) {
	if c := findBattlefieldCard(g, id); c != nil {
		c.ChosenCard = chosen
		g.layerVersion.Add(1)
	}
}

// ChosenExiledCardOf is the last card chosen for `source`, while it is
// still in exile as the object it was when chosen: nil when none was
// chosen, and when the chosen card has left exile (a return is a new
// object nobody chose, CR 400.7). A pointer into the exile zone, for a
// reader inside a layer pass. Read-only; safe under either lock.
func (g *Game) ChosenExiledCardOf(source *Card) *Card {
	if source == nil || source.ChosenCard.ID == uuid.Nil {
		return nil
	}
	return g.exiledObjectLocked(source.ChosenCard)
}

// exiledObjectLocked is the card in exile that is still the object `ref`
// names, or nil.
func (g *Game) exiledObjectLocked(ref ObjectRef) *Card {
	if g.Exile == nil || ref.ID == uuid.Nil {
		return nil
	}
	for i := range g.Exile.Cards {
		c := &g.Exile.Cards[i]
		if c.InstanceID == ref.ID {
			if c.ObjectEpoch != ref.Epoch {
				return nil
			}
			return c
		}
	}
	return nil
}
