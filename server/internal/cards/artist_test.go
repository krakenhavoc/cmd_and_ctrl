package cards

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// #2982: Scryfall's art_crop image drops the artist line, so the client
// credits the artist from GET /cards/{id}. The field is additive and
// omitempty: a record that prints no artist must not grow an empty key.
const artistFixture = `[
 {"id":"aaaaaaaa-0000-0000-0000-000000000001","name":"Credited","artist":"Mark Tedin","image_uris":{"normal":"https://x/n.jpg"}},
 {"id":"aaaaaaaa-0000-0000-0000-000000000002","name":"Front // Back","layout":"transform","artist":"Front Artist",
  "card_faces":[{"name":"Front","artist":"Front Artist"},{"name":"Back","artist":"Back Artist"}]},
 {"id":"aaaaaaaa-0000-0000-0000-000000000003","name":"Anonymous","image_uris":{"normal":"https://x/n.jpg"}}
]`

func artistServer(t *testing.T) *httptest.Server {
	t.Helper()
	path := filepath.Join(t.TempDir(), "default-cards.json")
	if err := os.WriteFile(path, []byte(artistFixture), 0o644); err != nil {
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

func getRaw(t *testing.T, url string) map[string]json.RawMessage {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var m map[string]json.RawMessage
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return m
}

func TestCardMetadataCarriesArtist(t *testing.T) {
	srv := artistServer(t)
	m := getRaw(t, srv.URL+"/cards/aaaaaaaa-0000-0000-0000-000000000001")
	if string(m["artist"]) != `"Mark Tedin"` {
		t.Fatalf("artist = %s, want \"Mark Tedin\"", m["artist"])
	}
}

func TestCardMetadataCarriesPerFaceArtist(t *testing.T) {
	srv := artistServer(t)
	m := getRaw(t, srv.URL+"/cards/aaaaaaaa-0000-0000-0000-000000000002")
	var faces []CardFace
	if err := json.Unmarshal(m["card_faces"], &faces); err != nil {
		t.Fatal(err)
	}
	if len(faces) != 2 || faces[0].Artist != "Front Artist" || faces[1].Artist != "Back Artist" {
		t.Fatalf("face artists = %+v", faces)
	}
}

func TestCardMetadataOmitsEmptyArtist(t *testing.T) {
	srv := artistServer(t)
	m := getRaw(t, srv.URL+"/cards/aaaaaaaa-0000-0000-0000-000000000003")
	if _, ok := m["artist"]; ok {
		t.Fatalf("artist key present for a card with none: %s", m["artist"])
	}
}
