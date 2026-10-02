package game

import (
	"fmt"

	"github.com/google/uuid"
)

// until_return.go — CR 610.3, an exile "until" an event (#1729).
//
// The rule, from the September 25, 2026 edition AGENTS.md §6 pins:
//
//	610.3. Some one-shot effects cause an object to change zones
//	"until" a specified event occurs. A second one-shot effect is
//	created immediately after the specified event. This second
//	one-shot effect returns the object to its previous zone.
//
//	610.3a If a resolving spell or activated ability creates the
//	initial one-shot effect that causes the object to change zones,
//	and the specified event has already occurred before that one-shot
//	effect would occur but after that spell or ability was put onto
//	the stack, the object doesn't move.
//
//	610.3b If a resolving triggered ability creates the initial
//	one-shot effect that causes the object to change zones, and the
//	specified event has already occurred before that one-shot effect
//	would occur but after that ability triggered, the object doesn't
//	move.
//
//	610.3c An object returned to the battlefield this way returns
//	under its owner's control unless otherwise specified.
//
//	610.3d If multiple one-shot effects are created this way
//	immediately after one or more simultaneous events, those one-shot
//	effects are also simultaneous.
//
// Before this file the return was a TRIGGERED ability: Palace Jailer
// scheduled an event-conditioned delayed trigger, and Ossification and
// Hostage Taker each carried a leaves-the-battlefield trigger. That was
// wrong in two observable ways.
//
//   - It used the stack. Every player got a window in which the card
//     was still in exile. The Hostage Taker ruling (2017-09-29) says
//     "Nothing happens between the two events, including state-based
//     actions."
//   - It left with its controller. A delayed trigger is controlled by
//     the player whose ability made it, and CR 800.4a and 800.4d take
//     a departed player's triggers with them. So a Palace Jailer whose
//     controller conceded kept its creature for good, even though CR
//     725.4 hands the crown to one of their opponents as they leave,
//     which is the printed "until". The same ruling covers the
//     leaves-the-battlefield shape: "if Hostage Taker's owner leaves
//     the game while the card is still exiled and another player owns
//     that card, the exiled card will return to the battlefield under
//     its owner's control. Because the one-shot effect that returns
//     the card isn't an ability that goes on the stack, it won't cease
//     to exist along with the leaving player's spells and abilities on
//     the stack."
//
// THE SHAPE. A return is a record in Game.DelayedTriggers with Until
// set, so it clones, snapshots and shows on the wire with everything
// else still owed. It is never fired by the trigger harvester. Two
// kinds of "until":
//
//   - UNTIL AN EVENT (Palace Jailer's "until an opponent becomes the
//     monarch"): On and Condition, exactly as an event-conditioned
//     delayed trigger names them. The untilReturns listener marks the
//     record Due when a matching event is emitted.
//   - UNTIL AN OBJECT LEAVES THE BATTLEFIELD (Ossification, Hostage
//     Taker, Sheltered by Ghosts): UntilLeaves names the object. The
//     record is due once that object is no longer on the battlefield,
//     by any route. A structural test rather than an event match,
//     because one route — the object's owner leaving the game — emits
//     no zone-change event at all (ADR 0060), and it is still the
//     object leaving the battlefield.
//
// WHEN THE RETURN HAPPENS. resolveUntilReturnsLocked runs every due
// return together (CR 610.3d) at the first point after the event
// where nothing else is in flight:
//
//   - as a player would receive priority (runStateChecksLocked), before
//     the state-based actions, and again after each state-based pass,
//     because a pass can be the event (a lethal state check takes a
//     player out of the game, and a creature dying is an object
//     leaving the battlefield);
//   - as play moves on to a new event batch (beginEventBatchLocked),
//     which catches a departure that moved the turn on before any
//     state check ran;
//   - right after a concession and after the sandbox's manual crown,
//     the two mutations that can be the event and run no state check.
//
// So no player can act between the event and the return. What it does
// not reproduce is the instant: the rest of a resolving spell runs
// first. A card that reads the battlefield later in the SAME resolution
// sees the returned object still in exile. Nothing in the catalog reads
// one that way.
//
// LEAVING THE GAME. A record is never dropped with its controller, for
// the ruling quoted above. A record whose card is no longer in exile as
// the object it exiled — cast from exile (Hostage Taker), moved by
// another effect, or taken out of the game with its owner — is
// discarded at the next flush, because there is nothing left to return.

