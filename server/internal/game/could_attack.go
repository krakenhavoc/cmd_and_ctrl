package game

import (
	"slices"

	"github.com/google/uuid"
)

// could_attack.go — #6211: "except for creatures that couldn't attack"
// (Season of the Witch).
//
// CR 508.1a: the active player chooses which creatures they control will
// attack; "the chosen creatures must be untapped, they can't also be
// battles, and each one must either have haste or have been controlled by
// the active player continuously since the turn began." CR 508.1c: and no
// restriction may be disobeyed. A creature that could not have been chosen
// in any declare attackers step this turn "couldn't attack". That is a
// fact of the declaration, so it is written down as the declaration passes
// its checkpoint, beside the CR 508.1d record (had_to_attack.go): the
// creatures declared as attackers, and every other creature the active
// player controls that AttackerEligible (the one copy of the CR 508.1a/c
// test that the verb and the enumerator share) would have let them
// declare. A non-attacker's tapped state and restrictions at the
// checkpoint are what they were as the step began, since declaring
// attackers changes nothing about it.
//
// Per turn and per OBJECT (TurnTally.CouldAttack): a creature that enters
// after combat, changes hands after the declaration, or leaves and comes
// back (CR 400.7) could not attack this turn as that object. A turn with
// two combats keeps both declarations' creatures.

// recordCouldAttackLocked adds this declaration's creatures to the turn's
// record.
//
// Caller must hold g.mu in write mode.
func (g *Game) recordCouldAttackLocked() {
	g.RecomputeLayersIfStaleLocked()
	ap := g.activePlayerIDLocked()
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != ap || !c.IsCreature() {
			continue
		}
		if c.AttackingTarget == uuid.Nil && !AttackerEligible(c, ap) {
			continue
		}
		ref := ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}
		if !slices.Contains(g.TurnTally.CouldAttack, ref) {
			g.TurnTally.CouldAttack = append(g.TurnTally.CouldAttack, ref)
		}
	}
}

// CouldHaveAttackedThisTurnForEffect reports whether the creature `cardID`
// names, as the object it is now, could have been declared as an attacker
// in a declare attackers step this turn (#6211, CR 508.1a, 508.1c). False
// for a creature that was tapped, summoning sick or under a "can't attack"
// restriction at every declaration this turn, for one that was not its
// controller's as the active player then, and on a turn with no declare
// attackers step.
//
// Caller must hold g.mu.
func (g *Game) CouldHaveAttackedThisTurnForEffect(cardID uuid.UUID) bool {
	if cardID == uuid.Nil || len(g.TurnTally.CouldAttack) == 0 {
		return false
	}
	return slices.Contains(g.TurnTally.CouldAttack, ObjectRef{ID: cardID, Epoch: g.objectEpochLocked(cardID)})
}
