package game

import (
	"testing"

	"github.com/google/uuid"
)

// turn_tally_noncombat_test.go — #2662: the per-turn record of
// noncombat damage amounts, keyed by the source's colour and controller
// as it dealt the damage (Temple of Power's "red sources you controlled
// dealt 4 or more noncombat damage this turn").

// colouredSourceForTest puts a 1/1 creature of the given colours under
// owner's control.
func colouredSourceForTest(g *Game, owner uuid.UUID, colors ...string) uuid.UUID {
	c := NewCard("Pinger", owner)
	c.TypeLine = "Creature — Goblin"
	c.Power, c.Toughness = 1, 1
	c.Colors = colors
	g.Battlefield.PushTop(c)
	g.layerVersion.Add(1)
	return c.InstanceID
}

func dealNoncombatForTest(t *testing.T, g *Game, source, target uuid.UUID, amount int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		if g.playerByIDLocked(target) != nil {
			err = g.DealDamageToPlayerForEffect(source, target, amount)
		} else {
			err = g.DealDamageToCreatureForEffect(source, target, amount)
		}
	})
	if err != nil {
		t.Fatalf("deal %d damage: %v", amount, err)
	}
}

func noncombatForTest(g *Game, controller uuid.UUID, color string) int {
	n := 0
	g.ReadSnapshot(func() { n = g.NoncombatDamageThisTurnForEffect(controller, color) })
	return n
}

func TestNoncombatDamageCountsBySourceColourAndController(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := colouredSourceForTest(g, me.ID, "R")
	green := colouredSourceForTest(g, me.ID, "G")
	theirRed := colouredSourceForTest(g, opp.ID, "R")
	victim := pushScopedTestCreature(g, opp.ID, 5, 5)

	dealNoncombatForTest(t, g, red, opp.ID, 3)
	dealNoncombatForTest(t, g, red, victim, 1) // damage to a permanent counts too
	dealNoncombatForTest(t, g, green, opp.ID, 2)
	dealNoncombatForTest(t, g, theirRed, me.ID, 5)

	if got := noncombatForTest(g, me.ID, "R"); got != 4 {
		t.Errorf("my red sources dealt %d noncombat damage, want 4", got)
	}
	if got := noncombatForTest(g, me.ID, ""); got != 6 {
		t.Errorf("all my sources dealt %d noncombat damage, want 6", got)
	}
	if got := noncombatForTest(g, opp.ID, "R"); got != 5 {
		t.Errorf("their red sources dealt %d, want 5", got)
	}
}

// "Red sources you CONTROLLED": the controller and colour are the
// source's as it dealt the damage, not as it is now.
func TestNoncombatDamageReadsTheSourceAsItDealtIt(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := colouredSourceForTest(g, me.ID, "R")
	dealNoncombatForTest(t, g, red, opp.ID, 2)

	stealForTest(t, g, red, opp.ID)
	if got := noncombatForTest(g, me.ID, "R"); got != 2 {
		t.Errorf("after the source changed hands I have %d, want 2", got)
	}
	if got := noncombatForTest(g, opp.ID, "R"); got != 0 {
		t.Errorf("the new controller has %d for damage dealt before, want 0", got)
	}
	dealNoncombatForTest(t, g, red, me.ID, 1)
	if got := noncombatForTest(g, opp.ID, "R"); got != 1 {
		t.Errorf("the new controller has %d, want 1", got)
	}
}

func TestCombatDamageIsNotNoncombat(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := colouredSourceForTest(g, me.ID, "R")
	findBattlefieldCard(g, attacker).SummonedThisTurn = false
	declareAndLock(t, g, opp.ID, attacker)
	passUntilStep(t, g, StepEndCombat)
	if got := opp.Life; got != StartingLife-1 {
		t.Fatalf("defender at %d life, want %d: the attack did not connect", got, StartingLife-1)
	}
	if got := noncombatForTest(g, me.ID, "R"); got != 0 {
		t.Errorf("combat damage counted as %d noncombat", got)
	}
}

func TestNoncombatDamageResetsClonesAndRestores(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := colouredSourceForTest(g, me.ID, "R")
	dealNoncombatForTest(t, g, red, opp.ID, 4)

	clone := g.Clone()
	g.WithWriteLock(func() { g.TurnTally.NoncombatDamage[0].Colors[0] = "G" })
	if got := noncombatForTest(clone, me.ID, "R"); got != 4 {
		t.Errorf("the clone shares the record's colours: %d, want 4", got)
	}
	g.WithWriteLock(func() { g.TurnTally.NoncombatDamage[0].Colors[0] = "R" })

	restored, err := throughJSON(t, g.CaptureSnapshot()).Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if got := noncombatForTest(restored, me.ID, "R"); got != 4 {
		t.Errorf("restored game has %d, want 4", got)
	}

	g.WithWriteLock(func() { g.resetTurnTallyLocked() })
	if got := noncombatForTest(g, me.ID, "R"); got != 0 {
		t.Errorf("a new turn starts with %d, want 0", got)
	}
}
