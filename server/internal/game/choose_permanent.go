package game

import "github.com/google/uuid"

// choose_permanent.go — "As this creature enters, you may choose a
// nonland permanent" (CR 614.12): Scheming Fence (ADR 0155 §8, #3202).
// The seventh member of the as-enters family beside a tribe, a colour, a
// player, a card name, an option and a number. The answer is an OBJECT,
// stored on the permanent as Card.ChosenPermanent.
//
// # No new prompt kind
//
// The question is "pick at most one of these permanents", which is the
// choose-cards prompt with a floor of zero and a ceiling of one: "you
// may" is the empty answer, which that prompt can never refuse. So the
// enumerator, the wire, the client and the live-zone re-check are all
// the existing ones.
//
// # The declared simplification, carried forward
//
// Like every member of the family, the prompt is queued from the
// permanent's AsEnters hook rather than by pausing the CR 614 entry
// pipeline (creature_type_choice.go has the long form). The permanent is
// on the battlefield with nothing chosen until the answer arrives;
// nothing can act in that window, and "nothing chosen" is the safe
// direction: no ability is stopped and none is lent.

// QueueChoosePermanentAsEntersForEffect asks `chooser` to choose up to
// one of `candidates` (permanents on the battlefield) for the permanent
// `source` as it enters, and stores the answer as its ChosenPermanent.
// Returns the choice ID, or uuid.Nil when there was nothing to choose.
//
// A candidate is stored as the OBJECT it was when it was offered: one
// that left and came back before the answer (a new object, CR 400.7) is
// not stored, and nothing is chosen.
//
// Caller must hold g.mu (an AsEnters hook does).
func (g *Game) QueueChoosePermanentAsEntersForEffect(chooser, source uuid.UUID, question string, candidates []uuid.UUID) uuid.UUID {
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
	// #4361 S7a: the answer is a registered continuation, and each
	// candidate's epoch rides the prompt beside it (Carry[i] has
	// CarryNums[i]), so a table waiting on the answer is a restore point.
	return g.QueueChooseCardsForEffect(ChooseCardsPrompt{
		Chooser:   chooser,
		Source:    source,
		Question:  question,
		Cards:     ids,
		Min:       0,
		Max:       1,
		LiveZone:  ZoneBattlefield,
		ThenKey:   choosePermanentAsEntersThen,
		Carry:     append([]uuid.UUID(nil), ids...),
		CarryNums: epochs,
	})
}

// choosePermanentAsEntersThen stores the chosen permanent on the one that
// asked (the prompt's Source) when it is still the object it was offered
// as. An on-disk key: never renamed, never reused.
var choosePermanentAsEntersThen = RegisterChooseCardsThen("choose-cards/as-enters-permanent", func(g *Game, r CardsChosen) error {
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
		g.SetChosenPermanentForEffect(r.Source, ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
		return nil
	}
	return nil
})

// SetChosenPermanentForEffect stores `chosen` on the battlefield
// permanent `id` as its ChosenPermanent, and drops the cached layer
// resolution: the answer is a borrow's lender (Scheming Fence). A
// permanent that has left stores nothing.
//
// Caller must hold g.mu.
func (g *Game) SetChosenPermanentForEffect(id uuid.UUID, chosen ObjectRef) {
	if c := findBattlefieldCard(g, id); c != nil {
		c.ChosenPermanent = chosen
		g.layerVersion.Add(1)
	}
}

// ChosenPermanentOf is the permanent chosen for `source` as it entered,
// when it is still that object on the battlefield: nil when none was
// chosen, when `source` has gone, and when the chosen permanent has left
// the battlefield (a return is a new object, CR 400.7). Read-only; safe
// under either lock.
func (g *Game) ChosenPermanentOf(source *Card) *Card {
	if source == nil || source.ChosenPermanent.ID == uuid.Nil {
		return nil
	}
	c := findBattlefieldCard(g, source.ChosenPermanent.ID)
	if c == nil || c.ObjectEpoch != source.ChosenPermanent.Epoch {
		return nil
	}
	return c
}
