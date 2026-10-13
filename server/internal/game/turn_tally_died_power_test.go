package game

import (
	"testing"

	"github.com/google/uuid"
)

// turn_tally_died_power_test.go — #3143. TurnTally.DiedPowerBySubtype is
// the table-wide total power of the creatures of a type that died this
// turn (Genesis of the Daleks), read as each one last existed.

func diedPowerPush(g *Game, owner uuid.UUID, name, typeLine string, power int, counters map[string]int, keywords ...string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{
			InstanceID: id, Name: name, TypeLine: typeLine, Power: power, Toughness: power,
			Counters: counters, Keywords: keywords, Owner: owner, Controller: owner,
		})
	})
	return id
}

func diedPower(g *Game, subtype string) int {
	var n int
	g.WithWriteLock(func() { n = g.DiedPowerWithSubtypeThisTurn(subtype) })
	return n
}

func diedPowerDestroy(t *testing.T, g *Game, ids ...uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		for _, id := range ids {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatal(err)
			}
		}
	})
}

func TestDiedPowerTotalsTheDalekPowerThatDiedThisTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	mine := diedPowerPush(g, me, "Dalek", "Token Artifact Creature — Dalek", 3, nil)
	theirs := diedPowerPush(g, opp, "Dalek Drone", "Artifact Creature — Dalek", 2, nil)
	bear := diedPowerPush(g, me, "Bear", "Creature — Bear", 5, nil)
	survivor := diedPowerPush(g, me, "Dalek", "Token Artifact Creature — Dalek", 3, nil)

	diedPowerDestroy(t, g, mine, theirs, bear)

	if got := diedPower(g, "Dalek"); got != 5 {
		t.Errorf("Dalek power that died = %d, want 5 (3 + 2: any controller, a token counts, a Bear and a survivor do not)", got)
	}
	if got := diedPower(g, "dalek"); got != 5 {
		t.Errorf("the type is case-insensitive: got %d, want 5", got)
	}
	if got := diedPower(g, "Bear"); got != 5 {
		t.Errorf("Bear power that died = %d, want 5", got)
	}
	if got := diedPower(g, "Human"); got != 0 {
		t.Errorf("Human power that died = %d, want 0", got)
	}
	_ = survivor
}

func TestDiedPowerIsThePowerItHadWithCounters(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	d := diedPowerPush(g, me, "Dalek", "Artifact Creature — Dalek", 3, map[string]int{"+1/+1": 2})

	diedPowerDestroy(t, g, d)

	if got := diedPower(g, "Dalek"); got != 5 {
		t.Errorf("a 3/3 with two +1/+1 counters died as a 5/5: got %d, want 5", got)
	}
}

func TestDiedPowerCountsAChangelingForEveryCreatureTypeButNotALandType(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	shifter := diedPowerPush(g, me, "Shifter", "Creature — Shapeshifter", 2, nil, KeywordChangeling)

	diedPowerDestroy(t, g, shifter)

	if got := diedPower(g, "Dalek"); got != 2 {
		t.Errorf("a changeling is a Dalek: got %d, want 2", got)
	}
	if got := diedPower(g, "Shapeshifter"); got != 2 {
		t.Errorf("a changeling counted once under its own type: got %d, want 2", got)
	}
	if got := diedPower(g, "Forest"); got != 0 {
		t.Errorf("Forest is not a creature type: got %d, want 0", got)
	}
}

func TestDiedPowerIgnoresAnythingThatDidNotDie(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	exiled := diedPowerPush(g, me, "Dalek", "Artifact Creature — Dalek", 3, nil)
	bounced := diedPowerPush(g, me, "Dalek", "Artifact Creature — Dalek", 3, nil)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(exiled); err != nil {
			t.Fatal(err)
		}
		if err := g.BounceToHandForEffect(bounced); err != nil {
			t.Fatal(err)
		}
	})
	if got := diedPower(g, "Dalek"); got != 0 {
		t.Errorf("an exile and a bounce are not deaths: got %d, want 0", got)
	}
}

func TestDiedPowerIsFlooredAtZeroAndStartsOverEachTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	weak := diedPowerPush(g, me, "Dalek", "Artifact Creature — Dalek", 1, map[string]int{"-1/-1": 3})
	diedPowerDestroy(t, g, weak)
	if got := diedPower(g, "Dalek"); got != 0 {
		t.Errorf("a negative total reads as zero: got %d", got)
	}

	d := diedPowerPush(g, me, "Dalek", "Artifact Creature — Dalek", 4, nil)
	diedPowerDestroy(t, g, d)
	if got := diedPower(g, "Dalek"); got != 2 {
		t.Errorf("4 + (1-3) = 2: got %d", got)
	}
	g.WithWriteLock(func() { g.resetTurnTallyLocked() })
	if got := diedPower(g, "Dalek"); got != 0 {
		t.Errorf("the tally starts over each turn: got %d", got)
	}
}

func TestDiedPowerSurvivesACloneAndASnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0].ID
	d := diedPowerPush(g, me, "Dalek", "Artifact Creature — Dalek", 3, nil)
	diedPowerDestroy(t, g, d)

	clone := g.Clone()
	if got := diedPower(clone, "Dalek"); got != 3 {
		t.Errorf("clone: got %d, want 3", got)
	}
	// The clone does not share the map with the original.
	d2 := diedPowerPush(clone, me, "Dalek", "Artifact Creature — Dalek", 2, nil)
	diedPowerDestroy(t, clone, d2)
	if got := diedPower(g, "Dalek"); got != 3 {
		t.Errorf("the original saw the clone's death: got %d, want 3", got)
	}
}
