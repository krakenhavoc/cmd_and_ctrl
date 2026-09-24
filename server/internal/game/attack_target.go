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
// WHAT IS NOT MODELLED. "Attacks each combat if able" redirection and
// the requirement that a battle be attacked by someone other than its
// protector's controller are enforcement questions the sandbox has
// never taken on for attacks; the one restriction that IS enforced is
// the protector's own, because a player attacking a battle they
// protect is nonsense rather than a judgement call (CR 310.9b).
//
// Propaganda-style attack TAXES used to be on that list and are not
// any more: they ship in attack_tax.go, charged by the declaration
// verbs at CR 508.1a (ADR 0080, #1063). They are not a target-legality
// question, which is why they are not here — defendingPlayerForAttack-
// Locked is what the pricer uses this file for, so an attack on a
// planeswalker is taxed by its controller's Propaganda.

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

// setAttackTargetLocked points `c`'s attack at `target` and records
// the defending player that attack has right now (#1364). Every write
// that makes a creature attack something goes through it — the two
// declaration verbs, an entry "attacking" (CR 506.3c) and a reselect
// (CR 508.7) — so the record cannot lag the target it describes.
//
// A target with no defending player (uuid.Nil, or one that no longer
// resolves) drops the record rather than keeping the previous one: a
// record names the attack the creature is making, never an older one.
//
// Caller must hold g.mu.
func (g *Game) setAttackTargetLocked(c *Card, target uuid.UUID) {
	if c == nil {
		return
	}
	c.AttackingTarget = target
	defender := g.defendingPlayerForAttackLocked(target)
	if defender == uuid.Nil {
		g.forgetAttackDefenderLocked(c.InstanceID)
		return
	}
	if g.attackDefenders == nil {
		g.attackDefenders = map[uuid.UUID]uuid.UUID{}
	}
	g.attackDefenders[c.InstanceID] = defender
}

// AttackingNothing is the Card.AttackingTarget of a creature that is
// still attacking though what it attacked has been removed from
// combat without leaving the battlefield (CR 506.4c, #1376, ADR 0045
// Decision 36): the planeswalker or battle changed control or phased
// out.
//
// WHY A SENTINEL RATHER THAN A FLAG. A creature attacking nothing must
// still read as attacking (every `AttackingTarget != uuid.Nil` check —
// dozens of them, in the engine, the enumerator, the view, the bots
// and the catalog), and must read as attacking NOTHING everywhere the
// target is resolved: combat damage, the card-side "attacking you"
// readers, the view's attacking_target_kind, the reselect label. Both
// are true of an id that names no seat and no card, with no edit at
// any of those readers. That is the shape a creature whose walker DIED
// has had all along — its target is the id of a card no longer on the
// battlefield — and it is the shape Decision 35's block fallback
// already handles. A flag beside a live walker id would have had to
// be consulted by every reader that resolves the id, and one that
// forgot would deal the damage.
//
// The value is a version-0 uuid, so it cannot collide with a seat or
// instance id (both uuid.New, version 4).
var AttackingNothing = uuid.MustParse("00000000-0000-0000-0000-000000000506")

// removeAttackedFromCombatLocked is CR 506.4 for the ATTACKED side: the
// planeswalker or battle `target` has been removed from combat while
// staying on the battlefield, so every creature attacking it "continues
// to be an attacking creature, although it is not attacking any player,
// planeswalker, or battle" (CR 506.4c).
//
// Each such attacker is re-pointed at AttackingNothing. Nothing else
// about it changes: it stays announced (it attacked, CR 508.7a's
// reasoning), stays blocked or unblocked, and keeps its
// Game.attackDefenders row — which is why this writes the field
// directly rather than through setAttackTargetLocked, whose Nil-defender
// arm would drop that row. The row still names the player who was
// defending "before it was removed from combat" (CR 802.2a), so the
// block path needs no change; damage resolves the sentinel to nothing.
//
// Only ANNOUNCED attackers. A creature merely staged in the
// declare-attackers step is not in combat yet, and its declaration is
// still the active player's to change (the reselect verb refuses it
// for the same reason).
//
// Called from materialiseControlLocked (a control change) and
// phaseOutLocked. Not from removeFromCombatLocked, whose other caller
// is regeneration — CR 701.19a removes only an attacking or blocking
// CREATURE from combat. A permanent LEAVING the battlefield needs no call:
// its instance id stops resolving on its own, and a return is a new
// object (CR 400.7).
//
// Caller must hold g.mu in write mode.
func (g *Game) removeAttackedFromCombatLocked(target uuid.UUID) {
	if target == uuid.Nil || g.Battlefield == nil {
		return
	}
	changed := false
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.AttackingTarget != target || c.InstanceID == target || !g.announcedAttacks[c.InstanceID] {
			continue
		}
		c.AttackingTarget = AttackingNothing
		changed = true
	}
	if changed {
		// "Is this creature attacking player P" just changed for every
		// one of them (#1218's reasoning).
		g.invalidateLayersForAttackChangeLocked()
	}
}

