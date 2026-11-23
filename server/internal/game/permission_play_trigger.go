package game

import "github.com/google/uuid"

// permission_play_trigger.go — #5904: "when they do" after a play made
// through a permission a permanent grants.
//
// Share the Spoils: "During each player's turn, that player may play a
// land or cast a spell from among cards exiled with this enchantment ...
// When they do, exile the top card of their library." The second
// sentence is a triggered ability of the enchantment that triggers on
// the play made THROUGH its permission (CR 603.2): not on any play of a
// card exiled with it, which another effect might open. So the trigger
// is created by the permission's FollowUp, which runs exactly when the
// cast or (with FollowUpOnLandPlay) the land play was made because of
// that permission, and it is handed to the same dispatch a reflexive
// trigger takes (dispatchReflexiveTriggerLocked): it goes on the stack
// the next time a player would receive priority, under the controller
// of the permanent that grants the permission, and every player may
// respond to it. The trigger is data (a body key and its Params), so a
// restore point carries it as it carries any reflexive trigger.

// runLandPlayFollowUpLocked runs the permission's follow-up for a land
// played through it, when the permission says its follow-up covers a
// land play (CastPermission.FollowUpOnLandPlay). `followUp` is the key
// the cast path already settled was the play's own (the permission was
// the reason for it); empty runs nothing.
//
// Caller must hold g.mu (write).
func (g *Game) runLandPlayFollowUpLocked(grant *CastPermission, followUp CastFollowUpKey, player, source, land uuid.UUID, object ObjectRef) error {
	if grant == nil || !grant.FollowUpOnLandPlay || followUp == "" {
		return nil
	}
	return g.runCastFollowUpLocked(followUp, CastFollowUp{Player: player, Source: source, Spell: land, Object: object})
}

// QueueTriggerForPermissionPlayForEffect creates, from a cast
// follow-up, the triggered ability "when they do, ..." of the permanent
// whose permission the play was made through (f.Source). The trigger's
// controller is that permanent's controller and its source is that
// object; rt names its Label, Body (a ReflexiveBody key) and Params.
// Payload and Optional are carried as a reflexive trigger's are.
//
// Returns false, creating nothing, for a malformed declaration or a
// source that is no longer on the battlefield (a permission derived
// from a permanent cannot outlive it, so this is a guard, not a rule).
//
// Caller must hold g.mu (write): a follow-up body always does.
func (g *Game) QueueTriggerForPermissionPlayForEffect(f CastFollowUp, rt ReflexiveTrigger) bool {
	if rt.Body.key == "" || rt.Label == "" || f.Source == uuid.Nil {
		return false
	}
	src := g.findBattlefieldCardLocked(f.Source)
	if src == nil {
		return false
	}
	ev := Event{
		Kind:   EventCast,
		Actor:  f.Player,
		Source: f.Source,
		CardID: f.Spell,
		Label:  rt.Label,
	}
	g.dispatchReflexiveTriggerLocked(src.InstanceID, src.Controller,
		ObjectRef{ID: src.InstanceID, Epoch: src.ObjectEpoch}, rt, ev)
	return true
}
