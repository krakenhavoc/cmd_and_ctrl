package game

import "github.com/google/uuid"

// reflexive.go — CR 603.12 reflexive triggered abilities.
//
// "Tap any number of untapped creatures you control. WHEN YOU TAP
// ONE OR MORE CREATURES THIS WAY, …"; "you may sacrifice another
// creature. WHEN YOU DO, …". The second sentence is a triggered
// ability created by the FIRST one while it resolves. It has no
// existence before that resolution and no existence afterwards: it
// triggers only if its condition was met during the resolution that
// made it, and then it goes on the stack like any other trigger, so
// every player gets a response window before it resolves.
//
// Two consequences the catalog kept getting wrong by folding the
// second sentence into the first one's effect:
//
//   - **The response window.** Folded, the follow-up happens inside
//     the parent's resolution and nobody can respond to it. This is
//     the same mistake ADR 0018 retired for ordinary triggers and
//     ADR 0026 §3 for delayed ones.
//   - **The targets.** A reflexive trigger's targets are chosen as
//     it is put on the stack (CR 603.3d), which is AFTER the parent
//     resolved — knowing what was revealed, what was sacrificed, what
//     was milled. Folded, the card has to hang the clause on the
//     parent instead and the controller picks a target before making
//     the choice the trigger is about. Every folded card declared
//     that as a caveat.
//
// So a reflexive trigger is not a new dispatch. It is the harvester's
// dispatch, entered from a resolving effect instead of from an event:
//
//	parent ability resolves → QueueReflexiveTriggerForEffect
//	  → dispatchTriggerLocked  (the same one harvestFromZone calls)
//	    → Targets: legal set computed NOW, empty ⇒ dropped with no
//	      prompt (CR 603.3d), otherwise a pick_target prompt
//	    → OptionalPrompt: the CR 603.5 yes/no
//	    → Build → PendingTriggers → StackMeta → resolve
//
// Reusing dispatchTriggerLocked is the whole design. A second
// targeting path would be a second place for the CR 603.3d drop, the
// CR 608.2b re-check, the "becomes the target" emit and the APNAP
// drain to disagree with the first one.
//
// Why it is NOT a DelayedTrigger (ADR 0026 named "every 'when you
// do'" as wanting that slot; it does not): a delayed trigger waits
// for a STEP and fires from the step-entry hook, and its condition is
// the step. A reflexive trigger's condition was already met — by the
// resolution that created it — so it goes on the stack at the next
// priority boundary, above the parent, which is the printed order. It
// never sits in a queue on the Game and so never has to survive a
// turn; what it does have to survive is Clone / RestoreFrom between
// creation and resolution, and it does that exactly as a harvested
// trigger does, because by then it IS one.
//
// Undo: the item on PendingTriggers is an ordinary StackItem, cloned
// by cloneStackItem with its Effect shared on the contract
// StackItem.Effect documents — the closure takes the live *Game and
// its own item at resolve time and captures neither. A reflexive
// trigger still waiting on its target prompt rides the ordinary
// pickTargetFrame. Which is why the payload is data on the item
// (Payload) rather than a variable the Effect closes over.

// ReflexiveTrigger is the declaration a resolving effect hands the
// engine to create a CR 603.12 reflexive triggered ability. Every
// field except Label and Effect is optional.
//
// The controller and the source are NOT fields: CR 603.12 fixes them
// as the controller and the source of the ability that created the
// trigger, which is the resolving stack item passed alongside.
type ReflexiveTrigger struct {
	// Label is the stack-overlay copy, in the same shape every other
	// trigger uses: "Ziatora, the Incinerator — damage equal to the
	// sacrificed creature's power".
	Label string

	// Targets is the trigger's target clause. Chosen when the
	// trigger goes on the stack (CR 603.3d) — which for a reflexive
	// trigger is the point of the exercise, because the parent has
	// already resolved and the controller now knows what it did. An
	// empty legal set drops the trigger with no prompt, exactly as
	// for a harvested trigger. Nil for an untargeted follow-up.
	Targets *TargetSpec

	// Optional makes the trigger a CR 603.5 "you may". Nil for the
	// usual mandatory "when you do".
	//
	// Note that the "you may" of "you MAY sacrifice another
	// creature. When you do, …" belongs to the PARENT ability, not
	// here: "when you do" is mandatory once you did.
	Optional *TriggerOptionalPrompt

	// Payload is what the parent resolution has to tell the trigger:
	// the creatures that were tapped, the card that was sacrificed,
	// the player who was chosen. Stamped onto the built item's
	// Payload and read back off the item by the Effect.
	//
	// It is deliberately not Targets. The trigger's Targets slot
	// belongs to its own target clause and is overwritten when the
	// pick comes in, so a payload parked there would be lost on
	// exactly the triggers that most need one.
	Payload []TargetRef

	// Effect is what the trigger does when its stack item resolves.
	// Same contract as StackItem.Effect: read the controller, the
	// source, the chosen targets and the payload off the item it is
	// handed, and capture neither a *Game nor a pointer into a zone
	// slice — undo restores a cloned game and the closure has to
	// resolve against that one.
	//
	// Runs under g.mu held in write mode.
	Effect func(g *Game, item *StackItem) error
}

