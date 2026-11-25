package game

import (
	"testing"
)

// permanent_query_without_keyword_test.go — PermanentQuery.WithoutKeyword,
// the "without flying" of Stoneshock Giant (S82 older sets batch 271).

func TestPermanentQueryWithoutKeywordRejectsAnyPermanentWithTheKeyword(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	p := g.Seats[1].ID
	ground := pushTypedTestCard(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: p, Controller: p})
	air := pushTypedTestCard(g, Card{Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: p, Controller: p})
	q := PermanentQuery{Types: []string{"creature"}, WithoutKeyword: "flying"}
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if !q.matchesLocked(g, findBattlefieldCard(g, ground)) {
			t.Error("a creature without flying did not match")
		}
		if q.matchesLocked(g, findBattlefieldCard(g, air)) {
			t.Error("a creature with flying matched a query without flying")
		}
	})
}

func TestPermanentQueryWithoutKeywordIsInEqualNounCloneAndCharacteristic(t *testing.T) {
	a := PermanentQuery{Types: []string{"creature"}, WithoutKeyword: "flying"}
	if !a.Equal(PermanentQuery{Types: []string{"creature"}, WithoutKeyword: "flying"}) {
		t.Error("two equal queries are not equal")
	}
	if a.Equal(PermanentQuery{Types: []string{"creature"}}) {
		t.Error("a query without flying equals an unrestricted one")
	}
	if got := clonePermanentQueries([]PermanentQuery{a}); len(got) != 1 || !got[0].Equal(a) {
		t.Errorf("clonePermanentQueries dropped WithoutKeyword: %+v", got)
	}
	if got, want := a.Noun(), "creature without flying"; got != want {
		t.Errorf("noun = %q, want %q", got, want)
	}
	if !a.matchesCharacteristic(&Characteristic{Types: []string{"creature"}}) {
		t.Error("a snapshot without flying did not match")
	}
	if a.matchesCharacteristic(&Characteristic{Types: []string{"creature"}, Abilities: []string{"flying"}}) {
		t.Error("a snapshot with flying matched")
	}
}
