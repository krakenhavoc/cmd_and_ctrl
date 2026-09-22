package game

import "github.com/google/uuid"

// control_effects.go is gain-and-exchange of control from a spell or
// an ability (CR 613.1b, CR 701.12; ADR 0063, #756).
//
// Before S38 the only way into the layer-2 control bucket was an
// Aura's static ability (`ControlAttachedBySource`, Mind Control).
// Everything the bucket needed was already there:
//
//   - `Card.BaseController`, captured on the first recompute after a
//     permanent enters (layers.go) and cleared by `MoveCard`, is the
//     value the printed characteristic reseeds from every pass. So
//     control REVERTS by itself when an effect stops applying, and
//     nothing has to remember who had it.
//   - The per-bucket timestamp sort is CR 613.7: an Act of Treason
//     resolved after a Mind Control wins because it is later, and
//     when it expires the Mind Control is applying again.
//   - `materialiseControlLocked` writes layer 2's answer back onto
//     `Card.Controller` and carries the two consequences with it —
//     summoning sickness under the new controller (CR 302.6) and
//     removal from combat (CR 506.4).
//
// So a gain of control is not a new mechanism. It is one
// `ScopedStatic` in the layer-2 bucket with a `Duration` on it, and
// the whole of this file is the two ways a spell can ask for one.

// GainControlForEffect registers a layer-2 continuous effect handing
// control of `target` to `controller` for `d` (CR 613.1b).
//
// Returns false, having registered nothing, when `target` is not on
// the battlefield: there is no object to change the control of.
//
// The effect is PINNED to the target as it is now — instance ID plus
// battlefield-entry stamp. That is CR 611.2c (the affected set is
// locked in when the effect starts) and CR 400.7 (a permanent that
// leaves and returns is a new object), and it is what makes a stolen
// creature flickered in response come back to its owner. It is also
// the registry's garbage collection: an indefinite theft stops
// existing the moment the stolen permanent does.
//
// `sourceID` is the spell or ability that created the effect, stored
// as last-known information for logs and for `source`-reading
// predicates. `controller` is an arbitrary player, so "target
// opponent gains control" is the same call with a different argument
// — what those cards still wait on is the choose-a-player prompt,
// not this primitive.
//
// Caller must hold g.mu (write). Effects call this from inside the
// resolution frame, which already holds it.
func (g *Game) GainControlForEffect(sourceID, target, controller uuid.UUID, d Duration, label string) bool {
	if controller == uuid.Nil {
		return false
	}
	c, ok := g.battlefieldCardLocked(target)
	if !ok {
		return false
	}
	g.RegisterScopedStaticForEffect(
		controlStatic(target, c.EnteredBattlefieldAt, controller),
		sourceID, label, g.PinnedTo(d, target))
	return true
}

// ExchangeControlForEffect exchanges control of two permanents
// (CR 701.12), as Switcheroo does.
//
// ALL OR NOTHING (CR 701.12b): both objects are checked before
// either effect is registered, and if either has left the
// battlefield no control is exchanged at all — the function returns
// false having changed nothing.
//
// ONE EFFECT, ONE TIMESTAMP (CR 613.7). The two halves are two
// entries in the registry because they apply to two different
// objects, but they share a single timestamp, so a later
// control-changer beats both of them or neither. Two clock reads
// would make the exchange something a well-timed third effect could
// split in half.
//
// Each half is `Indefinite` (CR 611.2a — an exchange states no
// duration) and pinned to its own object. If one of the two later
// leaves the battlefield, that half stops applying and the other
// stands: the exchange already happened.
//
// Caller must hold g.mu (write).
func (g *Game) ExchangeControlForEffect(sourceID, a, b uuid.UUID, label string) bool {
	if a == uuid.Nil || b == uuid.Nil || a == b {
		return false
	}
	ca, okA := g.battlefieldCardLocked(a)
	cb, okB := g.battlefieldCardLocked(b)
	if !okA || !okB {
		return false
	}
	aTo, bTo := cb.Controller, ca.Controller
	aStamp, bStamp := ca.EnteredBattlefieldAt, cb.EnteredBattlefieldAt
	if aTo == uuid.Nil || bTo == uuid.Nil {
		return false
	}
	ts := timeNowUnixNano()
	g.registerScopedStaticLocked(controlStatic(a, aStamp, aTo), sourceID, label,
		g.PinnedTo(IndefiniteDuration(), a), ts)
	g.registerScopedStaticLocked(controlStatic(b, bStamp, bTo), sourceID, label,
		g.PinnedTo(IndefiniteDuration(), b), ts)
	return true
}

