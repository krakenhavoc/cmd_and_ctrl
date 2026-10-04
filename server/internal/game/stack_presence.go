package game

import "github.com/google/uuid"

// StackPresence is a read-only picture of the stack for a seat that
// paces its passes by it (ADR 0119 §2): every item's ID, and the top
// item's ID and controller. The bot runner keeps a first-seen time per
// item and holds a pass on another seat's top item until it has been
// on the stack for the table's bot-speed hold.
//
// It is cheaper than a projection on purpose. The runner reads it on
// every commit it is woken for, including the many on which its seat
// has nothing to decide (#1261), so it copies three things and builds
// no view.
type StackPresence struct {
	// Items lists every item on the stack, in no particular order.
	Items []uuid.UUID
	// Top is the item that resolves next (CR 608.1), and TopController
	// the player who controls it now. Both are uuid.Nil on an empty
	// stack.
	Top           uuid.UUID
	TopController uuid.UUID
}

// StackPresenceSnapshot reads the stack under the read lock. The top
// is chosen the way resolveTopOfStackLocked chooses what resolves: the
// last spell card on the stack zone, unless an ability item with a
// higher insertion Seq sits above it.
func (g *Game) StackPresenceSnapshot() StackPresence {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var out StackPresence
	if len(g.StackMeta) == 0 {
		return out
	}
	out.Items = make([]uuid.UUID, 0, len(g.StackMeta))
	for id, item := range g.StackMeta {
		if item != nil {
			out.Items = append(out.Items, id)
		}
	}
	var top *StackItem
	if g.Stack != nil && len(g.Stack.Cards) > 0 {
		top = g.StackMeta[g.Stack.Cards[len(g.Stack.Cards)-1].InstanceID]
	}
	if ab := g.topAbilityLocked(); ab != nil && (top == nil || ab.Seq > top.Seq) {
		top = ab
	}
	if top != nil {
		out.Top = top.ID
		out.TopController = top.Controller
	}
	return out
}
