package game

import "github.com/google/uuid"

// attack_target.go is S27's polymorphic attack declaration: an
// attacking creature may be declared against a PLAYER, a
// PLANESWALKER, or a BATTLE (CR 506.2, 508.1d).
//
// WHY Card.AttackingTarget IS OVERLOADED RATHER THAN SPLIT. The field
// already held "the thing this creature is attacking" and happened to
// only ever hold a player ID. Adding a second field would mean two
// fields with an invariant between them — exactly one set, ever — and
// nothing in Go to enforce it, plus a second entry in the snapshot,
// the clone, the drift test and the wire. Instead the field's TYPE is
// unchanged and its DOMAIN widened: a uuid that resolves to a seated
// player is a player attack, one that resolves to a battlefield
// permanent is a permanent attack. uuid collisions between a seat and
// a card do not happen, and every read goes through the helpers below
// rather than assuming.
//
// The consequence worth stating: `g.playerByIDLocked(c.AttackingTarget)`
// now returns nil for a legal attack, so any call site that treated
// nil as "corrupt state" would silently drop the attack. The two that
// mattered — combat damage and the block-decision scan — are routed
// through defendingPlayerForAttackLocked here.
//
// WHAT IS NOT MODELLED. "Attacks each combat if able" redirection,
// propaganda-style attack taxes, and the requirement that a battle be
// attacked by someone other than its protector's controller are
// enforcement questions the sandbox has never taken on for attacks;
// the one restriction that IS enforced is the protector's own, because
// a player attacking a battle they protect is nonsense rather than a
// judgement call (CR 310.9b).

// AttackTargetKind classifies what an attack declaration names.
type AttackTargetKind string

const (
	// AttackTargetPlayer — the ordinary case, CR 506.2.
	AttackTargetPlayer AttackTargetKind = "player"
	// AttackTargetPlaneswalker — CR 508.1d.
	AttackTargetPlaneswalker AttackTargetKind = "planeswalker"
	// AttackTargetBattle — CR 508.1d, post-MoM.
	AttackTargetBattle AttackTargetKind = "battle"
	// AttackTargetNone — the id names nothing attackable.
	AttackTargetNone AttackTargetKind = ""
)

// classifyAttackTargetLocked reports what `target` is: a seated,
// non-eliminated player, a battlefield planeswalker, a battlefield
// battle, or nothing.
//
// Players are checked FIRST. A seat id and a card id are both random
// uuids and cannot collide in practice, but the order makes the
// precedence explicit rather than incidental.
//
// Caller must hold g.mu.
func (g *Game) classifyAttackTargetLocked(target uuid.UUID) AttackTargetKind {
	if target == uuid.Nil {
		return AttackTargetNone
	}
	if p := g.playerByIDLocked(target); p != nil {
		if p.Eliminated {
			return AttackTargetNone
		}
		return AttackTargetPlayer
	}
	c := findBattlefieldCard(g, target)
	if c == nil {
		return AttackTargetNone
	}
	// Effective types, not printed: a permanent animated out of being
	// a planeswalker has stopped being attackable as one, and a land
	// turned into a battle by nothing that exists yet would become
	// attackable the day it does.
	if c.IsPlaneswalker() {
		return AttackTargetPlaneswalker
	}
	if c.IsBattle() {
		return AttackTargetBattle
	}
	return AttackTargetNone
}

// defendingPlayerForAttackLocked resolves the player who is defending
// against an attack declared at `target`.
//
// For a player attack that is the player. For a planeswalker it is
// the walker's controller (CR 506.2 — the planeswalker's controller
// is the defending player, and it is their creatures that may block).
// For a battle it is the PROTECTOR, not the controller: a battle's
// controller is the player who cast it, and the whole point of the
// protector mechanic is that somebody ELSE defends it (CR 310.9d).
//
// Returns uuid.Nil when the target no longer resolves — a
// planeswalker that died before the damage step, a battle already
// exiled. Callers treat that as "the attack hits nothing", which is
// the rule: an attacker whose defending target has left the
// battlefield deals its damage to nothing (CR 510.1b).
//
// Caller must hold g.mu.
func (g *Game) defendingPlayerForAttackLocked(target uuid.UUID) uuid.UUID {
	switch g.classifyAttackTargetLocked(target) {
	case AttackTargetPlayer:
		return target
	case AttackTargetPlaneswalker:
		if c := findBattlefieldCard(g, target); c != nil {
			return c.Controller
		}
	case AttackTargetBattle:
		if c := findBattlefieldCard(g, target); c != nil {
			if c.ProtectorPlayerID != uuid.Nil {
				return c.ProtectorPlayerID
			}
			// A battle with no protector — one that entered before
			// the protector prompt was answered, or a fixture — falls
			// back to its controller so blockers and combat damage
			// still have a seat to talk about. Weaker than printed
			// (the controller would not normally be the defender) and
			// never stronger.
			return c.Controller
		}
	}
	return uuid.Nil
}

