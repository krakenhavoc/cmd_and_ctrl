package game

import "github.com/google/uuid"

// block_unblocked_attackers.go — "all unblocked creatures attacking you
// become blocked by this creature" (Noble Ox, #5948). An effect makes an
// existing creature block attackers without declaring it a blocker (CR
// 509.1h: an effect can say a creature becomes blocked; CR 509.3c: such a
// creature's "becomes blocked" abilities trigger when it was an unblocked
// creature at that time). stampEntryBlockerLocked's posture for the many:
// the block is stamped and recorded as announced (no EventBlock, so no
// "whenever this creature blocks" ability fires, CR 509.4), the attacker is
// marked blocked, and one EventBecomesBlocked is emitted per attacker that
// was unblocked.

// BlockUnblockedAttackersForEffect makes `blocker` block every unblocked
// creature attacking `defender` (the player), and returns how many it
// blocks as a result. Nothing happens before blockers are declared (an
// attacker is neither blocked nor unblocked until its defending player
// has declared, #1279), for a blocker that is not a creature on the
// battlefield, or for a blocker not controlled by `defender` (CR 509.4a's
// posture: only a defending player's creature blocks). Caller must hold
// g.mu.
func (g *Game) BlockUnblockedAttackersForEffect(blocker, defender uuid.UUID) int {
	c := findBattlefieldCard(g, blocker)
	if c == nil || !c.IsCreature() || c.Controller != defender {
		return 0
	}
	var attackers []uuid.UUID
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if a.AttackingTarget != defender || c.IsBlockingAttacker(a.InstanceID) {
			continue
		}
		if g.UnblockedAttackerForEffect(a.InstanceID) {
			attackers = append(attackers, a.InstanceID)
		}
	}
	if len(attackers) == 0 {
		return 0
	}
	set := append(c.BlockedAttackers(), attackers...)
	c.setBlockingSet(set)
	if g.announcedBlocks == nil {
		g.announcedBlocks = map[uuid.UUID][]uuid.UUID{}
	}
	g.announcedBlocks[blocker] = append([]uuid.UUID(nil), set...)
	if g.blockedAttackers == nil {
		g.blockedAttackers = map[uuid.UUID]bool{}
	}
	for _, a := range attackers {
		g.blockedAttackers[a] = true
	}
	for _, a := range attackers {
		g.EmitEvent(Event{
			Kind:   EventBecomesBlocked,
			Actor:  c.Controller,
			Source: a,
			CardID: a,
			Target: a,
		})
	}
	return len(attackers)
}
