package game

import (
	"testing"

	"github.com/google/uuid"
)

// turn_tally_excess_damage_test.go — #5663: the per-turn record that a
// creature or planeswalker was dealt excess damage (CR 120.10), keyed
// by who controlled it as it was dealt (Rith, Liberated Primeval's "if
// a creature or planeswalker an opponent controlled was dealt excess
// damage this turn").

// excessSourceForTest puts a 1/1 creature with the printed keywords
// under owner's control.
func excessSourceForTest(g *Game, owner uuid.UUID, keywords ...string) uuid.UUID {
	c := NewCard("Excess Source", owner)
	c.TypeLine = "Creature — Goblin"
	c.Power, c.Toughness = 1, 1
	c.Keywords = keywords
	g.Battlefield.PushTop(c)
	g.layerVersion.Add(1)
	return c.InstanceID
}

func opponentDealtExcessForTest(g *Game, you uuid.UUID) bool {
	var got bool
	g.ReadSnapshot(func() { got = g.OpponentPermanentDealtExcessDamageThisTurn(you) })
	return got
}

// Lethal damage is not excess; one point past it is, and damage already
// marked on the creature counts toward lethal (CR 120.4a, 120.6).
func TestExcessDamageToACreatureIsDamagePastLethal(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := excessSourceForTest(g, me.ID)
	exact := pushScopedTestCreature(g, opp.ID, 3, 3)
	dealNoncombatForTest(t, g, src, exact, 3)
	if opponentDealtExcessForTest(g, me.ID) {
		t.Fatal("exactly lethal damage was recorded as excess")
	}

	twice := pushScopedTestCreature(g, opp.ID, 3, 3)
	dealNoncombatForTest(t, g, src, twice, 2)
	if opponentDealtExcessForTest(g, me.ID) {
		t.Fatal("2 damage to a 3/3 was recorded as excess")
	}
	dealNoncombatForTest(t, g, src, twice, 2)
	if !opponentDealtExcessForTest(g, me.ID) {
		t.Error("2 more damage to a 3/3 with 2 marked is 1 excess, and was not recorded")
	}
	if opponentDealtExcessForTest(g, opp.ID) {
		t.Error("the creature's own controller reads it as an opponent's")
	}
}

// From a source with deathtouch any amount greater than 1 is excess
// (CR 120.4a); exactly 1 is lethal and not excess.
func TestExcessDamageFromADeathtouchSource(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := excessSourceForTest(g, me.ID, "deathtouch")
	big := pushScopedTestCreature(g, opp.ID, 5, 5)
	dealNoncombatForTest(t, g, src, big, 1)
	if opponentDealtExcessForTest(g, me.ID) {
		t.Fatal("1 deathtouch damage was recorded as excess")
	}
	other := pushScopedTestCreature(g, opp.ID, 5, 5)
	dealNoncombatForTest(t, g, src, other, 2)
	if !opponentDealtExcessForTest(g, me.ID) {
		t.Error("2 deathtouch damage to a 5/5 is 1 excess, and was not recorded")
	}
}

// A planeswalker is dealt excess damage past its loyalty (CR 120.10),
// and my own permanents are not an opponent's.
func TestExcessDamageToAPlaneswalkerIsDamagePastLoyalty(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := excessSourceForTest(g, me.ID)
	walker := pushPlaneswalkerForTest(g, opp.ID, "Test Walker", 3)
	dealNoncombatForTest(t, g, src, walker, 3)
	if opponentDealtExcessForTest(g, me.ID) {
		t.Fatal("damage equal to loyalty was recorded as excess")
	}
	second := pushPlaneswalkerForTest(g, opp.ID, "Second Walker", 3)
	dealNoncombatForTest(t, g, src, second, 4)
	if !opponentDealtExcessForTest(g, me.ID) {
		t.Error("4 damage to a planeswalker with 3 loyalty is 1 excess, and was not recorded")
	}

	mine := pushScopedTestCreature(g, me.ID, 1, 1)
	theirs := excessSourceForTest(g, opp.ID)
	g.WithWriteLock(func() { g.resetTurnTallyLocked() })
	dealNoncombatForTest(t, g, theirs, mine, 5)
	if opponentDealtExcessForTest(g, me.ID) {
		t.Error("excess damage to my own creature reads as an opponent's")
	}
	if !opponentDealtExcessForTest(g, opp.ID) {
		t.Error("my opponent does not see the excess damage my creature was dealt")
	}
}

