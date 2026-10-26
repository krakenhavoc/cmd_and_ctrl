package game

import "github.com/google/uuid"

// asked_item.go — the stack item that asked a keyed prompt, as plain data
// (#4264).
//
// A keyed prompt (a pay-unless, an any-player offer, an upkeep pay-or-else)
// carries no closure, and the item that asked it has long resolved by the
// time the prompt is answered. A closure used to hold that item's pointer,
// so a branch could read the trigger's event, the targets, the X and the
// object the ability came from. The keyed form hands a branch only the
// payer, the card and some carried IDs, which leaves the Context a branch
// runs against with a stand-in that has none of the rest: an ability whose
// source left and came back (CR 400.7) read as the same permanent, and a
// branch that reads the trigger's event or the item's targets had nowhere
// to get them.
//
// AskedItem is the part of the item an answer reads, as data the prompt
// carries (PendingChoice.Asked), written to a restore point and cloned by
// an undo. StackItem builds the item a branch runs against. It is a copy
// of the item as it was when the prompt was queued; the item's Effect and
// its target and mode clauses are not part of it, because the effect that
// asked has run and nothing re-resolves it.

// AskedItem is the data of the stack item that asked a keyed prompt.
type AskedItem struct {
	// ID is the asking item's ID: the key a branch's log lines and a
	// reflexive trigger it queues name it by.
	ID   uuid.UUID     `json:"id"`
	Kind StackItemKind `json:"kind"`
	// Controller and Owner are the item's, SourceCardID the card it came
	// from.
	Controller   uuid.UUID `json:"controller"`
	Owner        uuid.UUID `json:"owner"`
	SourceCardID uuid.UUID `json:"sourceCardId"`
	// SourceObject is the object the ability came from (CR 400.7): the
	// card and the epoch it had when the ability triggered or was
	// activated. A pointer so an unstamped item (a spell) writes nothing.
	SourceObject    *ObjectRef `json:"sourceObject,omitempty"`
	SourceTurnOvers int        `json:"sourceTurnOvers,omitempty"`
	SourceEpoch     int        `json:"sourceEpoch,omitempty"`
	Label           string     `json:"label,omitempty"`
	// Targets are the item's chosen targets, Trigger the triggering
	// event, XValue the X the spell was cast with.
	Targets []TargetRef     `json:"targets,omitempty"`
	Trigger *TriggerContext `json:"trigger,omitempty"`
	XValue  int             `json:"xValue,omitempty"`
	// Modes are the modes a modal spell was cast with, which a target
	// re-check reads.
	Modes []int `json:"modes,omitempty"`
	// Payload is what the item recorded as it resolved (#4253 slice 10b):
	// the players a ChoosePlayer chose (ChosenPlayerOn, ChosenPlayersOn)
	// and the cards a clause filed. A keyed answer that asks again reads
	// the answers before it, as Gluntch, the Bestower's "a second player"
	// does.
	Payload []TargetRef `json:"payload,omitempty"`
}

// NewAskedItem is the data of `item` that a keyed prompt carries. Nil for
// a nil item.
func NewAskedItem(item *StackItem) *AskedItem {
	if item == nil {
		return nil
	}
	return &AskedItem{
		ID:              item.ID,
		Kind:            item.Kind,
		Controller:      item.Controller,
		Owner:           item.Owner,
		SourceCardID:    item.SourceCardID,
		SourceObject:    item.SourceObject.stamped(),
		SourceTurnOvers: item.SourceTurnOvers,
		SourceEpoch:     item.SourceEpoch,
		Label:           item.Label,
		Targets:         copyTargetRefs(item.Targets),
		Trigger:         cloneTriggerContext(item.Trigger),
		XValue:          item.XValue,
		Modes:           copyInts(item.Modes),
		Payload:         copyTargetRefs(item.Payload),
	}
}

// StackItem is the item a keyed branch runs against: a fresh copy each
// call, so a branch that edits what it is handed edits nothing the prompt
// carries. Nil for nil.
func (a *AskedItem) StackItem() *StackItem {
	if a == nil {
		return nil
	}
	return &StackItem{
		ID:              a.ID,
		Kind:            a.Kind,
		Controller:      a.Controller,
		Owner:           a.Owner,
		SourceCardID:    a.SourceCardID,
		SourceObject:    a.SourceObject.value(),
		SourceTurnOvers: a.SourceTurnOvers,
		SourceEpoch:     a.SourceEpoch,
		Label:           a.Label,
		Targets:         copyTargetRefs(a.Targets),
		Trigger:         cloneTriggerContext(a.Trigger),
		XValue:          a.XValue,
		Modes:           copyInts(a.Modes),
		Payload:         copyTargetRefs(a.Payload),
	}
}

// copyAskedItem is a deep copy, nil for nil.
func copyAskedItem(a *AskedItem) *AskedItem {
	if a == nil {
		return nil
	}
	out := *a
	if a.SourceObject != nil {
		o := *a.SourceObject
		out.SourceObject = &o
	}
	out.Targets = copyTargetRefs(a.Targets)
	out.Trigger = cloneTriggerContext(a.Trigger)
	out.Modes = copyInts(a.Modes)
	out.Payload = copyTargetRefs(a.Payload)
	return &out
}
