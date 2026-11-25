package game

import (
	"github.com/google/uuid"
)

// delayed_recast.go — #6221: Ertai's Meddling.
//
//	X can't be 0. / Target spell's controller exiles it with X delay
//	counters on it. / At the beginning of each of that player's upkeeps,
//	if that card is exiled, remove a delay counter from it. If the card
//	has no delay counters on it, the player puts it onto the stack as a
//	copy of the original spell.
//
// CR 707.10: "A copy of a spell or ability copies both the characteristics
// of the spell or ability and all decisions made for it, including modes,
// targets, the value of X, and additional or alternative costs." So the
// spell is remembered as it stood on the stack, exactly as epic remembers
// one (epic.go): the card and its stack item as their snapshot records
// (EpicSpell), carried by the delayed trigger as plain data, so a table
// waiting on it is a restore point and the trigger adds no route from
// Game to a closure.
//
//   - The exile is the spell leaving the stack without being countered
//     (ExileSpellKeyedForEffect: "can't be countered" does not stop it,
//     and nothing watching for a counter fires). Its continuation puts
//     the delay counters on the card in exile and binds the trigger to
//     that object.
//   - The trigger fires at each of that player's upkeeps while the card
//     is still that object in exile (DelayedTrigger.WhileExiled, CR
//     603.4's intervening "if"); once the card has left exile it is a new
//     object (CR 400.7) and the trigger is dropped. Its resolution checks
//     again, removes a delay counter, and when none is left puts the CARD
//     onto the stack with the original spell's choices (putExiledCardOnStackLocked).
//   - That spell is not cast (CR 707.10), so no cast trigger sees it, and
//     nobody chooses new targets: the copy has the original's.
//
// The card itself goes onto the stack, as the text says ("puts it onto the
// stack"), so when it resolves or is countered it goes where a card goes.

// CounterDelay is the delay counter (Ertai's Meddling, CR 122.1). No SBA:
// a countdown its trigger reads. MoveCard clears counters on the way out
// of exile (CR 400.7).
const CounterDelay = "delay"

// delayRecastBody is the upkeep trigger's body, an on-disk key.
var delayRecastBody = SimpleDelayedBody("meddling/delay", resolveDelayRecast)

// delayRecastExiled is the exile's continuation as data: it puts N delay
// counters on the card and binds the trigger Carry[0] names to it. Player
// is the spell's controller. An on-disk key.
var delayRecastExiled = RegisterTail("engine/meddling-exiled", func(g *Game, o TailOutcome) error {
	var triggerID uuid.UUID
	if len(o.Tail.Carry) > 0 {
		triggerID = o.Tail.Carry[0]
	}
	if len(o.IDs) != 1 || exiledCardByIDLocked(g, o.IDs[0]) == nil {
		// The spell did not reach exile (a commander took CR 903.9's
		// offer, or it had already left the stack): nothing to wait on.
		g.dropDelayedTriggerLocked(triggerID)
		return nil
	}
	cardID := o.IDs[0]
	if err := g.AddCounterByForEffect(o.Tail.Player, cardID, CounterDelay, o.Tail.N); err != nil {
		return err
	}
	ref := g.sourceObjectRefLocked(cardID)
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.ID == triggerID {
			dt.WhileExiled = ref
		}
	}
	return nil
})

// ExileSpellWithDelayCountersForEffect is Ertai's Meddling's resolution on
// the spell `spellID`: its controller exiles it with `n` delay counters on
// it, and at the beginning of each of their upkeeps, while it is exiled, a
// delay counter is removed; when none is left they put it onto the stack
// as a copy of the original spell (CR 707.10). `source` is the Meddling,
// `label` names the trigger. Returns false and does nothing when spellID
// is not a spell on the stack or n is not positive.
//
// Caller must hold g.mu (write), which a resolution frame does.
func (g *Game) ExileSpellWithDelayCountersForEffect(spellID uuid.UUID, n int, source uuid.UUID, label string) (bool, error) {
	if n <= 0 {
		return false, nil
	}
	src, item, ok := g.stackSpellLocked(spellID)
	if !ok || item == nil || item.Kind != StackItemSpell {
		return false, nil
	}
	controller := item.Controller
	cen := &ContinuationCensus{}
	rec := &EpicSpell{
		Card: snapshotCard(src, cen),
		Item: snapshotStackItemAs(item, src.OracleID, cen),
	}
	if label == "" {
		label = src.Name + " — remove a delay counter"
	}
	forever := IndefiniteDuration()
	id := uuid.New()
	g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
		ID:                 id,
		Controller:         controller,
		SourceCardID:       source,
		Label:              label,
		At:                 StepUpkeep,
		ControllerTurnOnly: true,
		Duration:           &forever,
		Body:               delayRecastBody,
		Params:             EffectParams{Object: ObjectRef{ID: id}},
		CopiesSpell:        rec,
		// Bound to the card once it is in exile (delayRecastExiled);
		// until then a placeholder that no card in exile matches, so
		// the trigger cannot fire for a spell that never got there.
		WhileExiled: ObjectRef{ID: spellID, Epoch: -1},
	})
	err := g.ExileSpellKeyedForEffect(spellID, KeyedTail{Ref: delayRecastExiled, Player: controller, N: n, Carry: []uuid.UUID{id}})
	return true, err
}

