package game

import "github.com/google/uuid"

// exit_alone.go — #6098: "if it was attacking or blocking alone"
// (Thijarian Witness) on a leaving permanent's EventLTB.
//
// CR 506.5: "A creature is attacking alone if it's attacking but no other
// creatures are. ... A creature is blocking alone if it's blocking but no
// other creatures are." A leaves-the-battlefield ability looks back in time
// (CR 603.10a), so the question is asked of the moment before the event.
//
// Asking the live battlefield as each card leaves is wrong for an event
// that removes several permanents at once. Two attackers that trade with
// their blockers in one combat damage step die together (CR 704.3): the
// first to be moved still sees the second attacking, but the second, moved
// after, would find nobody else and read as alone. Neither was. So the
// other combatants are counted on the live battlefield AND among the
// pre-move copies of the open simultaneous exit (simultaneous.go) that
// have already left: those were still in combat when the event began.

// otherCombatantLocked reports whether a creature other than cardID was
// attacking (attacking true) or blocking (attacking false) at the start
// of the event cardID is leaving in.
//
// Caller must hold g.mu.
func (g *Game) otherCombatantLocked(cardID uuid.UUID, attacking bool) bool {
	inCombat := func(c *Card) bool {
		if c.InstanceID == cardID || !c.IsCreature() {
			return false
		}
		if attacking {
			return c.AttackingTarget != uuid.Nil
		}
		return c.BlockingTarget != uuid.Nil
	}
	for i := range g.Battlefield.Cards {
		if inCombat(&g.Battlefield.Cards[i]) {
			return true
		}
	}
	for i := range g.simultaneousExit {
		c := &g.simultaneousExit[i]
		// A batch member still on the battlefield was counted above,
		// as it is now; one that has left is read as the batch caught it.
		if findCardOnBattlefield(g, c.InstanceID) >= 0 {
			continue
		}
		if inCombat(c) {
			return true
		}
	}
	return false
}
