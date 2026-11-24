package game

import "github.com/google/uuid"

// had_to_attack.go — #6098: "if that creature had to attack this combat"
// (Firkraag, Cunning Instigator).
//
// CR 508.1d: "The active player checks each creature they control to see
// whether it's affected by any requirements (effects that say a creature
// attacks if able, or that it attacks if some condition is met)." A
// creature had to attack when it was affected by one or more of those
// requirements as attackers were declared: goaded (CR 701.15b), "attacks
// each combat if able", and the rest of attack_requirements.go's shapes.
//
// Reading it later off the creature is wrong both ways. A goad can wear off
// or be added between the declaration and combat damage, and a creature
// that was goaded only after it attacked never had to. So the answer is
// written down once, as the declaration passes its checkpoint
// (markAttackDeclarationJudgedLocked), from the same requirement list the
// checkpoint judged (attackRequirementsOfLocked with the final
// declaration). It is cleared with the rest of combat
// (clearAttackAnnouncementsLocked), and an object that leaves the
// battlefield is a new object that never had to attack (CR 400.7,
// forgetPerObjectTurnStateLocked). Being removed from combat (CR 506.4)
// does not erase it: the creature still had to attack this combat.
//
// Data, not a func: a set of instance IDs on Game, carried by Clone,
// RestoreFrom and the persisted snapshot (Snapshot.HadToAttack).

// recordHadToAttackLocked writes this combat's record: every creature the
// active player controls that carries at least one attack requirement
// under the declaration as it now stands.
//
// Caller must hold g.mu in write mode.
func (g *Game) recordHadToAttackLocked() {
	g.RecomputeLayersIfStaleLocked()
	ap := g.activePlayerIDLocked()
	assign := g.currentAttackAssignmentLocked()
	var rec map[uuid.UUID]bool
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != ap || !c.IsCreature() {
			continue
		}
		if len(g.attackRequirementsOfLocked(c, assign)) == 0 {
			continue
		}
		if rec == nil {
			rec = map[uuid.UUID]bool{}
		}
		rec[c.InstanceID] = true
	}
	g.hadToAttack = rec
}

// forgetHadToAttackLocked drops one object from the record (CR 400.7).
//
// Caller must hold g.mu in write mode.
func (g *Game) forgetHadToAttackLocked(cardID uuid.UUID) {
	delete(g.hadToAttack, cardID)
	if len(g.hadToAttack) == 0 {
		g.hadToAttack = nil
	}
}

// HadToAttackThisCombatForEffect reports whether the creature `cardID`
// had to attack in the current combat (#6098, CR 508.1d): it carried an
// attack requirement as the attack declaration passed its checkpoint, and
// it has not left the battlefield since. False outside combat and before
// the declaration is over.
//
// Caller must hold g.mu.
func (g *Game) HadToAttackThisCombatForEffect(cardID uuid.UUID) bool {
	return g.hadToAttack[cardID]
}
