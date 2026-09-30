package cards

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

// token_pool_test.go — ADR 0078 decision 3a: the eligible pool
// Index.Load builds for internal/cards/tokenart, and nothing else in
// this package reads. Each case below pins one exclusion clause on a
// pool candidate that a naive `layout == "token"` scan would let
// through.
func TestTokenPrintingPoolFiltersEligibility(t *testing.T) {
	good := uuid.New().String()
	dump := `[
{"id":"` + good + `","name":"Treasure","set_type":"expansion","lang":"en","layout":"token","border_color":"black","type_line":"Token Artifact — Treasure","image_uris":{"normal":"https://example.test/treasure.jpg"}},
{"id":"` + uuid.New().String() + `","name":"Treasure Non-English","set_type":"expansion","lang":"ja","layout":"token","border_color":"black","type_line":"Token Artifact — Treasure","image_uris":{"normal":"https://example.test/treasure-ja.jpg"}},
{"id":"` + uuid.New().String() + `","name":"Treasure No Art","set_type":"expansion","lang":"en","layout":"token","border_color":"black","type_line":"Token Artifact — Treasure"},
{"id":"` + uuid.New().String() + `","name":"Treasure Gold Border","set_type":"expansion","lang":"en","layout":"token","border_color":"gold","type_line":"Token Artifact — Treasure","image_uris":{"normal":"https://example.test/treasure-gold.jpg"}},
{"id":"` + uuid.New().String() + `","name":"Treasure Memorabilia","set_type":"memorabilia","lang":"en","layout":"token","border_color":"black","type_line":"Token Artifact — Treasure","image_uris":{"normal":"https://example.test/treasure-memo.jpg"}},
{"id":"` + uuid.New().String() + `","name":"Lightning Bolt","set_type":"expansion","lang":"en","layout":"normal","border_color":"black","type_line":"Instant","image_uris":{"normal":"https://example.test/bolt.jpg"}}
]`
	dir := t.TempDir()
	path := filepath.Join(dir, "default-cards.json")
	if err := os.WriteFile(path, []byte(dump), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	idx := NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}

	pool := idx.TokenPrintingPool()
	if len(pool) != 1 {
		ids := make([]string, len(pool))
		for i, c := range pool {
			ids[i] = c.Name
		}
		t.Fatalf("TokenPrintingPool: got %d entries %v, want 1 (only the eligible Treasure)", len(pool), ids)
	}
	if pool[0].ID.String() != good {
		t.Errorf("TokenPrintingPool: got id %s, want %s", pool[0].ID, good)
	}
}

func TestTokenPrintingPoolEmptyOnUnloadedIndex(t *testing.T) {
	idx := NewIndex()
	if pool := idx.TokenPrintingPool(); len(pool) != 0 {
		t.Errorf("TokenPrintingPool on a fresh Index: got %d entries, want 0", len(pool))
	}
}
