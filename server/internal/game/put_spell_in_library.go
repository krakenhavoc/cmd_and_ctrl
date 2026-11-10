package game

import "github.com/google/uuid"

// put_spell_in_library.go — "the owner of target spell … puts it on their
// choice of the top or bottom of their library" (Sudden Setback, #4469).
//
// The spell is not countered (CR 701.6a): the effect moves it off the
// stack to its owner's library, so a spell that can't be countered still
// moves, "whenever a spell is countered" does not see it, and flashback's
// CR 702.34a exile still wins over the chosen destination.

// PutSpellInLibraryKeyedForEffect moves the spell `stackID` from the stack
// to its owner's library at `at` WITHOUT countering it, then runs `then`, a
// registered tail told the spell as TailOutcome.IDs once the move settles
// (and no IDs when it went elsewhere, or had already left the stack). A
// zero tail is "no continuation".
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) PutSpellInLibraryKeyedForEffect(stackID uuid.UUID, at TuckOptions, then KeyedTail) error {
	item, ok := g.StackMeta[stackID]
	if !ok || item == nil || item.Kind != StackItemSpell || g.Stack == nil || !g.Stack.Contains(stackID) {
		return g.runKeyedTailLocked(then, nil, 0)
	}
	r, err := g.spellExitRouteLocked(stackID, &ZoneRef{Kind: ZoneLibrary, Owner: item.Owner}, ZoneGraveyard, false, at)
	if err != nil {
		return err
	}
	if !then.IsZero() {
		r.batchLegs = &routeBatchRun{Leg: stackID, Then: snapshotKeyedTail(then)}
	}
	_, err = g.routeCardToZoneLocked(r)
	return err
}
