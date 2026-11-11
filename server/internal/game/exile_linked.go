package game

import "github.com/google/uuid"

// exile_linked.go — "you may exile that spell instead of putting it into its
// owner's graveyard as it resolves", linked to the permanent that offered
// it (Forger's Foundry; CR 607.2a). The sibling of exile_with_counter.go and
// exile_and_plot.go: the spell's own resolution exit (resolution_exits.go),
// never for a countered or fizzled spell, competing with the other exits
// under CR 616.1. The card lands in exile "exiled with" the object, and
// that object's other ability reads the cards exiled with it.

// MarkSpellExilesLinkedForEffect marks the spell `spellID` so that it is
// exiled, linked to `link`, instead of going to its owner's graveyard as it
// resolves. Reports false when the ID is not a spell still on the stack, is a
// copy (a copy is not a card, CR 707.10a), or the link names no object.
//
// Caller must hold g.mu.
func (g *Game) MarkSpellExilesLinkedForEffect(spellID uuid.UUID, link PermissionCardRef) bool {
	_, item, ok := g.stolenSpellLocked(spellID)
	if !ok || item.IsCopy || link.ID == uuid.Nil {
		return false
	}
	l := link
	item.ExileLinkedTo = &l
	return true
}

// exileLinkCurrentLocked reports whether `ref` still names the object it
// did when the spell was marked: the card is on the battlefield with the
// same epoch. An object that left (a new one is a different object, CR
// 400.7) is no link, and the spell goes to the graveyard as usual.
func (g *Game) exileLinkCurrentLocked(ref PermissionCardRef) bool {
	c, ok := g.battlefieldCardLocked(ref.ID)
	return ok && c.ObjectEpoch == ref.Epoch
}

// exileLinkRefLocked is the link to the battlefield object `id` as it
// stands, for a prompt that carries IDs only.
func (g *Game) exileLinkRefLocked(id uuid.UUID) PermissionCardRef {
	if c, ok := g.battlefieldCardLocked(id); ok {
		return PermissionCardRef{ID: id, Epoch: c.ObjectEpoch}
	}
	return PermissionCardRef{}
}

// exileLinkNameLocked is the name the choice prompt reads for the exit.
func (g *Game) exileLinkNameLocked(ref PermissionCardRef) string {
	if c, ok := g.battlefieldCardLocked(ref.ID); ok {
		return c.Name
	}
	return "its source"
}
