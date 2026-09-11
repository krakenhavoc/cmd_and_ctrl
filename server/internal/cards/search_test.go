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
