package game

import (
	"testing"

	"github.com/google/uuid"
)

// multiply_tokens_test.go — ModMultiplyTokens (#5368, Kaya, Geist Hunter):
// twice as many tokens under the record's player's control this turn.

func countTokensNamed(g *Game, owner uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == owner && c.IsToken() {
			n++
		}
	}
	return n
}

func TestMultiplyTokensDoublesOnlyYourTokensUntilCleanup(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tmpl := Card{Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1}
	g.WithWriteLock(func() {
		if !g.MultiplyTokensThisTurnForEffect(uuid.Nil, me.ID, 2, "Kaya") {
			t.Fatal("no record was written")
		}
		if err := g.CreateTokenForEffect(me.ID, tmpl, 3); err != nil {
			t.Fatal(err)
		}
		if err := g.CreateTokenForEffect(opp.ID, tmpl, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := countTokensNamed(g, me.ID, "Spirit"); got != 6 {
		t.Errorf("my tokens = %d, want 6", got)
	}
	if got := countTokensNamed(g, opp.ID, "Spirit"); got != 1 {
		t.Errorf("an opponent's tokens = %d, want 1", got)
	}
	g.WithWriteLock(func() {
		g.sweepScopedEffectsLocked(true)
		if err := g.CreateTokenForEffect(me.ID, tmpl, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := countTokensNamed(g, me.ID, "Spirit"); got != 7 {
		t.Errorf("after cleanup: my tokens = %d, want 7", got)
	}
}

func TestMultiplyTokensRefusesNoPlayerAndSmallFactors(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		if g.MultiplyTokensThisTurnForEffect(uuid.Nil, uuid.Nil, 2, "x") {
			t.Error("a record with no player was written")
		}
		if g.MultiplyTokensThisTurnForEffect(uuid.Nil, g.Seats[0].ID, 1, "x") {
			t.Error("a record with factor 1 was written")
		}
	})
}