// UntilReturn describes the second one-shot effect of an exile "until"
// an event: what returns, and which event returns it. Exactly one of
// Leaves and On should be set.
type UntilReturn struct {
	// Controller is the controller of the effect that exiled the card
	// (CR 603.7d's reading of who "you" and "an opponent" are). It
	// takes no action of its own; a Condition reads it.
	Controller uuid.UUID

	// SourceCardID is the card whose ability did the exiling, for
	// attribution on the wire and in the log.
	SourceCardID uuid.UUID

	// Label is what the table sees while the card is held.
	Label string

	// Leaves is "until <this object> leaves the battlefield".
	Leaves ObjectRef

	// On, Condition and CondParams are "until <an event>", in the same
	// vocabulary an event-conditioned delayed trigger uses.
	On         []EventKind
	Condition  ConditionRef
	CondParams EffectParams
}

// untilReturnBody is the return itself: put Params.Object back onto
// the battlefield under its owner's control if it is still that object
// in exile. Registered under a key of its own rather than reusing
// flicker's, because the key is also what an older binary refuses: a
// binary from before #1729 does not know this record is not a trigger,
// and must not restore it as one (ErrUnknownEffectKey keeps the file
// for a newer build). The flush does not call it per record; it
// batches every due return into one entry (CR 610.3d) through the same
// helper.
var untilReturnBody BodyRef

func init() {
	untilReturnBody = DelayedBody("until/return-to-battlefield", func(g *Game, _ *StackItem, p EffectParams) error {
		return g.returnUntilObjectsLocked([]ObjectRef{p.Object})
	})
}

// ScheduleUntilReturnForEffect records the return of the card `cardID`,
// which an "until" effect has just exiled. The card must be in exile:
// the record names the object it is NOW (instance and CR 400.7 epoch),
// so a card that leaves exile and comes back is not returned. Reports
// the record's ID, or uuid.Nil when nothing was recorded (the card is
// not in exile, or the request names no event).
//
// Call it after the exile has landed — from the exile's continuation,
// because a commander can pause the exile on the CR 903.9 prompt and
// one that went to the command zone has nothing to come back from.
//
// Caller must hold g.mu (write).
func (g *Game) ScheduleUntilReturnForEffect(cardID uuid.UUID, u UntilReturn) uuid.UUID {
	if cardID == uuid.Nil || (u.Leaves.ID == uuid.Nil && len(u.On) == 0) {
		return uuid.Nil
	}
	c, ok := g.cardInZoneLocked(g.Exile, cardID)
	if !ok {
		return uuid.Nil
	}
	if u.Condition.key != "" {
		if _, known := lookupCondition(u.Condition.key); !known {
			effectKeyFault(fmt.Sprintf("game: until return %q names an unregistered condition %q", u.Label, u.Condition.key))
			return uuid.Nil
		}
	}
	duration := IndefiniteDuration()
	return g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
		Controller:   u.Controller,
		SourceCardID: u.SourceCardID,
		Label:        u.Label,
		Cards:        []uuid.UUID{cardID},
		Body:         untilReturnBody,
		Params:       EffectParams{Object: ObjectRef{ID: cardID, Epoch: c.ObjectEpoch}},
		On:           append([]EventKind(nil), u.On...),
		Condition:    u.Condition,
		CondParams:   cloneEffectParams(u.CondParams),
		Duration:     &duration,
		Until:        true,
		UntilLeaves:  u.Leaves,
	})
}

