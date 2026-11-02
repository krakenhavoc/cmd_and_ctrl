package cards

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

// set_name rides the Scryfall dump into Card.SetName, and EachPrinting
// walks every printing the index holds: the two inputs of the public
// catalogue's per-set counts (#4790).
func TestLoadCarriesSetNameAndEachPrintingVisitsAll(t *testing.T) {
	oracle := uuid.New().String()
	dump := `[
{"id":"` + uuid.New().String() + `","oracle_id":"` + oracle + `","name":"Lightning Bolt","set":"spm","set_name":"Marvel's Spider-Man","released_at":"2026-09-26","set_type":"expansion","lang":"en","layout":"normal","type_line":"Instant"},
{"id":"` + uuid.New().String() + `","oracle_id":"` + oracle + `","name":"Lightning Bolt","set":"spm","set_name":"Marvel's Spider-Man","released_at":"2026-09-26","set_type":"expansion","lang":"en","layout":"normal","type_line":"Instant"},
{"id":"` + uuid.New().String() + `","oracle_id":"` + oracle + `","name":"Lightning Bolt","set":"lea","set_name":"Limited Edition Alpha","released_at":"1993-08-05","set_type":"core","lang":"en","layout":"normal","type_line":"Instant"}
]`
	path := filepath.Join(t.TempDir(), "default-cards.json")
	if err := os.WriteFile(path, []byte(dump), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	idx := NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}

	seen := map[string]int{}
	idx.EachPrinting(func(c Card) { seen[c.SetCode+"="+c.SetName+"@"+c.ReleasedAt]++ })
	want := map[string]int{
		"spm=Marvel's Spider-Man@2026-09-26":   2,
		"lea=Limited Edition Alpha@1993-08-05": 1,
	}
	if len(seen) != len(want) {
		t.Fatalf("EachPrinting saw %v, want %v", seen, want)
	}
	for k, n := range want {
		if seen[k] != n {
			t.Fatalf("EachPrinting saw %v, want %v", seen, want)
		}
	}
}
