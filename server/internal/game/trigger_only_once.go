package game

import "github.com/google/uuid"

// trigger_only_once.go — "This ability triggers only once." on a
// triggered ability (#5245). Acrobatic Cheerleader's "Survival — At the
// beginning of your second main phase, if this creature is tapped, put a
// flying counter on it. This ability triggers only once."
//
// It is the triggered twin of "Activate only once" (activate_only_once.go)
// and the game-lifetime twin of "This ability triggers only once each
// turn" (effects.TriggersOnlyOnceEachTurn, which reads TurnTally.Triggered
// and is emptied by the turn advance). The printed rulings (2024-09-20)
// fix what it means:
//
//   - "Once Acrobatic Cheerleader's ability triggers, it won't trigger
//     again as long as it remains on the battlefield, even if the
//     triggered ability is countered or the flying counter is somehow
//     removed." So the record is written when the ability TRIGGERS, not
//     when it resolves, and never reset.
//   - "If an effect would cause Acrobatic Cheerleader's ability to
//     trigger again, that part of the effect won't do anything." So a
//     trigger doubler adds no instance: harvestMatchLocked dispatches the
//     one instance and does not ask the doublers.
//   - "If Acrobatic Cheerleader leaves the battlefield and then returns,
//     it's a new object with no memory of its previous existence. Its
//     ability will be able to trigger again." CR 400.7. So the key is the
//     object, ObjectTallyKey(source, Card.ObjectEpoch, Key), the key every
//     per-object tally uses. A flicker bumps the epoch, and the returned
//     permanent reads a key nothing has written.
//
// A trigger that never happened writes nothing: a failed intervening if
// (AppliesTo said no), a suppressor (Torpor Orb's kind) or a zero
// occurrence count each return before the claim. A change of control is
// not a zone change, so the ability stays spent for its new controller.
//
// The record is Game.OnlyOnceTriggered, carried by the snapshot (additive
// within v9) and by the undo clone, so a restore keeps it spent and an
// undo past the trigger gives it back.

// claimTriggerOnlyOnceLocked reports whether the "triggers only once"
// ability `key` of `source` may trigger now, and records that it has when
// it may. False when this object has already triggered it.
//
// The epoch is the one on the harvested value copy, which is the object
// the event saw. Caller must hold g.mu in write mode.
func (g *Game) claimTriggerOnlyOnceLocked(source *Card, key string) bool {
	k := ObjectTallyKey(source.InstanceID, source.ObjectEpoch, key)
	if g.OnlyOnceTriggered[k] > 0 {
		return false
	}
	if g.OnlyOnceTriggered == nil {
		g.OnlyOnceTriggered = map[string]int{}
	}
	g.OnlyOnceTriggered[k]++
	return true
}

// TriggeredOnlyOnceSpent reports whether the "triggers only once" ability
// labelled `key` of the object `source` names right now has already
// triggered. A source that has changed zones since reads false, because
// it is a new object (CR 400.7).
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) TriggeredOnlyOnceSpent(source uuid.UUID, key string) bool {
	return g.OnlyOnceTriggered[g.objectTallyKeyLocked(source, key)] > 0
}