// UntilEventHappenedForEffect is CR 610.3a/b, asked by the resolving
// effect BEFORE it exiles: has the event that would end the exile
// already happened since the ability was put on the stack? If so the
// object doesn't move.
//
// For "until <object> leaves the battlefield", the event has happened
// when that object is not on the battlefield now. For an event-keyed
// "until", the event log is read back from the ability's triggering
// event (CR 610.3b's "after that ability triggered"). An item with no
// triggering event (a spell or an activated ability) has nothing to
// read back from yet, and answers false.
//
// Caller must hold g.mu.
func (g *Game) UntilEventHappenedForEffect(u UntilReturn, item *StackItem) bool {
	if u.Leaves.ID != uuid.Nil {
		return !g.objectOnBattlefieldLocked(u.Leaves)
	}
	if len(u.On) == 0 || item == nil || item.Trigger == nil || !item.Trigger.Fired() {
		return false
	}
	since := item.Trigger.Event.Seq
	probe := &DelayedTrigger{Controller: u.Controller, SourceCardID: u.SourceCardID,
		On: u.On, Condition: u.Condition, CondParams: u.CondParams}
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Seq <= since {
			break
		}
		if probe.matchesEventLocked(ev, g) {
			return true
		}
	}
	return false
}

// UntilReturnPendingForEffect reports whether an "until" return is
// still recorded for the card `cardID` — the legacy guard on the
// leaves-the-battlefield triggers the Oblivion Ring family kept (see
// effects/exile_until.go): a card exiled before #1729 has no record, and
// its old trigger still brings it back.
//
// Caller must hold g.mu.
func (g *Game) UntilReturnPendingForEffect(cardID uuid.UUID) bool {
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.Until && dt.Params.Object.ID == cardID {
			return true
		}
	}
	return false
}

// objectOnBattlefieldLocked reports whether the object `ref` names is
// still on the battlefield — the same instance with the same CR 400.7
// epoch. A phased-out permanent is still on the battlefield (CR
// 702.26d treats it as though it doesn't exist, but it has not left),
// so the holding slice counts.
//
// Caller must hold g.mu.
func (g *Game) objectOnBattlefieldLocked(ref ObjectRef) bool {
	if ref.ID == uuid.Nil {
		return false
	}
	for _, z := range []*Zone{g.Battlefield, g.PhasedOut} {
		if z == nil {
			continue
		}
		for i := range z.Cards {
			if z.Cards[i].InstanceID == ref.ID {
				return z.Cards[i].ObjectEpoch == ref.Epoch
			}
		}
	}
	return false
}

// untilRecordLiveLocked reports whether the card an "until" record
// holds is still in exile as the object it exiled. A record that fails
// has nothing to return and is discarded.
//
// Caller must hold g.mu.
func (g *Game) untilRecordLiveLocked(dt *DelayedTrigger) bool {
	c, ok := g.cardInZoneLocked(g.Exile, dt.Params.Object.ID)
	return ok && c.ObjectEpoch == dt.Params.Object.Epoch
}

// untilRecordDueLocked reports whether an "until" record's event has
// happened.
//
// Caller must hold g.mu.
func (g *Game) untilRecordDueLocked(dt *DelayedTrigger) bool {
	if dt.Due {
		return true
	}
	return dt.UntilLeaves.ID != uuid.Nil && !g.objectOnBattlefieldLocked(dt.UntilLeaves)
}