// resolveDelayRecast is the upkeep trigger resolving: if the card is still
// the object in exile, remove a delay counter; if none is left, put it onto
// the stack as a copy of the original spell (CR 603.4, 707.10).
func resolveDelayRecast(g *Game, item *StackItem) error {
	dt := g.delayRecastTriggerLocked(item)
	if dt == nil || dt.CopiesSpell == nil || !g.objectInExileLocked(dt.WhileExiled) {
		return nil
	}
	cardID := dt.WhileExiled.ID
	if err := g.AddCounterForEffect(cardID, CounterDelay, -1); err != nil {
		return err
	}
	c := exiledCardByIDLocked(g, cardID)
	if c == nil || c.Counters[CounterDelay] > 0 {
		return nil
	}
	rec := dt.CopiesSpell
	g.dropDelayedTriggerLocked(dt.ID)
	return g.putExiledCardOnStackLocked(cardID, rec, item.Controller)
}

// delayRecastTriggerLocked finds the trigger that put `item` on the stack,
// which its params name. Caller must hold g.mu.
func (g *Game) delayRecastTriggerLocked(item *StackItem) *DelayedTrigger {
	if item == nil {
		return nil
	}
	if id := item.Params.Object.ID; id != uuid.Nil {
		for _, dt := range g.DelayedTriggers {
			if dt != nil && dt.ID == id {
				return dt
			}
		}
	}
	return nil
}

// dropDelayedTriggerLocked removes the delayed trigger `id`, if queued.
// The queue is replaced, not written through (an undo snapshot shares
// it). Caller must hold g.mu.
func (g *Game) dropDelayedTriggerLocked(id uuid.UUID) {
	if id == uuid.Nil {
		return
	}
	var keep []*DelayedTrigger
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.ID == id {
			continue
		}
		keep = append(keep, dt)
	}
	g.DelayedTriggers = keep
}

// objectInExileLocked reports whether `ref` (instance and epoch) is a card
// in exile right now. Caller must hold g.mu.
func (g *Game) objectInExileLocked(ref ObjectRef) bool {
	if ref.ID == uuid.Nil {
		return false
	}
	c := exiledCardByIDLocked(g, ref.ID)
	return c != nil && c.ObjectEpoch == ref.Epoch
}

// putExiledCardOnStackLocked moves the card `cardID` from exile onto the
// stack as the spell `rec` was, under `controller`: the original's
// targets, modes, X and paid costs (CR 707.10). It is put there, not
// cast, so no EventCast is emitted and no cast trigger sees it; the
// things it targets become the target of a spell (CR 115.3). Caller must
// hold g.mu (write).
func (g *Game) putExiledCardOnStackLocked(cardID uuid.UUID, rec *EpicSpell, controller uuid.UUID) error {
	spell, _ := restoreStackItem(&rec.Item)
	if spell == nil {
		return nil
	}
	face := rec.Card.ActiveFace
	if _, err := MoveCard(g.Exile, g.Stack, cardID); err != nil {
		return err
	}
	for i := range g.Stack.Cards {
		if g.Stack.Cards[i].InstanceID == cardID {
			g.Stack.Cards[i].SetFace(face)
			g.Stack.Cards[i].Controller = controller
		}
	}
	g.markCardKnownInZoneLocked(g.Stack, cardID)
	if g.StackMeta == nil {
		g.StackMeta = make(map[uuid.UUID]*StackItem)
	}
	spell.ID = cardID
	spell.Kind = StackItemSpell
	spell.Controller = controller
	spell.SourceCardID = cardID
	g.StackMeta[cardID] = spell
	g.recomputeSplitSecondLocked()
	g.EmitEvent(Event{
		Kind:    EventZoneMove,
		Actor:   controller,
		CardID:  cardID,
		OldZone: ZoneExile,
		NewZone: ZoneStack,
	})
	g.emitBecameTargetLocked(controller, cardID, cardID, spell.Targets)
	return nil
}
