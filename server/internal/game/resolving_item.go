package game

import "github.com/google/uuid"

// resolving_item.go — the stack item whose resolution is in flight
// (#920, CR 707.10).
//
// # The bug this closes
//
// resolveTopOfStackLocked deletes an item's StackMeta entry before it
// calls OnResolve, and resolveTopAbilityLocked does the same. That is
// correct as far as the STACK is concerned: the object has left it,
// nothing may target it, and nothing may respond. But CR 707.10 lets a
// resolving spell create a copy of ITSELF — "that player may copy this
// spell and may choose a new target for that copy" is the whole Chain
// cycle — and CopySpellForEffect finds its source card and its
// announce-time choices through StackMeta. With the entry already
// gone, the copy leg found nothing and did nothing.
//
// # One slot, and its lifetime
//
// Game.resolving holds the item and a VALUE copy of the card the stack
// held, because by the time the copy is actually made neither is
// reachable any more:
//
//   - the meta is out of StackMeta (this file's whole subject);
//   - the CARD is out of the Stack zone, because the copy decision is
//     a prompt. Chain of Vapor's "may sacrifice a land" pauses
//     resolution, resolveTopOfStackLocked carries on and routes the
//     spell to its owner's graveyard, and the answer arrives later.
//     What the copy is made from is therefore last-known information
//     (CR 608.2n — the spell is put into the graveyard as the final
//     step of its own resolution, so in the rules it is still on the
//     stack when the copy is created; here it is not, and LKI is how
//     that difference is spelled).
//
// The slot is SET when an item begins to resolve and CLEARED by
// beginEventBatchLocked — which runs at exactly the points where play
// moves on (#829, CR 603.2c; #1341 added a third): the next
// resolution, the cursor entering a step, and a special action being
// taken. That is the terminal outcome, and it is the right one for a
// pause: a resolution-time prompt BLOCKS the table (choice_gate.go),
// so the cursor cannot walk past an open copy decision, nothing else
// can resolve under it, and no special action can be taken under it
// either — so clearing the slot at that third point never races a
// paused continuation. The slot therefore survives any number of
// paused continuations belonging to the resolution that opened it,
// and not one event past it.
//
// Nothing else reads it. stackSpellLocked consults it only when the
// requested ID is not on the stack AND is this item's own — so an
// ordinary Reverberate pointed at a spell that has already left still
// gets ErrCardNotFound, which is what its callers are written for.

// resolvingItem is the meta and the card of the item currently
// resolving. Replaced wholesale at each event batch and never mutated
// in place, which is why the clone can share the pointer.
type resolvingItem struct {
	// item is the live StackItem, the same pointer the resolution and
	// every continuation it queued are holding.
	item *StackItem

	// card is the spell as the stack held it (CR 608.2n LKI), and
	// hasCard says whether there was one: an ability item has no card
	// on the stack at all.
	card    Card
	hasCard bool

	// sourceEpoch is the Card.ObjectEpoch the item's SOURCE card had
	// as the resolution began, or -1 when it was in no zone (#1432).
	// A source whose epoch has changed since was moved by this
	// resolution's own effect — nothing else can move a card while an
	// item resolves — and CR 400.7's exception lets the rest of that
	// effect find the object it moved. Read by
	// resolvingSourceEpochLocked; recorded for ability items only.
	sourceEpoch int
}

// beginResolvingLocked parks an item in the resolving slot. Caller
// must hold g.mu.
func (g *Game) beginResolvingLocked(item *StackItem) {
	if item == nil {
		g.resolving = nil
		return
	}
	g.resolving = &resolvingItem{item: item, sourceEpoch: -1}
	if item.SourceCardID != uuid.Nil {
		g.resolving.sourceEpoch = g.cardObjectEpochLocked(item.SourceCardID)
	}
}

// resolvingSourceEpochLocked is the epoch `item`'s source card had
// when `item` began to resolve (#1432), and false when `item` is not
// the item the resolving slot holds — nothing resolving, a spell, or
// a different item. Matched by ID as well as pointer, because Clone
// shares the slot and an undo's item is a copy. Caller must hold g.mu.
func (g *Game) resolvingSourceEpochLocked(item *StackItem) (int, bool) {
	r := g.resolving
	if r == nil || r.hasCard || r.item == nil || item == nil {
		return 0, false
	}
	if r.item != item && r.item.ID != item.ID {
		return 0, false
	}
	return r.sourceEpoch, true
}

// beginResolvingSpellLocked parks a spell item together with the card
// the stack held, so a copy made after the spell has been routed away
// still has its copiable values. Caller must hold g.mu.
func (g *Game) beginResolvingSpellLocked(item *StackItem, card Card) {
	if item == nil {
		g.resolving = nil
		return
	}
	g.resolving = &resolvingItem{item: item, card: card, hasCard: true}
}

// resolvingSpellLocked returns the card and item of the spell
// currently resolving when `spellID` names it, and false otherwise.
//
// The ID check is what keeps the slot from widening any other lookup:
// a caller that did not ask for the resolving spell by name never
// reaches it. Caller must hold g.mu.
func (g *Game) resolvingSpellLocked(spellID uuid.UUID) (Card, *StackItem, bool) {
	r := g.resolving
	if r == nil || !r.hasCard || r.item == nil {
		return Card{}, nil, false
	}
	if r.item.ID != spellID {
		return Card{}, nil, false
	}
	return r.card, r.item, true
}
