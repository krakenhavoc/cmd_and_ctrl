package game

import "github.com/google/uuid"

// put_blocking.go — a permanent put onto the battlefield BLOCKING an
// attacking creature without being declared a blocker (CR 506.3d, 509.5;
// #3172). The twin of stampEntryAttackerLocked.

// stampEntryBlockerLocked makes a creature that has just landed on the
// battlefield a blocker of `attacker` without declaring it one:
//
//   - stamp the block (Card.BlockingTarget), which is what the combat
//     damage steps build each attacker's blockers from;
//   - record the pair as announced (announcedBlocks), which is the
//     rule itself — it was never DECLARED as a blocker, so the block
//     lock-in must not emit an EventBlock for it, and no "whenever this
//     creature blocks" trigger fires;
//   - mark the attacker blocked (blockedAttackers, CR 509.1h) without
//     an EventBecomesBlocked: the attacker was already blocked by the
//     creature that made this one (Brimaz), or, if it was not, the
//     blocking token does not make it "become blocked" (CR 509.5).
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
	c.setBlockingSet([]uuid.UUID{attacker})
	if g.announcedBlocks == nil {
		g.announcedBlocks = map[uuid.UUID][]uuid.UUID{}
	}
	g.announcedBlocks[id] = []uuid.UUID{attacker}
	if g.blockedAttackers == nil {
		g.blockedAttackers = map[uuid.UUID]bool{}
	}
	g.blockedAttackers[attacker] = true
}
