package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const derelictAtticOracle = "993b7b94-ed06-422d-9c7e-52a74ce9d045"

// TestDerelictAtticDrawsTwoAndLosesTwoOnUnlock casts Widow's Walk first
// (its door is the one that unlocks on entry), then unlocks the Attic.
func TestDerelictAtticDrawsTwoAndLosesTwoOnUnlock(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	room := roomCardC(me.ID, derelictAtticOracle, "Derelict Attic", "{2}{B}", "Widow's Walk", "{3}{B}", "B")
	castRoomC(t, g, me.ID, room, 1)
	settleOrdering(t, g, me.ID)
	hand, life := me.Hand.Size(), me.Life
	unlockDoorC(t, g, me.ID, room.InstanceID, game.DoorLeft)
	settleOrdering(t, g, me.ID)
	if me.Hand.Size() != hand+2 || me.Life != life-2 {
		t.Errorf("hand %d→%d, life %d→%d; want +2 cards and -2 life", hand, me.Hand.Size(), life, me.Life)
	}
}

// TestWidowsWalkPumpsALoneAttackerOnly: one attacker gets +1/+0 and
// deathtouch; two attackers get nothing (CR 508.3); and with only the
// Attic unlocked a lone attacker gets nothing either.
func TestWidowsWalkPumpsALoneAttackerOnly(t *testing.T) {
	for _, tc := range []struct {
		name      string
		face      int
		attackers int
		pumped    bool
	}{
		{"lone attacker, Widow's Walk unlocked", 1, 1, true},
		{"two attackers", 1, 2, false},
		{"lone attacker, Widow's Walk still locked", 0, 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceToMain(t, g)
			room := roomCardC(me.ID, derelictAtticOracle, "Derelict Attic", "{2}{B}", "Widow's Walk", "{3}{B}", "B")
			castRoomC(t, g, me.ID, room, tc.face)
			settleOrdering(t, g, me.ID)
			bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			if tc.attackers == 2 {
				other := b16Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
				declareAttack(t, g, opp.ID, bear, other)
			} else {
				declareAttack(t, g, opp.ID, bear)
			}
			passPriorityAroundTable(t, g)
			p, tough := sizeOfC(t, g, bear)
			dt := containsString(effectiveAbilities(t, g, bear), "deathtouch")
			if tc.pumped && (p != 3 || tough != 2 || !dt) {
				t.Errorf("lone attacker is %d/%d deathtouch=%v, want 3/2 with deathtouch", p, tough, dt)
			}
			if !tc.pumped && (p != 2 || dt) {
				t.Errorf("attacker is %d/%d deathtouch=%v, want an unpumped 2/2", p, tough, dt)
			}
		})
	}
}
