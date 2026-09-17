package game

import "github.com/google/uuid"

// attackers.go — the attack declaration's lock-in (#859).
//
// CR 508.1 makes declaring attackers ONE turn-based action, and
// CR 508.2 gives the active player priority afterwards — the boundary
// at which the triggers the declaration produced go on the stack.
// This engine's DeclareAttacker is a per-creature verb the client
// sends once per click while the declare-attackers window is open,
// and the sandbox's "Re-declare attacker" re-points an already
// declared attacker at a different defending player, planeswalker or
// battle before the declaration is finished.
//
// So the verb only STAGES the attack on Card.AttackingTarget.
// Nothing is announced until the declaration is locked in, which is
// the first point at which play moves on inside the step:
//
//   - runStateChecksLocked, the engine's "a player would receive
//     priority" boundary (a trick cast in the step, a resolution), and
//   - the two places the cursor can leave the step — AdvanceStep and
//     PassPriority's wrap — which call runStateChecksLocked themselves
//     when a declaration is still staged, so the lock-in always
//     happens INSIDE declare_attackers and never a step late.
//
// commitAttackDeclarationLocked is the one place attack-declaration
// triggers are harvested from: it emits the events, and the ordinary
// kind-keyed event harvester (triggers.go) does the rest. This is the
// attack-side twin of #830's commitBlockDeclarationLocked, and it is
// here for the same reason: before #859 the per-click EventAttack was
// the trigger-bearing event, so a creature declared against player A
// and then re-pointed at player B kept naming A — every "attacks
// <player>" reader, and every trigger CONDITION gated on the defender
// (Curse of Opulence's enchanted player, Breena's "one of your
// opponents", the "attacked a player" tests a planeswalker re-point
// flips), saw the defender the attacker had already left.
//
// One difference from the block side, and it is CR 508.1's: a
// creature is declared as an attacker ONCE. announcedAttacks is keyed
// on PRESENCE, not on the defender it named, so an attacker that has
// been announced is never announced again this combat however often
// it is re-pointed afterwards — exactly the "one EventAttack per
// declared attacker" contract Adeline and every other OncePerBatch
// attack payoff is built on. EventBlock, which is per (blocker,
// attacker) PAIR, compares the pairing instead.

// attackDeclarationPendingLocked reports whether any creature is
// staged as an attacker without having been announced yet — i.e.
// whether a lock-in would emit anything. Cheap battlefield scan;
// callers use it to avoid running a state-check pass for nothing.
//
// Caller must hold g.mu.
func (g *Game) attackDeclarationPendingLocked() bool {
	if g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.AttackingTarget == uuid.Nil {
			continue
		}
		if !g.announcedAttacks[c.InstanceID] {
			return true
		}
	}
	return false
}

// commitAttackDeclarationLocked locks the staged attack declaration
// in: one EventAttack per creature that is attacking and has not been
// announced yet, carrying the defender it is attacking NOW.
//
// Idempotent, and cheap when there is nothing to do, so it can sit at
// the top of runStateChecksLocked. Everything it emits lands in one
// event batch (nothing here opens a new one), which is what makes a
// whole declaration one occurrence for OncePerBatch — the property
// Adeline's one batch of Humans depends on (#854, ADR 0049).
//
// The battlefield is walked in order, so the events of one
// declaration are deterministic and a replay reproduces them.
//
// Caller must hold g.mu in write mode.
func (g *Game) commitAttackDeclarationLocked() {
	if !g.attackDeclarationPendingLocked() {
		return
	}
	// Collect by value first: EmitEvent harvests triggers, whose
	// Build functions can reallocate the battlefield slice out from
	// under a *Card.
	type declaration struct {
		attacker   uuid.UUID
		controller uuid.UUID
		defender   uuid.UUID
	}
	var fresh []declaration
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.AttackingTarget == uuid.Nil || g.announcedAttacks[c.InstanceID] {
			continue
		}
		fresh = append(fresh, declaration{
			attacker:   c.InstanceID,
			controller: c.Controller,
			defender:   c.AttackingTarget,
		})
	}
	for _, d := range fresh {
		g.noteAttackAnnouncedLocked(d.attacker)
	}
	for _, d := range fresh {
		g.EmitEvent(Event{
			Kind:   EventAttack,
			Actor:  d.controller,
			CardID: d.attacker,
			Target: d.defender,
		})
	}
}

// noteAttackAnnouncedLocked records that `id` has had its one
// EventAttack for this combat, so no later lock-in announces it
// again.
//
// It is also how a permanent that was PUT onto the battlefield
// attacking is kept out of the declaration entirely (CR 506.3c): such
// a permanent was never DECLARED as an attacker, so it fires no
// "whenever ~ attacks" trigger and nothing that watches attack
// declarations sees it. Marking it here is what stops the
// battlefield scan above from mistaking its AttackingTarget for a
// staged declaration — see CreateTokensForEffect and
// CreateTokensAttackingForEffect in effect_api.go, the two entry
// points that can mint one.
//
// Caller must hold g.mu.
func (g *Game) noteAttackAnnouncedLocked(id uuid.UUID) {
	if id == uuid.Nil {
		return
	}
	if g.announcedAttacks == nil {
		g.announcedAttacks = map[uuid.UUID]bool{}
	}
	g.announcedAttacks[id] = true
}

// clearAttackAnnouncementsLocked forgets this combat's announcements.
// Called from clearCombatLocked, alongside the AttackingTarget wipe
// they describe: next combat's declaration is a new declaration and
// announces again. Caller must hold g.mu.
func (g *Game) clearAttackAnnouncementsLocked() {
	g.announcedAttacks = nil
}