// controlStatic is the layer-2 continuous effect itself: "that
// permanent is controlled by this player". Both closures capture
// values only — an instance ID, a stamp and a player ID — per the
// closure contract on ScopedStatic.
func controlStatic(target uuid.UUID, enteredAt int64, controller uuid.UUID) StaticAbility {
	return StaticAbility{
		Layer: Layer2Control,
		AppliesTo: func(c *Card, _ *Game, _ *Card) bool {
			return c.InstanceID == target && c.EnteredBattlefieldAt == enteredAt
		},
		Apply: func(ch *Characteristic, _ *Card, _ *Game, src *Card) {
			ch.Controller = controller
			// #930: whoever wrote Controller last in this bucket is
			// the effect that won CR 613.7, and the event names it.
			// `src` is the ScopedStatic's stored source card, so this
			// is the spell or ability that took the permanent.
			if src != nil {
				ch.ControlSource = src.InstanceID
			}
		},
	}
}

// controlChange is one permanent's layer-2 control delta, collected
// by materialiseControlLocked and emitted by the recompute once the
// pass has finished. Values only — the battlefield slice the walk
// read them off may be reallocated by the time they are emitted.
type controlChange struct {
	card   uuid.UUID
	from   uuid.UUID
	to     uuid.UUID
	source uuid.UUID
}

// emitControlChangesLocked emits one EventControlChanged per delta
// the materialise step found (#930, CR 613.1b).
//
// ONE event and ONE emission point, for the reason
// materialiseControlLocked is one step: every way control can move —
// an Aura's static, a spell's scoped static, an exchange, a duration
// expiring, the Mind Control being destroyed — is a layer-2 delta and
// arrives here. A gain emitted from GainControlForEffect would have
// missed all four of the others.
//
// Nothing here opens an event batch, so the whole pass is one
// occurrence: CR 701.12's exchange is two events with one Batch, and
// a "whenever one or more" ability guarded by OncePerBatch fires once
// for it (#829, CR 603.2c).
//
// The order is battlefield order, which is deterministic and what a
// replay reproduces.
//
// Caller must hold g.mu in write mode.
func (g *Game) emitControlChangesLocked(changed []controlChange) {
	for _, ch := range changed {
		g.EmitEvent(Event{
			Kind:   EventControlChanged,
			Actor:  ch.to,
			Target: ch.from,
			CardID: ch.card,
			Source: ch.source,
		})
	}
}

// removeFromCombatLocked takes one permanent out of combat
// (CR 506.4): its declaration AND the announcement bookkeeping that
// describes the declaration.
//
// The announcements are the part that was missing. `announcedAttacks`
// (#830/#859) records that a creature has had its one `EventAttack`
// for this combat, and `announcedBlocks` / `blockedAttackers` do the
// same for blocks. Clearing `AttackingTarget` without clearing
// them left a creature that a control change pulled out of combat
// marked as already-announced, so if it was declared as an attacker
// again in the same combat — under its new controller, after a second
// declare-attackers step, or after the control effect ended — it
// would fire no "whenever ~ attacks" trigger. `clearCombatLocked`
// clears both together for the whole board at end of combat; this is
// the same pairing for one permanent.
//
// Caller must hold g.mu.
func (g *Game) removeFromCombatLocked(c *Card) {
	if c == nil {
		return
	}
	// #1218: only when this permanent was actually ATTACKING — a
	// blocker pulled out of combat never was, and bumping for it would
	// pay for a recompute nothing needs. See
	// invalidateLayersForAttackChangeLocked.
	if c.AttackingTarget != uuid.Nil {
		g.invalidateLayersForAttackChangeLocked()
	}
	c.AttackingTarget = uuid.Nil
	c.BlockingTarget = uuid.Nil
	delete(g.announcedAttacks, c.InstanceID)
	delete(g.announcedBlocks, c.InstanceID)
	// Its own blocked state only: an attacker removed from combat is
	// no longer blocked because it is no longer in combat at all
	// (CR 506.4). Taking a BLOCKER out of combat leaves the attacker
	// it was blocking blocked, which is CR 509.1h and the whole point
	// of #715 — that row is keyed by the attacker and is not touched
	// here.
	delete(g.blockedAttackers, c.InstanceID)
}
