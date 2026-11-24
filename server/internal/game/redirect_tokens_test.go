package game

import (
	"testing"

	"github.com/google/uuid"
)

// redirect_tokens_test.go — ModRedirectTokens (Crafty Cutpurse): tokens an
// opponent would create this turn are created under the record's player's
// control instead.

func TestRedirectTokensTakesOnlyOpponentsTokensUntilCleanup(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tmpl := Card{Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1}
	g.WithWriteLock(func() {
		if !g.RedirectOpponentTokensThisTurnForEffect(uuid.Nil, me.ID, "Crafty Cutpurse") {
			t.Fatal("no record was written")
		}
		if err := g.CreateTokenForEffect(me.ID, tmpl, 1); err != nil {
			t.Fatal(err)
		}
		if err := g.CreateTokenForEffect(opp.ID, tmpl, 2); err != nil {
			t.Fatal(err)
		}
	})
	if got := countTokensNamed(g, me.ID, "Spirit"); got != 3 {
		t.Errorf("my tokens = %d, want 3", got)
	}
	if got := countTokensNamed(g, opp.ID, "Spirit"); got != 0 {
		t.Errorf("the opponent's tokens = %d, want 0", got)
	}
	g.WithWriteLock(func() {
		g.sweepScopedEffectsLocked(true)
		if err := g.CreateTokenForEffect(opp.ID, tmpl, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := countTokensNamed(g, opp.ID, "Spirit"); got != 1 {
		t.Errorf("after cleanup: the opponent's tokens = %d, want 1", got)
	}
}

func TestRedirectTokensRefusesNoPlayer(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		if g.RedirectOpponentTokensThisTurnForEffect(uuid.Nil, uuid.Nil, "x") {
			t.Error("a record with no player was written")
		}
	})
}
