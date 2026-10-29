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
//   - the CARD has no meta behind it, because the copy decision is a
//     prompt. Chain of Vapor's "may sacrifice a land" pauses
//     resolution, and the answer arrives later. Since #3613 the card
//     waits in the Stack zone for that answer (CR 608.2n — the spell
//     is put into the graveyard as the final step of its own
//     resolution; resolution_tail.go), but its StackMeta entry is
//     gone, so stackSpellLocked's stack scan cannot pair the two and
//     falls back here. The copy is made from the card as it was when
//     the resolution began, which is the same card.
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
// # Across a restore (#4270)
//
// The snapshot carries the slot (resolvingSnapshot, below). It used to
// drop it, on the premise that a prompt paused inside a resolution always
// holds a resume frame and so no snapshot is a restore point while the
// slot is set. A KEYED prompt (ADR 0127: a pay-unless, a confirm, a
// ThenKey choice) holds none, so a table waiting on one is a restore
// point whose answer reads the slot: Wandering Archaic's "copy that
// spell?" finds the Bolt that already resolved only here. A restored
// slot has the same lifetime as a live one, because eventBatch is
// carried and beginEventBatchLocked clears the slot at the next batch.
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
	g.beginPromptKeysForItemLocked(item)
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
	g.beginPromptKeysForItemLocked(item)
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

// ResolvingAbilityTriggerForEffect is the triggering event of the
// ABILITY resolving now, when itemID names it: the EventSagaChapter a
// chapter ability was put on the stack for, with its chapter number in
// Amount. It is read by a trigger on that ability's EventResolve, whose
// item has already left StackMeta (Tom Bombadil's "whenever the final
// chapter ability of a Saga you control resolves", #4505, CR 714.2e).
//
// False for a spell, for an item that is not the resolving one, and for
// an ability with no triggering event (an activated ability, a
// reflexive trigger). A copy of the context, so the reader cannot
// change the item. Caller must hold g.mu.
func (g *Game) ResolvingAbilityTriggerForEffect(itemID uuid.UUID) (TriggerContext, bool) {
	r := g.resolving
	if r == nil || r.hasCard || r.item == nil || r.item.ID != itemID || r.item.Trigger == nil {
		return TriggerContext{}, false
	}
	return *cloneTriggerContext(r.item.Trigger), true
}

// resolvingSnapshot is the resolving slot as plain data (#4270). It
// exists because a keyed prompt (ADR 0127: a pay-unless, a confirm, a
// choice with a ThenKey) carries no resume frame, so a table waiting on
// one is a restore point, and the answer's continuation can read this
// slot. Wandering Archaic's "copy that spell?" found it: the cast Bolt
// had already resolved when the question was answered, so it is
// reachable only through the slot, and a restore that dropped the slot
// answered "yes" with nothing to copy.
//
// The slot is data, not behaviour. Almost every reader takes a field off
// the item or the card (its controller, its targets, its label, its
// source object, its catalog ref). The exception is a copy of the
// resolving ability itself (stackAbilityLocked), which carries its
// Effect. restoreStackItem re-derives the Effect and the target clause
// when the item names a catalog row and leaves them nil when it does
// not, as it does for a spell in lastKnownStack, and snapshotResolving
// counts an item it could not rebuild while the resolution is open
// (#4361), so such a table is never a restore point.
type resolvingSnapshot struct {
	Item stackItemSnapshot `json:"item"`
	// Card is the spell as the stack held it (CR 608.2n last-known
	// information). Nil for an ability item, which has no card on the
	// stack at all.
	Card *cardSnapshot `json:"card,omitempty"`
	// SourceEpoch is resolvingItem.sourceEpoch: -1 when the source was
	// in no zone or the item is a spell.
	SourceEpoch int `json:"sourceEpoch"`
}

// snapshotResolving is the capture half; nil for an empty slot. Caller
// must hold g.mu.
//
// What the item and the card would cost a restore point depends on
// whether the resolution is still open (#4361):
//
//   - OPEN (resolutionOpen: a prompt paused inside it). The rest of the
//     resolution can still read the item as an object, not just as data:
//     copying the resolving ability (stackAbilityLocked, CR 707.10)
//     carries its Effect, and copying the resolving spell carries the
//     card's abilities. So the item and the card are counted in the live
//     census exactly as a stack item and a last-known spell are: an
//     unkeyed Effect, a clause restore cannot re-derive, or a card whose
//     instance abilities cannot be re-derived is
//     census:IntrinsicAbilityCards, and the table is not a restore point.
//     Counting them in a scratch census instead let such a table restore
//     with the Effect silently gone.
//   - CLOSED. The slot is spent: the item's effect has run, nothing left
//     reads more than its fields, and the slot is cleared at the next
//     event batch. A scratch census takes the item and the card, so a
//     table that has finished a resolution stays a restore point.
func snapshotResolving(g *Game, cen *ContinuationCensus) *resolvingSnapshot {
	r := g.resolving
	if r == nil || r.item == nil {
		return nil
	}
	var spent ContinuationCensus
	count := &spent
	if g.resolutionOpen {
		count = cen
	}
	oracleID := oracleIDOfLocked(g, r.item.SourceCardID)
	if r.hasCard {
		oracleID = r.card.OracleID
	}
	out := &resolvingSnapshot{
		Item:        snapshotStackItemAs(r.item, oracleID, count),
		SourceEpoch: r.sourceEpoch,
	}
	if r.hasCard {
		c := snapshotCard(r.card, count)
		out.Card = &c
	}
	return out
}

// restoreResolving is the restore half. The batch counter restores with
// the slot, so the next beginEventBatchLocked clears a restored slot at
// the same point it would have cleared a live one.
func restoreResolving(in *resolvingSnapshot) *resolvingItem {
	if in == nil {
		return nil
	}
	item, _ := restoreStackItem(&in.Item)
	r := &resolvingItem{item: item, sourceEpoch: in.SourceEpoch}
	if in.Card != nil {
		r.card = restoreCard(in.Card)
		r.hasCard = true
	}
	return r
}
