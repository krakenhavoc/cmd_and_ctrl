package cards

import (
	"testing"

	"github.com/google/uuid"
)

func searchIndex(t *testing.T) *Index {
	t.Helper()
	idx := NewIndex()
	for _, name := range []string{
		"Bolt", "Lightning Bolt", "Lightning Helix", "Chain Lightning",
		"Brainstorm", "Bolt Bend",
	} {
		idx.Put(Card{ID: uuid.New(), Name: name, TypeLine: "Instant"})
	}
	return idx
}

func names(cs []Card) []string {
	out := make([]string, 0, len(cs))
	for _, c := range cs {
		out = append(out, c.Name)
	}
	return out
}

func TestSearchRanksExactThenPrefixThenSubstring(t *testing.T) {
	got := names(searchIndex(t).Search("bolt", 10))
	want := []string{"Bolt", "Bolt Bend", "Lightning Bolt"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestSearchIsCaseAndSpaceInsensitive(t *testing.T) {
	if got := searchIndex(t).Search("  LIGHTNING  ", 10); len(got) != 3 {
		t.Fatalf("got %v, want 3 lightning cards", names(got))
	}
}

func TestSearchHonoursLimit(t *testing.T) {
	if got := searchIndex(t).Search("l", 2); len(got) != 2 {
		t.Fatalf("got %d results, want 2", len(got))
	}
}

// An empty query must not dump the whole index into a response.
func TestSearchEmptyQueryReturnsNothing(t *testing.T) {
	idx := searchIndex(t)
	for _, q := range []string{"", "   "} {
		if got := idx.Search(q, 10); got != nil {
			t.Errorf("Search(%q) = %v, want nil", q, names(got))
		}
	}
	if got := idx.Search("bolt", 0); got != nil {
		t.Errorf("limit 0 = %v, want nil", names(got))
	}
}

func TestSearchMissReturnsNothing(t *testing.T) {
	if got := searchIndex(t).Search("zzzznotacard", 10); len(got) != 0 {
		t.Errorf("got %v, want none", names(got))
	}
}

// A double-faced or split card is filed under its full name and under
// each face's name (Put), and a query that matches several of those keys
// used to return the SAME card once per key. The dev spawner keys its
// result list by card id, so the repeat threw Svelte's each_key_duplicate
// and took the table's UI down with it (#4424).
func TestSearchReturnsEachCardOnce(t *testing.T) {
	idx := NewIndex()
	dfc := Card{
		ID: uuid.New(), Name: "Aang, Swift Savior // Aang and La, Ocean's Fury", TypeLine: "Legendary Creature",
		CardFaces: []CardFace{
			{Name: "Aang, Swift Savior", TypeLine: "Legendary Creature"},
			{Name: "Aang and La, Ocean's Fury", TypeLine: "Legendary Creature"},
		},
	}
	other := Card{ID: uuid.New(), Name: "Aang's Journey", TypeLine: "Kindred Enchantment"}
	idx.Put(dfc)
	idx.Put(other)

	for _, q := range []string{"aang", "swift savior", "ocean's fury", "aang and la"} {
		seen := map[uuid.UUID]bool{}
		for _, c := range idx.Search(q, 40) {
			if seen[c.ID] {
				t.Errorf("Search(%q) returned %q more than once", q, c.Name)
			}
			seen[c.ID] = true
		}
		if !seen[dfc.ID] {
			t.Errorf("Search(%q) lost the double-faced card", q)
		}
	}

	// The repeat must not eat the limit either: with limit 2 both cards
	// still come back for "aang".
	if got := idx.Search("aang", 2); len(got) != 2 {
		t.Errorf("Search(aang, 2) = %v, want both cards", names(got))
	}
}
