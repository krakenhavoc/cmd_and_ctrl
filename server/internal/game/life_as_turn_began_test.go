package game

import "testing"

// life_as_turn_began_test.go pins #6164: every player's life total as the
// turn began is recorded as the turn begins, survives a snapshot taken
// mid-turn, and rides the elimination event.

func TestLifeAsTurnBeganIsRecordedAsTheTurnBegins(t *testing.T) {
	g := newActiveGame(t)
	opp := g.Seats[1]
	g.WithWriteLock(func() { opp.Life = 12 })
	start := g.Turn.Seq
	for g.Turn.Seq == start {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	g.WithWriteLock(func() { opp.Life = 5 })
	var got int
	g.ReadSnapshot(func() { got, _ = g.LifeAsTurnBeganForEffect(opp.ID) })
	if got != 12 {
		t.Fatalf("life as the turn began = %d, want 12", got)
	}
	// A snapshot taken mid-turn keeps the recorded value.
	restored := restoreStrictForTest(t, g)
	restored.ReadSnapshot(func() { got, _ = restored.LifeAsTurnBeganForEffect(opp.ID) })
	if got != 12 {
		t.Errorf("after a restore: life as the turn began = %d, want 12", got)
	}
	// The next turn records the life it begins with.
	seq := g.Turn.Seq
	for g.Turn.Seq == seq {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	g.ReadSnapshot(func() { got, _ = g.LifeAsTurnBeganForEffect(opp.ID) })
	if got != 5 {
		t.Errorf("next turn: life as the turn began = %d, want 5", got)
	}
}

// With no record for this turn (a file from before the field), the reader
// falls back to the current life.
func TestLifeAsTurnBeganFallsBackToTheCurrentLife(t *testing.T) {
	g := newActiveGame(t)
	opp := g.Seats[1]
	g.WithWriteLock(func() {
		opp.LifeAsTurnBeganSeq = 0
		opp.Life = 33
	})
	var got int
	var ok bool
	g.ReadSnapshot(func() { got, ok = g.LifeAsTurnBeganForEffect(opp.ID) })
	if !ok || got != 33 {
		t.Errorf("fallback = %d, %v, want 33", got, ok)
	}
}
