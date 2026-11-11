package game

import "github.com/google/uuid"

// exile_spell_linked.go — "exile target spell", linked to the source
// (CR 607.2a): Smirking Spelljacker's "When this creature enters, exile
// target spell an opponent controls. Whenever this creature attacks, if a
// card is exiled with it, you may cast the exiled card without paying its
// mana cost."
//
// ExileSpellKeyedForEffect on the spell's exit route with the source OBJECT
// stamped on it, the way ExileLinkedForEffect stamps a card's: once the
// spell has landed in exile its Card.ExiledWith names `source`
// ({ID, Epoch}), so a source that left and came back is a new object with no
// claim on the card (CR 400.7). A spell that has already left the stack, or
// that went to the command zone instead (CR 903.9), is not linked, and
// `then` is told it did not reach exile.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) ExileSpellLinkedKeyedForEffect(source ObjectRef, stackID uuid.UUID, then KeyedTail) error {
	if item, ok := g.StackMeta[stackID]; !ok || item == nil || item.Kind != StackItemSpell ||
		g.Stack == nil || !g.Stack.Contains(stackID) {
		return g.runKeyedTailLocked(then, nil, 0)
	}
	r, err := g.spellExitRouteLocked(stackID, &ZoneRef{Kind: ZoneExile}, ZoneExile, false, TuckOptions{})
	if err != nil {
		return err
	}
	r.ExiledWith = PermissionCardRef(source)
	if !then.IsZero() {
		r.batchLegs = &routeBatchRun{Leg: stackID, Then: snapshotKeyedTail(then)}
	}
	_, err = g.routeCardToZoneLocked(r)
	return err
}
