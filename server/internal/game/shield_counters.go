package game

import "github.com/google/uuid"

// shield_counters.go — CR 122.1c (#2000):
//
//	122.1c One or more shield counters on a permanent create a single
//	       replacement effect and a single prevention effect that
//	       protect the permanent. These effects are "If this permanent
//	       would be destroyed as the result of an effect, instead remove
//	       a shield counter from it" and "If damage would be dealt to
//	       this permanent, prevent that damage and remove a shield
//	       counter from it."
//
// Two engine-owned replacements (builtin_replacements.go), because the
// effects are the counter's, printed in the rules, and outlive whatever
// put the counter on. "Single" is the registration: one built-in each,
// however many counters, so two counters never ask a CR 616 question of
// each other.
//
//   - DESTRUCTION "as the result of an effect" only. A state-based
//     destruction (CR 704.5g lethal damage, 704.5h deathtouch) takes
//     sbaDestroyRoute and is not replaced. Damage that reaches the
//     permanent was not prevented, so it can only be damage that can't
//     be prevented. Indestructible filters before the window opens, so
//     an indestructible permanent keeps its counter. Beside a
//     regeneration shield, the controller orders the two (CR 616.1).
//   - DAMAGE is a PREVENTION effect (CR 615) whose removal of a counter
//     is its additional effect (CR 615.5), so it is the ADR 0108 §8
//     shape the Phantoms use: Replace only prevents, and Then
//     (shieldCounterRemoveBody) is owed once per RECIPIENT per damage
//     instance (ThenPerRecipient). Damage dealt at the same time by
//     several sources — two blockers — is all prevented and costs one
//     counter. Under "can't be prevented" the effect still applies and
//     prevents nothing, and its additional effect still happens (CR
//     615.12): the damage is dealt AND a counter is removed. A built-in
//     has no object of its own, so thenOnDamaged owes the follow-up on
//     the damaged permanent (prevention_then.go).
//
// Neither is Preemptive: the permanent's controller orders them against
// any other replacement that applies (CR 616.1).

// shieldCounterLabel names both built-ins in a CR 616 prompt, and is how
// the trigger-order check (trigger_independence.go) knows them.
const shieldCounterLabel = "Shield counter"

// shieldCounterDestroyReplacement is CR 122.1c's replacement effect.
var shieldCounterDestroyReplacement = ReplacementEffect{
	Watches: []EventKind{EventZoneMove},
	AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
		return g.shieldCounterStopsDestructionLocked(ev)
	},
	Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
		ev.Cancel()
		return g.applyCounterLocked(ev.CardID, CounterShield, -1)
	},
	Controller: func(ev *ReplacementEvent, g *Game, _ *Card) uuid.UUID {
		return g.controllerOfBattlefieldCardLocked(ev.CardID)
	},
	Label: shieldCounterLabel,
}

// shieldCounterPreventsDamageReplacement is CR 122.1c's prevention
// effect.
var shieldCounterPreventsDamageReplacement = ReplacementEffect{
	Watches:    []EventKind{EventDealDamage},
	Prevention: true,
	AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
		return g.shieldCounterPreventsDamageLocked(ev)
	},
	Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
		// CR 615.1: prevented damage is not dealt at all.
		ev.Cancel()
		return nil
	},
	Then:          shieldCounterRemoveBody,
	ThenPer:       ThenPerRecipient,
	thenOnDamaged: true,
	Controller: func(ev *ReplacementEvent, g *Game, _ *Card) uuid.UUID {
		return g.controllerOfBattlefieldCardLocked(ev.DamageTarget)
	},
	Label: shieldCounterLabel,
}

// shieldCounterRemoveBody is the prevention's additional effect, "remove
// a shield counter from it", registered under "shield-counter/remove"
// (ADR 0041 P9): an on-disk identity, never renamed. The follow-up names
// the permanent as SourceObject, so a permanent that left and came back
// is a new object and loses nothing (CR 400.7).
var shieldCounterRemoveBody = SimpleDelayedBody("shield-counter/remove", func(g *Game, item *StackItem) error {
	ref := item.SourceObject
	c := findBattlefieldCard(g, ref.ID)
	if c == nil || c.ObjectEpoch != ref.Epoch || c.Counters[CounterShield] <= 0 {
		return nil
	}
	return g.applyCounterLocked(c.InstanceID, CounterShield, -1)
})

// shieldCounterStopsDestructionLocked is the replacement's AppliesTo: a
// destruction by an effect of a permanent with a shield counter.
//
// Caller must hold g.mu.
func (g *Game) shieldCounterStopsDestructionLocked(ev *ReplacementEvent) bool {
	if ev == nil || ev.Kind != RepEventMove || !ev.Destruction || ev.StateBasedDestruction {
		return false
	}
	c := findBattlefieldCard(g, ev.CardID)
	return c != nil && c.Counters[CounterShield] > 0
}

// shieldCounterPreventsDamageLocked is the prevention's AppliesTo:
// damage to a permanent with a shield counter.
//
// Caller must hold g.mu.
func (g *Game) shieldCounterPreventsDamageLocked(ev *ReplacementEvent) bool {
	if ev == nil || ev.Kind != RepEventDamage || ev.DamageAmount <= 0 {
		return false
	}
	c := findBattlefieldCard(g, ev.DamageTarget)
	return c != nil && c.Counters[CounterShield] > 0
}

// anyShieldCounterLocked reports whether any permanent has a shield
// counter on it: the only board on which CR 122.1c's effects can apply.
//
// Caller must hold g.mu.
func (g *Game) anyShieldCounterLocked() bool {
	if g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Counters[CounterShield] > 0 {
			return true
		}
	}
	return false
}
