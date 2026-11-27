package game

import "github.com/google/uuid"

// resolving_spell_returns_to_hand.go — S82 OLD b411 (Unexpected Results):
// "If it's a land card, you may put it onto the battlefield and return
// Unexpected Results to its owner's hand."
//
// The spell is returned to hand as part of its own resolution, so it is not
// put into its owner's graveyard as the final step (CR 608.2n). That is
// Soulfire Grand Master's exit (StackItem.HandInsteadOfGraveyard, buyback's
// route, resolution_exits.go) decided by the spell's own effect instead of an
// earlier promise: the effect marks the resolving item, and the exit applies
// when the resolution finishes — after the answer, when the effect paused on
// a prompt of its own (resolution_tail.go).

// MarkResolvingSpellReturnsToHandForEffect marks the spell `spellID`, which is
// resolving now, so it goes to its owner's hand instead of its graveyard as
// its resolution ends. Reports false when `spellID` is not the spell resolving
// (or parked behind its own prompt), or when it is a copy, which ceases to
// exist (CR 707.10a, CR 608.2n) rather than going anywhere.
//
// Caller must hold g.mu.
func (g *Game) MarkResolvingSpellReturnsToHandForEffect(spellID uuid.UUID) bool {
	marked := false
	if t := g.pausedResolutionTail; t != nil && t.item != nil && t.item.ID == spellID && t.isSpell() && !t.item.IsCopy {
		t.item.HandInsteadOfGraveyard = true
		marked = true
	}
	if _, item, ok := g.resolvingSpellLocked(spellID); ok && !item.IsCopy {
		item.HandInsteadOfGraveyard = true
		marked = true
	}
	return marked
}
