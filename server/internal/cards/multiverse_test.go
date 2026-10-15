package cards

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// #3831: the client links a card's rulings on Gatherer when it has the
// printing's multiverse ids. The field is additive and omitempty: a
// printing Gatherer does not list must not grow an empty key.
const multiverseFixture = `[
 {"id":"bbbbbbbb-0000-0000-0000-000000000001","name":"Delver of Secrets // Insectile Aberration","layout":"transform","set":"isd","collector_number":"51","multiverse_ids":[226749,226755],"image_uris":{"normal":"https://x/n.jpg"}},
 {"id":"bbbbbbbb-0000-0000-0000-000000000002","name":"Purphoros, God of the Forge","set":"fca","collector_number":"14","multiverse_ids":[],"image_uris":{"normal":"https://x/n.jpg"}}
]`

func multiverseServer(t *testing.T) *httptest.Server {
	t.Helper()
	path := filepath.Join(t.TempDir(), "default-cards.json")
	if err := os.WriteFile(path, []byte(multiverseFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	idx := NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	cache, _ := NewImageCache(t.TempDir())
	srv := httptest.NewServer(Handler(idx, cache))
	t.Cleanup(srv.Close)
	return srv
}

func TestCardMetadataCarriesMultiverseIDs(t *testing.T) {
	srv := multiverseServer(t)
	m := getRaw(t, srv.URL+"/cards/bbbbbbbb-0000-0000-0000-000000000001")
	var ids []int
	if err := json.Unmarshal(m["multiverse_ids"], &ids); err != nil {
		t.Fatalf("multiverse_ids = %s: %v", m["multiverse_ids"], err)
	}
	if len(ids) != 2 || ids[0] != 226749 || ids[1] != 226755 {
		t.Fatalf("multiverse_ids = %v, want [226749 226755]", ids)
	}
	if string(m["set"]) != `"isd"` || string(m["collector_number"]) != `"51"` {
		t.Fatalf("set/collector = %s/%s", m["set"], m["collector_number"])
	}
}

func TestCardMetadataOmitsEmptyMultiverseIDs(t *testing.T) {
	srv := multiverseServer(t)
	resp, err := http.Get(srv.URL + "/cards/bbbbbbbb-0000-0000-0000-000000000002")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var m map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["multiverse_ids"]; ok {
		t.Fatalf("multiverse_ids key present for a printing with none: %s", m["multiverse_ids"])
	}
}
