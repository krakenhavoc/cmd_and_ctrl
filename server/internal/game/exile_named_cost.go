package game

import (
	"slices"

	"github.com/google/uuid"
)

// exile_named_cost.go — AbilityCost.ExileNamed (#5683): "Exile … and
// artifacts you control named Eye of Vecna and Hand of Vecna" (The Book
// of Vile Darkness), paid beside ExileSelf.
//
// The other exile components name a count the activator picks
// (ExilePermanentsCost) or one fixed object (ExileSelf, ExileGrantor).
// This one names a SET the board decides: every permanent the activator
// controls that has the card type and one of the names, the source left
// out because the cost's own "Exile <this>" pays for it. There is
// nothing to choose (CR 601.2b, through CR 602.2b), so nothing is
// announced: the set is read once, as the cost is determined (CR
// 601.2f), and all of it is exiled as the cost is paid (CR 601.2h). The
// set may be empty; the clause then exiles nothing and the rest of the
// cost is still payable, which is how the Book makes a Vecna with fewer
// than three cards.
//
// The permanents it moved are recorded on PaidCost.Exiled, after any
// other exile component's, so the resolving ability finds "the exiled
// cards" by instance ID in exile, as it finds ExilePermanentsCost's. A
// token is exiled like any permanent and then ceases to exist (CR
// 111.7), so it is in the record and never among the cards.
//
// Plain data: a card type and names, so the cost survives Clone and a
// snapshot and adds no func-typed route from Game (ADR 0041).

// ExileNamedCost is the "and <card type>s you control named A and B"
// component of an activated ability's cost. Nil, and a value with no
// name, demand nothing.
type ExileNamedCost struct {
	// CardType is the card type the permanent must have, lowercase
	// ("artifact"), read off its current types. Empty is any permanent.
	CardType string
	// Names are the names a permanent may have; it is exiled when its
	// current name is one of them (CR 201.2a: the same name).
	Names []string
}

// Empty reports whether the component demands nothing. Nil-safe.
func (ec *ExileNamedCost) Empty() bool { return ec == nil || len(ec.Names) == 0 }

// exileNamedSetLocked is the set an ExileNamed component exiles right
// now: every permanent `playerID` controls, other than the source, with
// the component's card type and one of its names, in battlefield order.
// Caller must hold g.mu.
func (g *Game) exileNamedSetLocked(playerID, sourceID uuid.UUID, ec *ExileNamedCost) []uuid.UUID {
	if ec.Empty() || g.Battlefield == nil {
		return nil
	}
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID == sourceID || c.Controller != playerID {
			continue
		}
		if ec.CardType != "" && !hasCardType(c, ec.CardType) {
			continue
		}
		if slices.Contains(ec.Names, currentNameOf(c)) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// currentNameOf is a permanent's name after the layers (a copy effect
// writes the copied name into the printed field; a later name change is
// on the layered characteristic).
func currentNameOf(c *Card) string {
	if c.effective != nil && c.effective.Name != "" {
		return c.effective.Name
	}
	return c.Name
}

// ExileNamedSetForEffect is the set an ExileNamed component of `source`'s
// cost would exile if `playerID` paid it now: for the view and tests.
func (g *Game) ExileNamedSetForEffect(playerID, sourceID uuid.UUID, ec *ExileNamedCost) []uuid.UUID {
	return g.exileNamedSetLocked(playerID, sourceID, ec)
}

// validateExileNamedCostLocked checks the set against the other
// components of the same payment: one permanent pays one component (CR
// 118.3), so a permanent the clause names may not also be sacrificed,
// returned or exiled by another. Caller must hold g.mu.
func validateExileNamedCostLocked(named []uuid.UUID, moved ...[]uuid.UUID) error {
	for _, id := range named {
		for _, ids := range moved {
			if slices.Contains(ids, id) {
				return ErrInvalidParam
			}
		}
	}
	return nil
}
