package game

import "github.com/google/uuid"

// linked_permanent.go — a permanent remembers one permanent object its
// own ability affected, for a linked ability of the same object to name
// later (CR 607.1, CR 607.2e). The Best of Both Worlds' chapter I
// assimilates a creature; chapter III exiles "the Borg assimilated with
// this Saga" and returns it. Card.LinkedPermanent holds the ref.

// SetLinkedPermanentForEffect stores `linked` on the battlefield
// permanent `id` as its LinkedPermanent. A permanent that has left
// stores nothing: it is a new object with no memory (CR 400.7).
//
// Caller must hold g.mu.
func (g *Game) SetLinkedPermanentForEffect(id uuid.UUID, linked ObjectRef) {
	if c := findBattlefieldCard(g, id); c != nil {
		c.LinkedPermanent = linked
	}
}

// LinkedPermanentForEffect is the permanent linked to the source object
// `source` names, read at resolution under CR 608.2h: off the live
// permanent while it is still that object, and off its last-known
// information once it has left. The answer is the linked permanent only
// while it is still that object on the battlefield; nil otherwise.
//
// Caller must hold g.mu.
func (g *Game) LinkedPermanentForEffect(source ObjectRef) *Card {
	info, ok := g.PermanentForEffect(source)
	if !ok {
		return nil
	}
	var linked ObjectRef
	if info.Left {
		if info.LinkedPermanent == nil {
			return nil
		}
		linked = *info.LinkedPermanent
	} else if c := findBattlefieldCard(g, source.ID); c != nil {
		linked = c.LinkedPermanent
	}
	if linked.ID == uuid.Nil {
		return nil
	}
	c := findBattlefieldCard(g, linked.ID)
	if c == nil || c.ObjectEpoch != linked.Epoch {
		return nil
	}
	return c
}