// canAttackTargetLocked reports whether `attacker`'s controller may
// declare an attack at `target`, and why not when they may not.
//
// The rules enforced, and only these:
//
//	CR 506.2   you cannot attack yourself, and you cannot attack a
//	           planeswalker you control
//	CR 310.9b  you cannot attack a battle you protect — you are the
//	           one defending it
//	           (and a battle you control that nobody protects is not
//	           attackable by you either, for the same reason the
//	           fallback above makes you its defender)
//
// Everything else about attack LEGALITY — propaganda taxes, "can't
// attack unless", goad — remains the sandbox's to arbitrate, exactly
// as it was for player attacks.
//
// Caller must hold g.mu.
func (g *Game) canAttackTargetLocked(attackerController, target uuid.UUID) error {
	switch g.classifyAttackTargetLocked(target) {
	case AttackTargetNone:
		return ErrIllegalAttackTarget
	case AttackTargetPlayer:
		if target == attackerController {
			return ErrIllegalAttackTarget
		}
		return nil
	case AttackTargetPlaneswalker, AttackTargetBattle:
		if g.defendingPlayerForAttackLocked(target) == attackerController {
			return ErrIllegalAttackTarget
		}
		return nil
	}
	return ErrIllegalAttackTarget
}

// AttackTargetsForEffect lists every legal attack target for
// `attackerController` right now: the other seated players, the
// planeswalkers they do not control, and the battles they do not
// protect.
//
// Read-only; the view layer calls it under the read lock to stamp
// the client's picker, and the legal-move enumerator calls it to
// price an attack. It does not consider the attacking creature at
// all — summoning sickness, defender and tapped state gate the
// ATTACKER, not the target, and are checked separately.
func (g *Game) AttackTargetsForEffect(attackerController uuid.UUID) []AttackTargetRef {
	var out []AttackTargetRef
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == attackerController {
			continue
		}
		out = append(out, AttackTargetRef{Kind: AttackTargetPlayer, ID: p.ID})
	}
	for _, c := range g.Battlefield.Cards {
		var kind AttackTargetKind
		switch {
		case c.IsPlaneswalker():
			kind = AttackTargetPlaneswalker
		case c.IsBattle():
			kind = AttackTargetBattle
		default:
			continue
		}
		if g.canAttackTargetLocked(attackerController, c.InstanceID) != nil {
			continue
		}
		out = append(out, AttackTargetRef{Kind: kind, ID: c.InstanceID})
	}
	return out
}

// dealCombatDamageToAttackTargetLocked routes an attacker's combat
// damage to whatever it was declared against (CR 510.1b): a player's
// life total, a planeswalker's loyalty, or a battle's defense.
//
// The split is here rather than inside markCombatDamageToPlayerLocked
// because the two halves are genuinely different events: damage to a
// player runs the commander-damage tally and the life change, damage
// to a permanent runs the CR 120.3 counter removal. Only the
// replacement pipeline and lifelink are common, and both of those
// already live inside the two helpers this dispatches to.
//
// A target that has left the battlefield since declaration deals its
// damage to nothing (CR 510.1b — the attacker is removed from combat
// and is simply not dealt with further). Silently dropping it is the
// rule, not a shortcut.
//
// step is the substep's Event.CombatStep value (#187), passed through
// to whichever helper deals the damage.
//
// Caller must hold g.mu.
func (g *Game) dealCombatDamageToAttackTargetLocked(target, source uuid.UUID, amount int, step string) {
	if amount <= 0 {
		return
	}
	switch g.classifyAttackTargetLocked(target) {
	case AttackTargetPlayer:
		g.markCombatDamageToPlayerLocked(target, source, amount, step)
	case AttackTargetPlaneswalker, AttackTargetBattle:
		g.markCombatDamageOnCardLocked(target, amount, source, step)
	}
}

// DefendingPlayerForAttackForEffect is the exported read of
// defendingPlayerForAttackLocked, for the legal-move enumerator and
// the view layer — both of which need to know which seat an attack is
// really aimed at without importing the game package's internals.
//
// Read-only. Caller must hold g.mu (read or write); every current
// caller runs inside ReadSnapshot or the enumerator's own frame.
func (g *Game) DefendingPlayerForAttackForEffect(target uuid.UUID) uuid.UUID {
	return g.defendingPlayerForAttackLocked(target)
}

// ClassifyAttackTargetForEffect is the exported read of
// classifyAttackTargetLocked, for the view layer's
// AttackingTargetKind projection. Returns AttackTargetNone (the empty
// string) for an id that names nothing attackable.
//
// Read-only. Caller must hold g.mu (read or write).
func (g *Game) ClassifyAttackTargetForEffect(target uuid.UUID) AttackTargetKind {
	return g.classifyAttackTargetLocked(target)
}

// AttackTargetRef is one entry in the legal attack-target set: what
// it is, and its id. The id is a seat id for a player and an instance
// id for a permanent — the same overload Card.AttackingTarget carries,
// with the kind alongside so the client never has to guess.
type AttackTargetRef struct {
	Kind AttackTargetKind
	ID   uuid.UUID
}