// QueueReflexiveTriggerForEffect creates the reflexive triggered
// ability `rt` from inside the resolution of `parent`, and hands it to
// the same dispatch a harvested trigger takes.
//
// `parent` is the stack item currently resolving. Its controller
// controls the reflexive trigger and its source card is the trigger's
// source — which may already have left: a sorcery's reflexive trigger
// has its source in a graveyard, and a land that sacrificed itself has
// one there too. That is the normal case and needs no special
// handling, the same way harvestLTB and the delayed-trigger queue
// need none: the trigger carries the source's ID for attribution and
// reads everything else off its own item.
//
// The condition ("when you DO") is the caller's to evaluate — call
// this only once the thing happened. Nothing here re-checks it.
//
// Returns false when the declaration is malformed (no parent, no
// Effect, no Label), which is dropped rather than queued. A true
// return does NOT promise a stack item: a targeted trigger with no
// legal target is removed from the stack per CR 603.3d, and an
// optional one can still be declined.
//
// Caller must hold g.mu in write mode — the caller is a resolving
// effect, which already does.
func (g *Game) QueueReflexiveTriggerForEffect(parent *StackItem, rt ReflexiveTrigger) bool {
	if parent == nil || rt.Effect == nil || rt.Label == "" {
		return false
	}
	source, lki := g.reflexiveSourceLocked(parent)
	label := rt.Label
	effect := rt.Effect
	payload := append([]TargetRef(nil), rt.Payload...)
	// CR 603.12: the reflexive trigger's source is the parent's, and
	// so is the OBJECT (#1418) — the resolving parent already names
	// it, and the card may have moved since. A move the parent itself
	// made onto the battlefield is followed (#1432,
	// followedSourceObjectLocked): "return this to the battlefield.
	// When you do, …" is about the permanent it returned.
	sourceObject := g.followedSourceObjectLocked(parent)
	ability := TriggeredAbility{
		// Watches / AppliesTo stay empty: this ability is never
		// harvested off an event, so nothing ever looks at them. The
		// dispatch reads Targets, OptionalPrompt and Build, and it
		// is handed the match rather than asked to find one.
		Key:            label,
		Targets:        rt.Targets,
		OptionalPrompt: rt.Optional,
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			item := NewTriggeredItem(source, label, effect)
			item.SourceObject = sourceObject
			item.Payload = append([]TargetRef(nil), payload...)
			return item
		},
	}
	// The event the dispatch carries into the resume frames and back
	// into Build. A reflexive trigger has no triggering event — it
	// was created by a resolution — so this describes that
	// resolution. It is never emitted; Build ignores it. It exists
	// because dispatchTriggerLocked's signature is the harvester's,
	// and a zero Event in a stashed frame is harder to read in a
	// debugger than an accurate one.
	ev := Event{
		Kind:   EventResolve,
		Actor:  parent.Controller,
		Source: parent.SourceCardID,
		Label:  parent.Label,
	}
	g.dispatchTriggerLocked(ev, source, lki, ability)
	return true
}

// reflexiveSourceLocked builds the source-card value the dispatch
// works from: the parent ability's source card as it can still be
// found, with the controller and instance ID forced to the parent's.
//
// Forcing them is the rule, not a convenience. CR 603.12 gives the
// reflexive trigger to the controller of the ability that created it,
// so a permanent that changed hands mid-resolution — or one that is
// sitting in a graveyard with a stale Controller — must not decide
// who gets the trigger.
//
// A source that has left the battlefield keeps its last known
// characteristics when the LKI snapshot still holds them (CR 603.10),
// and a source that cannot be found anywhere yields a bare Card with
// just the two forced fields, which is all NewTriggeredItem and the
// prompts need.
//
// Caller must hold g.mu.
func (g *Game) reflexiveSourceLocked(parent *StackItem) (Card, Characteristic) {
	return g.triggerSourceLocked(parent.SourceCardID, parent.Controller)
}

// triggerSourceLocked is that body, shared with the #663
// event-conditioned delayed trigger, which needs exactly the same
// thing for exactly the same reason: a source card that may be
// anywhere or nowhere, and a controller the RULE fixes rather than the
// card's current Controller field.
//
// Caller must hold g.mu.
func (g *Game) triggerSourceLocked(sourceCardID, controller uuid.UUID) (Card, Characteristic) {
	var source Card
	var lki Characteristic
	if sourceCardID != uuid.Nil {
		if c := g.findCardByIDLocked(sourceCardID); c != nil {
			source = *c
			if known, ok := g.lastKnownBattlefield[sourceCardID]; ok {
				lki = known
			} else {
				lki = c.Effective()
			}
		}
	}
	source.InstanceID = sourceCardID
	source.Controller = controller
	return source, lki
}
