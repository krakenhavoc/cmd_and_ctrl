package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// land_mana_colorless_test.go pins Pale Moon's replacement (S82 OLD b679,
// #5618): what a nonbasic land tapped for mana makes under it. The card is
// pinned in cards/effects.

func TestNonbasicLandManaColorlessRewritesOnlyANonbasicLandsTap(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	nonbasic, basic, rock, theirs := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{InstanceID: nonbasic, Name: "Karoo", TypeLine: "Land", Owner: me, Controller: me})
		g.Battlefield.PushTop(Card{InstanceID: basic, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me, Controller: me})
		g.Battlefield.PushTop(Card{InstanceID: rock, Name: "Rock", TypeLine: "Artifact", Owner: me, Controller: me})
		g.Battlefield.PushTop(Card{InstanceID: theirs, Name: "Their Land", TypeLine: "Land", Owner: opp, Controller: opp})
		if !g.NonbasicLandsProduceColorlessThisTurnForEffect(uuid.Nil, me, "Pale Moon") {
			t.Fatal("the replacement was not registered")
		}
	})
	for _, tc := range []struct {
		name    string
		player  uuid.UUID
		source  uuid.UUID
		fromTap bool
		in      []string
		want    []string
	}{
		{"a nonbasic land's two colours keep the amount", me, nonbasic, true, []string{"W", "U"}, []string{"C", "C"}},
		{"an opponent's nonbasic land", opp, theirs, true, []string{"G"}, []string{"C"}},
		{"already colourless", me, nonbasic, true, []string{"C"}, []string{"C"}},
		{"a basic land", me, basic, true, []string{"G"}, []string{"G"}},
		{"not tapped for mana", me, nonbasic, false, []string{"G"}, []string{"G"}},
		{"not a land", me, rock, true, []string{"G"}, []string{"G"}},
	} {
		var got []string
		g.ReadSnapshot(func() { got = g.producedManaPreviewLocked(tc.player, tc.source, tc.in, tc.fromTap) })
		if !slices.Equal(got, tc.want) {
			t.Errorf("%s: %v became %v, want %v", tc.name, tc.in, got, tc.want)
		}
	}
}

func TestLandSubtypeManaColorlessRewritesOnlyThatTypesTap(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	mountain, forest, theirs := uuid.New(), uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(Card{InstanceID: mountain, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me, Controller: me})
		g.Battlefield.PushTop(Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me, Controller: me})
		g.Battlefield.PushTop(Card{InstanceID: theirs, Name: "Their Mountain", TypeLine: "Basic Land — Mountain", Owner: opp, Controller: opp})
		if !g.LandSubtypeProducesColorlessThisTurnForEffect(uuid.Nil, me, "Mountain", "Chaos Moon") {
			t.Fatal("the replacement was not registered")
		}
	})
	for _, tc := range []struct {
		name    string
		player  uuid.UUID
		source  uuid.UUID
		fromTap bool
		in      []string
		want    []string
	}{
		{"a Mountain makes colourless", me, mountain, true, []string{"R"}, []string{"C"}},
		{"an opponent's Mountain", opp, theirs, true, []string{"R"}, []string{"C"}},
		{"a Forest is untouched", me, forest, true, []string{"G"}, []string{"G"}},
		{"not tapped for mana", me, mountain, false, []string{"R"}, []string{"R"}},
	} {
		var got []string
		g.ReadSnapshot(func() { got = g.producedManaPreviewLocked(tc.player, tc.source, tc.in, tc.fromTap) })
		if !slices.Equal(got, tc.want) {
			t.Errorf("%s: %v became %v, want %v", tc.name, tc.in, got, tc.want)
		}
	}
}
