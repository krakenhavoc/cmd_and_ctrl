package game

import "github.com/google/uuid"

// permission_play_delayed_trigger.go — #6091: "When you play a card this
// way, ..." (Fires of Mount Doom) and "When you cast that spell, ..."
// (Ogre Battlecaster).
//
// These are DELAYED triggered abilities, created by the activated or
// triggered ability that granted the play (CR 603.7a). Their source is
// that ability's source and their controller is the player who
// controlled that ability as it resolved (CR 603.7e). Unlike Share the
// Spoils's "when they do" (permission_play_trigger.go), which is a
// triggered ability of a permanent and needs that permanent on the
// battlefield, a delayed trigger still triggers after its source has
// left: the source is then read by last-known information (CR 608.2h),
// which dispatchReflexiveTriggerLocked's source lookup already does.
//
// The trigger is created from the permission's FollowUp, which runs
// exactly when the play was made BECAUSE of that permission (with
// FollowUpOnLandPlay, a land play too), so it is data on disk as a
// reflexive trigger is.

// QueueDelayedTriggerForPermissionPlayForEffect creates, from a cast
// follow-up, the delayed triggered ability "when you play a card this
// way, ..." of the ability that granted the permission: its source is
// f.Source (the permission's Source), read by last-known information when
// it has left the battlefield, and its controller is controller (the
// player who controlled the granting ability as it resolved, CR 603.7e).
// rt names its Label, Body and Params, as QueueTriggerForPermissionPlayForEffect's.
// A permission whose FollowUpObject names its Source carries the source
// object (its epoch), so the trigger's source is that object even after
// it has left.
//
// Returns false, creating nothing, for a malformed declaration.
//
// Caller must hold g.mu (write): a follow-up body always does.
func (g *Game) QueueDelayedTriggerForPermissionPlayForEffect(f CastFollowUp, controller uuid.UUID, rt ReflexiveTrigger) bool {
	if rt.Body.key == "" || rt.Label == "" || f.Source == uuid.Nil || controller == uuid.Nil {
		return false
	}
	// The granting ability's source object, as it was put on the stack,
	// when the permission carried it (FollowUpObject naming f.Source);
	// otherwise the object now on the battlefield, if any.
	ref := ObjectRef{ID: f.Source}
	if f.Object.ID == f.Source {
		ref = f.Object
	} else if src := g.findBattlefieldCardLocked(f.Source); src != nil {
		ref.Epoch = src.ObjectEpoch
	}
	ev := Event{
		Kind:   EventCast,
		Actor:  f.Player,
		Source: f.Source,
		CardID: f.Spell,
		Label:  rt.Label,
	}
	g.dispatchReflexiveTriggerLocked(f.Source, controller, ref, rt, ev)
	return true
}
