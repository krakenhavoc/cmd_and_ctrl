package game

import "github.com/google/uuid"

// move_reflexive.go — a reflexive trigger created by a replacement that
// applied to a move (ADR 0178 amendment of 2026-10-29, #4517).
//
// Head of the Hunt: "If a creature an opponent controls would die,
// exile it instead. When you do, create a 2/2 green Wolf creature
// token." The second sentence is a CR 603.12 reflexive trigger, but its
// creator is not a resolving spell or ability: it is the replacement,
// and "when you do" means "when you exile it this way". So the trigger
// exists only for a move this replacement modified, and only once that
// move has put the card where the replacement sent it.
//
// Triggering on the exile EVENT instead would be stronger than printed:
// it would fire when Rest in Peace, or any other effect, did the
// exiling. So, as ADR 0178 does for a life change and ADR 0188 for an
// extra turn, the replacement only DECLARES the trigger on the move
// event (QueueReflexiveTriggerAfterReplacedMoveForEffect), and the
// move's finisher (runMoveTailLocked) creates it once the card has
// landed, through the same dispatch a resolving effect's reflexive
// trigger takes (dispatchReflexiveTriggerLocked). The dispatch puts it
// on PendingTriggers, so it goes on the stack the next time a player
// would receive priority (CR 603.3), under the replacement's
// controller.
//
// Each move is its own event, so a board wipe that exiles three
// creatures declares three instructions and creates three triggers
// (CR 603.12a). Queuing a trigger never pauses, so, like the extra
// turn, it needs none of moveLifeSequence's continuation machinery.
//
// The instruction is plain data (a body key, not a closure), so a move
// event paused on a later CR 616 question with one declared is still a
// restore point; its frame is written under a key of its own,
// ReplacementResumeMoveReflexive, which a binary that would drop the
// instruction refuses.

// reflexiveMoveInstruction is one declared reflexive trigger: the
// replacement's source and controller, fixed as it applies (CR 603.12:
// the trigger's controller is the creator's), the zone the replacement
// sent the card to, and the trigger itself as data.
type reflexiveMoveInstruction struct {
	Source       uuid.UUID
	SourceObject ObjectRef
	Controller   uuid.UUID
	Zone         ZoneKind
	Label        string
	Body         string
	Params       EffectParams
}

// QueueReflexiveTriggerAfterReplacedMoveForEffect declares, from a
// replacement's Replace callback, the CR 603.12 reflexive trigger the
// replacement creates ("exile it instead. When you do, …"). Call it
// AFTER rewriting the event's destination: the trigger is created once
// the card has landed in ev.NewZone as it stands now, and not at all if
// a later replacement sends it elsewhere or cancels the move (the
// replacement's action did not happen), or if the move is abandoned
// (CR 614.7).
//
// src is the replacement's source as the callback receives it; its
// controller controls the trigger. rt names the trigger's Label, Body
// (a ReflexiveBody key, with any target clause declared there) and
// Params. Optional and Payload are not carried and are refused.
//
// Returns ErrInvalidParam for an event that is not a zone change or a
// malformed declaration.
//
// Caller must hold g.mu (Replace callbacks always do).
func (g *Game) QueueReflexiveTriggerAfterReplacedMoveForEffect(ev *ReplacementEvent, src *Card, rt ReflexiveTrigger) error {
	if ev == nil || ev.Kind != RepEventMove || src == nil || rt.Body.key == "" || rt.Label == "" ||
		rt.Optional != nil || len(rt.Payload) > 0 {
		return ErrInvalidParam
	}
	// A fresh backing array: an undo snapshot taken while this event was
	// paused shares the old one (cloneReplacementResume copies too).
	n := len(ev.reflexiveAfterMove)
	ev.reflexiveAfterMove = append(ev.reflexiveAfterMove[:n:n], reflexiveMoveInstruction{
		Source:       src.InstanceID,
		SourceObject: ObjectRef{ID: src.InstanceID, Epoch: src.ObjectEpoch},
		Controller:   src.Controller,
		Zone:         ev.NewZone,
		Label:        rt.Label,
		Body:         rt.Body.key,
		Params:       cloneEffectParams(rt.Params),
	})
	return nil
}

// queueReflexiveTriggersAfterMoveLocked creates each declared reflexive
// trigger whose condition was met: the move was not cancelled and the
// card is now in the zone the replacement sent it to. It clears the list
// first, so a re-entered tail cannot create them twice. A controller who
// has left the game gets none (CR 800.4a).
//
// Caller must hold g.mu.
func (g *Game) queueReflexiveTriggersAfterMoveLocked(ev *ReplacementEvent) {
	if ev == nil || len(ev.reflexiveAfterMove) == 0 {
		return
	}
	in := ev.reflexiveAfterMove
	ev.reflexiveAfterMove = nil
	if ev.Canceled {
		return
	}
	z := g.findCardZoneLocked(ev.CardID)
	for _, r := range in {
		if z == nil || z.Kind != r.Zone {
			continue
		}
		if p := g.playerByIDLocked(r.Controller); p == nil || p.Eliminated {
			continue
		}
		// The modified event, as a description of what created the
		// trigger. It is never emitted.
		created := Event{
			Kind:    EventZoneMove,
			Actor:   r.Controller,
			Source:  r.Source,
			CardID:  ev.CardID,
			OldZone: ev.OldZone,
			NewZone: r.Zone,
			Label:   r.Label,
		}
		g.dispatchReflexiveTriggerLocked(r.Source, r.Controller, r.SourceObject,
			ReflexiveTrigger{Label: r.Label, Body: BodyRef{key: r.Body}, Params: r.Params}, created)
	}
}