// forgetAttackDefenderLocked drops one attacker's last-known defending
// player. Caller must hold g.mu.
func (g *Game) forgetAttackDefenderLocked(id uuid.UUID) {
	delete(g.attackDefenders, id)
	if len(g.attackDefenders) == 0 {
		g.attackDefenders = nil
	}
}

// defendingPlayerForAttackerLocked is the defending player of
// `attacker`'s attack FOR BLOCKING (CR 802.4a, 509.1a): the live
// defendingPlayerForAttackLocked answer while the target resolves, and
// the player recorded when the attack was pointed once it does not.
//
// The fallback is CR 506.4c plus CR 802.2a. A creature attacking a
// planeswalker or battle that has been removed from combat "continues
// to be an attacking creature … It may be blocked", and the player who
// may block it is the one it was attacking "before it was removed from
// combat" — the walker's controller or the battle's protector at the
// time. The live read cannot name that player once the permanent has
// gone; attackDefenders can.
//
// BLOCKING ONLY. Combat damage keeps reading the live target through
// dealCombatDamageToAttackTargetLocked, so an unblocked creature whose
// target left still deals its damage to nothing (CR 506.4c, 510.1b);
// and the card-side "attacking you" readers keep the live answer,
// because such a creature "is not attacking any player".
//
// Returns uuid.Nil for a creature not attacking, and when the recorded
// player has left the game (CR 800.4a).
//
// Caller must hold g.mu. Reads only.
func (g *Game) defendingPlayerForAttackerLocked(attacker *Card) uuid.UUID {
	if attacker == nil || attacker.AttackingTarget == uuid.Nil {
		return uuid.Nil
	}
	if d := g.defendingPlayerForAttackLocked(attacker.AttackingTarget); d != uuid.Nil {
		return d
	}
	d, ok := g.attackDefenders[attacker.InstanceID]
	if !ok {
		return uuid.Nil
	}
	if p := g.playerByIDLocked(d); p == nil || p.Eliminated {
		return uuid.Nil
	}
	return d
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
// Everything else about attack LEGALITY — goad, "attacks if able" —
// remains the sandbox's to arbitrate, exactly as it was for player
// attacks. The propaganda tax is no longer among them: it is a COST,
// not a legality, and the declaration verbs charge it through
// attack_tax.go (ADR 0080).
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

// DefendingPlayerForAttackerForEffect is the exported read of
// defendingPlayerForAttackerLocked: the seat whose creatures may block
// the attacking creature `attacker` — including one whose planeswalker
// or battle has left combat (CR 506.4c, #1364). The view's
// defending_player and the enumerator's block pre-filter read it, so
// they agree with the declaration verb. uuid.Nil when nobody may.
//
// For "is this creature attacking player P" use
// DefendingPlayerForAttackForEffect on its target instead: a creature
// whose target left is not attacking any player.
//
// Read-only. Caller must hold g.mu (read or write).
func (g *Game) DefendingPlayerForAttackerForEffect(attacker uuid.UUID) uuid.UUID {
	return g.defendingPlayerForAttackerLocked(findBattlefieldCard(g, attacker))
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
