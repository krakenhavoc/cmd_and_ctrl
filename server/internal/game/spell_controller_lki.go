package game

import "github.com/google/uuid"

// spell_controller_lki.go — "that spell's controller" for an effect that
// names a spell after something may have taken it off the stack (Mindswipe's
// "deals X damage to that spell's controller", S82 OLD batch 351, #5554).
//
// The spell is still on the stack when nothing has moved it (the controller is
// the stack item's), and otherwise CR 608.2h says the effect uses the spell's
// last-known information: its controller as it last existed there, from the
// record every non-resolving exit from the stack files (stack_lki.go). Mindswipe
// counters the spell unless its controller pays {X}, and that spell is gone
// by the time the damage clause reads who to hit.

// LastKnownSpellControllerForEffect is the controller of the spell `spellID`
// while it is on the stack, or as it last existed there when it has left
// without resolving. ok is false when it is neither. Caller must hold g.mu.
func (g *Game) LastKnownSpellControllerForEffect(spellID uuid.UUID) (uuid.UUID, bool) {
	if item := g.StackItemForEffect(spellID); item != nil && item.Controller != uuid.Nil {
		return item.Controller, true
	}
	_, item, ok := g.lastKnownSpellLocked(spellID)
	if !ok || item == nil || item.Controller == uuid.Nil {
		return uuid.Nil, false
	}
	return item.Controller, true
}
