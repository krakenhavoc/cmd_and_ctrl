package game

import "github.com/google/uuid"

// become_blocked.go — an EFFECT that says an attacking creature becomes
// blocked, with no creature blocking it (CR 509.1h, #5839). Fog Patch's
// "Attacking creatures become blocked." The sibling of
// stampEntryBlockerLocked (put_blocking.go), which blocks an attacker
// with a creature put onto the battlefield blocking; this blocks it
// with nothing at all.
//
// THE STATE. CR 509.1h: an attacker's blocked-or-unblocked status
// "remains unchanged until the creature is removed from combat, an
// effect says that it becomes blocked or unblocked, or the combat phase
// ends", and "a creature remains blocked even if all the creatures
// blocking it are removed from combat". That status is
// Game.blockedAttackers, the record the block lock-in writes and both
// combat damage steps read, so marking the attacker there is the whole
// of "becomes blocked". No new state, and nothing new in the snapshot:
// the record already rides it (GameSnapshot.BlockedAttackers), so the
// mark survives a restore and a rewind like any declared block.
//
// THE DAMAGE. Nothing changes in the combat damage steps, which already
// treat a blocked attacker with no creature blocking it as CR 510.1c
// says: it assigns no combat damage, unless it has trample, which
// assigns all of it to the player, planeswalker or battle it is
// attacking (CR 702.19d). assignAndDealCombatDamageLocked reads the
// record, never the live blockers (#715).
//
// THE TRIGGERS. CR 509.3c: "whenever this creature becomes blocked"
// triggers when an effect blocks it, "but only if the attacking
// creature was an unblocked creature at that time". So an attacker
// that was unblocked gets its one EventBecomesBlocked, in the lock-in's
// shape; one that was already blocked does not become blocked a second
// time. CR 509.3d: "whenever this creature becomes blocked by a
// creature" does NOT trigger ("It won't trigger if the creature becomes
// blocked by an effect rather than a creature"), and since no creature
// is blocking, no EventBlock is emitted — the event those abilities and
// "whenever this creature blocks" watch.
//
// BEFORE THE DECLARATION. An attacker whose defending player has not
// completed their declaration is neither blocked nor unblocked, so it
// does not "become blocked" from being unblocked (CR 509.3c); it is
// marked blocked silently, exactly as stampEntryBlockerLocked marks an
// attacker in the same position. Outside the declare-blockers step and
// the combat steps after it there is no declaration at all, and the
// call does nothing. A catalog card reaches it only after the
// declaration: the declare-blockers step grants nobody priority until
// every defender has declared (#1501, block_completion.go).

// BecomeBlockedForEffect makes each of `attackers` that is attacking a
// BLOCKED creature with no creature blocking it (CR 509.1h): Fog
// Patch's "Attacking creatures become blocked". Each one that was an
// UNBLOCKED creature at that time becomes blocked and gets one
// EventBecomesBlocked (CR 509.3c), Actor its defending player; one
// already blocked is left as it is. A permanent that is gone or not
// attacking is skipped, and so is every one when the cursor is not at
// the declare-blockers step or a later combat step. Returns how many
// attackers it marked blocked.
//
// Caller must hold g.mu in write mode.
func (g *Game) BecomeBlockedForEffect(attackers []uuid.UUID) int {
	if !g.blockersDeclaredStepLocked() {
		return 0
	}
	marked := 0
	for _, id := range attackers {
		if id == uuid.Nil || g.attackerBlockedLocked(id) {
			continue
		}
		c := findBattlefieldCard(g, id)
		if c == nil || c.AttackingTarget == uuid.Nil {
			continue
		}
		// Asked BEFORE the mark, as stampEntryBlockerLocked asks it: once
		// the record names it, the attacker no longer reads as unblocked.
		becomesBlocked := g.UnblockedAttackerForEffect(id)
		if g.blockedAttackers == nil {
			g.blockedAttackers = map[uuid.UUID]bool{}
		}
		g.blockedAttackers[id] = true
		marked++
		if becomesBlocked {
			g.EmitEvent(Event{
				Kind:   EventBecomesBlocked,
				Actor:  g.defendingPlayerForAttackerLocked(c),
				Source: id,
				CardID: id,
				Target: id,
			})
		}
	}
	if marked > 0 {
		// A static that reads block status ("as long as it's blocked")
		// has a new answer, and no EventBlock carries it.
		g.invalidateLayersForAttackChangeLocked()
	}
	return marked
}
