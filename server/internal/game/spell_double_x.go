package game

import "github.com/google/uuid"

// DoubleSpellXForEffect doubles the value of X of the spell `spellID` while
// it is on the stack (Unbound Flourishing's "double the value of X", CR
// 107.3a): the value announced for it as it was cast, which every reader
// takes from StackItem.XValue (the mana value it has on the stack, CR
// 202.3e, what a resolving permanent "enters with X counters" counts, CR
// 107.3m, the ETB triggers that read X) now reads twice as much. The
// caster pays nothing more: the cost was fixed when the spell was cast.
// An X of 0 stays 0.
//
// False for anything that is not a spell still on the stack (a spell that
// was countered in response, an ability), which has no value left to double.
//
// Caller must hold g.mu (resolution frame).
func (g *Game) DoubleSpellXForEffect(spellID uuid.UUID) bool {
	if g == nil || g.Stack == nil || g.StackMeta == nil || !g.Stack.Contains(spellID) {
		return false
	}
	item := g.StackMeta[spellID]
	if item == nil || item.Kind != StackItemSpell {
		return false
	}
	item.XValue *= 2
	return true
}