// resolveUntilReturnsLocked is the flush: every "until" return whose
// event has happened is taken off the queue and performed, together
// (CR 610.3d), and every record whose card is no longer there to
// return is discarded. Reports whether it returned anything, so the
// state-check loop knows the board moved.
//
// Never on a game that has ended: its board is frozen (ADR 0060
// Decision 5).
//
// Caller must hold g.mu (write).
func (g *Game) resolveUntilReturnsLocked() bool {
	if g.State != StateActive || len(g.DelayedTriggers) == 0 {
		return false
	}
	// A resolution still running, or paused on one of its own prompts,
	// finishes first; the state-check boundary it ends at flushes. The
	// same test holdForOpenResolutionLocked makes, without its side
	// effect.
	if g.resolutionDepth > 0 || (g.resolutionOpen && g.pausedResolutionChoiceLocked() != nil) {
		return false
	}
	var due []ObjectRef
	changed := false
	kept := make([]*DelayedTrigger, 0, len(g.DelayedTriggers))
	for _, dt := range g.DelayedTriggers {
		if dt == nil || !dt.Until {
			kept = append(kept, dt)
			continue
		}
		if !g.untilRecordLiveLocked(dt) {
			changed = true
			continue
		}
		if g.untilRecordDueLocked(dt) {
			due = append(due, dt.Params.Object)
			changed = true
			continue
		}
		kept = append(kept, dt)
	}
	if !changed {
		return false
	}
	if len(kept) == 0 {
		kept = nil
	}
	// The queue is rewritten BEFORE the return, for the reason
	// fireDelayedTriggersLocked gives: the entry emits events, and a
	// listener reading a queue that still held these records would see
	// them twice.
	g.DelayedTriggers = kept
	if len(due) == 0 {
		return false
	}
	if err := g.returnUntilObjectsLocked(due); err != nil {
		g.EmitEvent(Event{
			Kind:     EventEffectError,
			ErrorMsg: "CR 610.3: an until return failed: " + err.Error(),
		})
	}
	return true
}

// returnUntilObjectsLocked puts every object in `refs` that is still in
// exile as that object back onto the battlefield, as one entry and
// under each card's owner's control (CR 610.3c, 610.3d). A token has
// ceased to exist (CR 111.8) and a card that is not a permanent card
// cannot enter, so neither is part of the entry.
//
// Caller must hold g.mu (write).
func (g *Game) returnUntilObjectsLocked(refs []ObjectRef) error {
	entries := make([]BatchEntry, 0, len(refs))
	for _, ref := range refs {
		c, ok := g.cardInZoneLocked(g.Exile, ref.ID)
		if !ok || c.ObjectEpoch != ref.Epoch || c.IsToken() || !c.IsPermanent() {
			continue
		}
		// CR 800.4b: an object that would be put onto the battlefield
		// under the control of a player who has left the game stays
		// where it is. Its owner leaving took it out of the game anyway
		// (CR 800.4a), so this only guards a card whose owner record is
		// gone.
		if !g.playerStillInLocked(c.Owner) {
			continue
		}
		entries = append(entries, BatchEntry{CardID: ref.ID, From: ZoneExile})
	}
	if len(entries) == 0 {
		return nil
	}
	return g.startEntryBatchLocked(entries, ZoneEntryOptions{}, nil)
}

// untilReturns is the listener half: an event-keyed "until" becomes
// due when its event is emitted. A listener of its own rather than a
// branch of the trigger harvester, because this is a rule of the game
// and not a triggered ability — it must work with no catalog loaded,
// and no trigger suppressor may stop it.
type untilReturns struct{}

// OnEvent marks every matching event-keyed record due. The queue is
// rewritten with copies rather than written through, because a record
// may be shared with an undo snapshot.
//
// Caller (notifyListenersLocked) holds g.mu in write mode.
func (untilReturns) OnEvent(g *Game, ev Event) {
	if len(g.DelayedTriggers) == 0 || ev.Kind == EventTrigger {
		return
	}
	var out []*DelayedTrigger
	for i, dt := range g.DelayedTriggers {
		if dt == nil || !dt.Until || dt.Due || len(dt.On) == 0 || !dt.matchesEventLocked(ev, g) {
			continue
		}
		if out == nil {
			out = append([]*DelayedTrigger(nil), g.DelayedTriggers...)
		}
		cp := cloneDelayedTrigger(dt)
		cp.Due = true
		out[i] = cp
	}
	if out != nil {
		g.DelayedTriggers = out
	}
}