// A battle that is neither a creature nor a planeswalker is not counted.
func TestExcessDamageToABattleIsNotCounted(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := excessSourceForTest(g, me.ID)
	battle := pushBattleForTest(g, me.ID, opp.ID, "Test Siege", 2)
	dealNoncombatForTest(t, g, src, battle, 5)
	if opponentDealtExcessForTest(g, me.ID) || opponentDealtExcessForTest(g, opp.ID) {
		t.Error("damage past a battle's defense was recorded for a creature or planeswalker")
	}
}

// "An opponent controlled": the controller is the permanent's as it was
// dealt the damage, and the record outlives the permanent.
func TestExcessDamageReadsTheControllerAsItWasDealt(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := excessSourceForTest(g, me.ID)
	victim := pushScopedTestCreature(g, opp.ID, 2, 2)
	dealNoncombatForTest(t, g, src, victim, 4)
	g.RunStateChecksForTest()
	if findBattlefieldCard(g, victim) != nil {
		t.Fatal("the 2/2 survived 4 damage")
	}
	if !opponentDealtExcessForTest(g, me.ID) {
		t.Error("the record was lost when the creature died")
	}
}

// Combat damage counts, and two blockers' damage to one attacker is
// read together (CR 120.10 "those sources together", CR 510.2): 2 and 2
// to a 0/3 is 1 excess though neither alone is lethal. The attacker has
// no power, so it has no damage of its own to assign.
func TestExcessCombatDamageIsReadAcrossSimultaneousSources(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushScopedTestCreature(g, me.ID, 0, 3)
	findBattlefieldCard(g, attacker).SummonedThisTurn = false
	b1 := pushScopedTestCreature(g, opp.ID, 2, 2)
	b2 := pushScopedTestCreature(g, opp.ID, 2, 2)
	declareAndLock(t, g, opp.ID, attacker)
	advanceIntoStep(t, g, StepDeclareBlockers)
	for _, b := range []uuid.UUID{b1, b2} {
		if err := g.DeclareBlocker(b, attacker); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	passUntilStep(t, g, StepEndCombat)
	if findBattlefieldCard(g, attacker) != nil {
		t.Fatal("the 0/3 survived 4 combat damage")
	}
	if !opponentDealtExcessForTest(g, opp.ID) {
		t.Error("2 and 2 combat damage to my 0/3 is 1 excess for my opponent's reader, and was not recorded")
	}
}

func TestExcessDamageResetsClonesAndRestores(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := excessSourceForTest(g, me.ID)
	victim := pushScopedTestCreature(g, opp.ID, 2, 2)
	dealNoncombatForTest(t, g, src, victim, 3)

	clone := g.Clone()
	g.WithWriteLock(func() { g.TurnTally.ExcessDamaged[0].Controller = me.ID })
	if !opponentDealtExcessForTest(clone, me.ID) {
		t.Error("the clone shares the record with the game it was taken from")
	}
	g.WithWriteLock(func() { g.TurnTally.ExcessDamaged[0].Controller = opp.ID })

	restored, err := throughJSON(t, g.CaptureSnapshot()).Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if !opponentDealtExcessForTest(restored, me.ID) {
		t.Error("the restored game lost the excess damage record")
	}
	if opponentDealtExcessForTest(restored, opp.ID) {
		t.Error("the restored record names the wrong controller")
	}

	g.WithWriteLock(func() { g.resetTurnTallyLocked() })
	if opponentDealtExcessForTest(g, me.ID) {
		t.Error("a new turn starts with the last turn's excess damage")
	}
}

// The intervening "if" only rises within a turn: it reads nothing while
// it holds, and is refused before any excess damage, since another item
// of the batch could deal some.
func TestExcessDamageConditionReadsNothingWhileItHolds(t *testing.T) {
	f := newCondFixture(t)
	if _, ok := f.access(t, cond(FootprintCondOpponentPermanentDealtExcessDamage, 0), drawOne); ok {
		t.Error("before any excess damage the condition is false, and the step is refused")
	}
	f.g.TurnTally.ExcessDamaged = []ExcessDamageRecord{{Controller: f.opp.ID, Creature: true}}
	a := f.mustAccess(t, cond(FootprintCondOpponentPermanentDealtExcessDamage, 0), drawOne)
	if a.boardReads || a.monarchRead {
		t.Errorf("read the board or the monarch: %+v", a)
	}
}
