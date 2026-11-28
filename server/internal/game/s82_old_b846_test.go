package game

import (
	"slices"
	"testing"
)

// TestNotesWithPrefixAreReadAndTakenBack: Seraph's notes of what it brought
// back are read with the prefix cut off, and taken back together, leaving the
// other notes alone.
func TestNotesWithPrefixAreReadAndTakenBack(t *testing.T) {
	g := newActiveGame(t)
	id := pushBattlefieldForTest(g, g.Seats[0].ID, "Seraph", "Creature — Angel", "")
	g.WithWriteLock(func() {
		g.AddNoteForEffect(id, "other")
		g.AddNoteForEffect(id, "seraph:a:1")
		g.AddNoteForEffect(id, "seraph:b:2")
		if got := g.NotesWithPrefixForEffect(id, "seraph:"); !slices.Equal(got, []string{"a:1", "b:2"}) {
			t.Fatalf("notes = %v, want [a:1 b:2]", got)
		}
		if !g.ClearNotesWithPrefixForEffect(id, "seraph:") {
			t.Fatal("the permanent was not found")
		}
		if got := g.NotesWithPrefixForEffect(id, "seraph:"); len(got) != 0 {
			t.Errorf("notes = %v after taking them back", got)
		}
		if got := g.NotesForEffect(id); !slices.Equal(got, []string{"other"}) {
			t.Errorf("the other note was touched: %v", got)
		}
	})
}
