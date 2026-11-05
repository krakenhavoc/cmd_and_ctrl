package game

import "github.com/google/uuid"

// put_blocking.go — a permanent put onto the battlefield BLOCKING an
// attacking creature without being declared a blocker (CR 506.3d,
// 509.3c, 509.4; #3172, #4956). The twin of stampEntryAttackerLocked.

// stampEntryBlockerLocked makes a creature that has just landed on the
// battlefield a blocker of `attacker` without declaring it one:
//
//   - stamp the block (Card.BlockingTarget), which is what the combat
//     damage steps build each attacker's blockers from;
//   - record the pair as announced (announcedBlocks): it was never
//     DECLARED as a blocker (CR 509.4), so the block lock-in must not
//     emit an EventBlock for it, and no "whenever this creature
//     blocks" trigger fires;
//   - mark the attacker blocked (blockedAttackers, CR 509.1h), and, if
//     the attacker was an UNBLOCKED creature at that time, emit its
//     one EventBecomesBlocked (#4956): CR 509.3c says a "becomes
//     blocked" ability triggers when a creature put onto the
//     battlefield as a blocker blocks an attacker that was unblocked
//     at that time. An attacker that was already blocked (Brimaz's,
//     blocked by Brimaz itself) does not become blocked a second time,
//     and one that was neither blocked nor unblocked yet (its
//     defending player has not declared blockers) does not "become
//     blocked" either; both are marked blocked silently. The event has
//     the lock-in's shape (commitBlockDeclarationLocked): the attacker
//     in Source / CardID / Target, and the blocker's controller, the
//     defending player, as Actor.
//
// An attacker that is gone, or no longer attacking, leaves the creature
// on the battlefield not blocking, and a creature that is not a creature
// does not block at all: the effect that asked for it has already
// resolved and the body is the part of it that can still be delivered
// (stampEntryAttackerLocked's posture).
//
// Caller must hold g.mu.
func (g *Game) stampEntryBlockerLocked(id, attacker uuid.UUID) {
	if id == uuid.Nil || attacker == uuid.Nil {
		return
	}
	c := findBattlefieldCard(g, id)
	if c == nil {
		return
	}
	a := findBattlefieldCard(g, attacker)
	if a == nil || a.AttackingTarget == uuid.Nil || !c.IsCreature() {
		c.clearBlocking()
		return
	}
	// Asked BEFORE the stamp: once this creature is blocking it, the
	// attacker no longer reads as unblocked.
	becomesBlocked := g.UnblockedAttackerForEffect(attacker)
	c.setBlockingSet([]uuid.UUID{attacker})
	if g.announcedBlocks == nil {
		g.announcedBlocks = map[uuid.UUID][]uuid.UUID{}
	}
	g.announcedBlocks[id] = []uuid.UUID{attacker}
	if g.blockedAttackers == nil {
		g.blockedAttackers = map[uuid.UUID]bool{}
	}
	g.blockedAttackers[attacker] = true
	if becomesBlocked {
		g.EmitEvent(Event{
			Kind:   EventBecomesBlocked,
			Actor:  c.Controller,
			Source: attacker,
			CardID: attacker,
			Target: attacker,
		})
	}
}
