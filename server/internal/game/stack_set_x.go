package game

import "github.com/google/uuid"

// SetStackItemXForEffect makes the value of X of a spell or an activated
// ability on the stack `x`: Glava, Five-Advents Mage's "you may have the
// value of X become 5". `id` is the spell card's ID or the ability's
// stack item ID (Event.StackItemID). The spell's mana value on the stack
// (CR 202.3e), the counters it enters with and the X its abilities read
// (CR 107.3m) all follow, since they read item.XValue; nothing is paid or
// refunded (CR 107.3a). It reports whether there was an item to change:
// a spell that was countered in response, or an ability that has
// resolved, has no value left.
//
// Caller must hold g.mu (resolution frame).
func (g *Game) SetStackItemXForEffect(id uuid.UUID, x int) bool {
	if g == nil || g.StackMeta == nil || x < 0 {
		return false
	}
	item := g.StackMeta[id]
	if item == nil {
		return false
	}
	switch item.Kind {
	case StackItemSpell:
		if g.Stack == nil || !g.Stack.Contains(id) {
			return false
		}
	case StackItemActivated:
	default:
		return false
	}
	item.XValue = x
	return true
}
