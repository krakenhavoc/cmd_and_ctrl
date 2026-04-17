package deck

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func withArchidektStub(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	orig := archidektAPIHost
	archidektAPIHost = srv.URL
	t.Cleanup(func() {
		archidektAPIHost = orig
		srv.Close()
	})
	return srv
}

const archidektSample = `{
  "name": "Atraxa Superfriends",
  "cards": [
    {"quantity": 1, "categories": ["Commander"],  "card": {"oracleCard": {"name": "Atraxa, Praetors' Voice"}}},
    {"quantity": 1, "categories": ["Mainboard"],  "card": {"oracleCard": {"name": "Sol Ring"}}},
    {"quantity": 2, "categories": [],             "card": {"oracleCard": {"name": "Forest"}}},
    {"quantity": 1, "categories": ["Sideboard"],  "card": {"oracleCard": {"name": "Relic of Progenitus"}}},
    {"quantity": 1, "categories": ["Maybeboard"], "card": {"oracleCard": {"name": "Doubling Season"}}},
    {"quantity": 1, "categories": ["commander"],  "card": {"oracleCard": {"name": "case insensitive commander"}}}
  ]
}`

func TestParseArchidektBasic(t *testing.T) {
	name, entries, err := ParseArchidekt([]byte(archidektSample))
	if err != nil {
		t.Fatalf("ParseArchidekt: %v", err)
	}
	if name != "Atraxa Superfriends" {
		t.Errorf("name: got %q", name)
	}

	// Expect 5 entries: Atraxa (commander), Sol Ring, Forest, Relic
	// (sideboard), case-insensitive commander. Maybeboard skipped.
	if len(entries) != 5 {
		t.Fatalf("entries: got %d, want 5; %+v", len(entries), entries)
	}

	// Cross-check the first commander and the sideboard flag survived.
	got := make(map[string]Entry, len(entries))
	for _, e := range entries {
		got[e.Name] = e
	}
	if !got["Atraxa, Praetors' Voice"].IsCommander {
		t.Error("Atraxa entry: IsCommander not set")
	}
	if !got["case insensitive commander"].IsCommander {
		t.Error("lowercase 'commander' category: IsCommander not set")
	}
	if !got["Relic of Progenitus"].IsSideboard {
		t.Error("Relic of Progenitus: IsSideboard not set")
	}
	if got["Forest"].Count != 2 {
		t.Errorf("Forest count: got %d, want 2", got["Forest"].Count)
	}
	if _, present := got["Doubling Season"]; present {
		t.Error("maybeboard card should be filtered out")
	}
}

func TestParseArchidektMaybeboardOnlyFailsGracefully(t *testing.T) {
	// A deck whose only cards are all maybeboarded produces the
	// "no playable cards" error rather than a silent empty list.
	// Guards against a user pasting a URL where they've moved
	// everything to the maybeboard mid-brew.
	raw := `{
		"name": "x",
		"cards": [
			{"quantity": 1, "categories": ["Maybeboard"], "card": {"oracleCard": {"name": "A"}}}
		]
	}`
	_, _, err := ParseArchidekt([]byte(raw))
	if err == nil {
		t.Fatal("ParseArchidekt: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no playable cards") {
		t.Errorf("error message: got %q, want \"no playable cards\" substring", err.Error())
	}
}

func TestParseArchidektEmptyOrMalformed(t *testing.T) {
	cases := []struct {
		name    string
		payload string
	}{
		{"bad json", "not json"},
		{"empty object", "{}"},
		{"no cards", `{"name":"x","cards":[]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := ParseArchidekt([]byte(tc.payload))
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestFetchFromURLArchidektHappyPath(t *testing.T) {
	withArchidektStub(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/api/decks/55555/" {
			t.Errorf("path: got %q, want /api/decks/55555/", got)
		}
		if got := r.Header.Get("User-Agent"); got != userAgent {
			t.Errorf("User-Agent: got %q, want %q", got, userAgent)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, archidektSample)
	}))
	for _, u := range []string{
		"https://archidekt.com/decks/55555",
		"https://archidekt.com/decks/55555/atraxa-superfriends",
		"https://www.archidekt.com/decks/55555",
	} {
		t.Run(u, func(t *testing.T) {
			name, entries, err := FetchFromURL(context.Background(), DefaultClient(), u)
			if err != nil {
				t.Fatalf("FetchFromURL: %v", err)
			}
			if name != "Atraxa Superfriends" {
				t.Errorf("name: got %q", name)
			}
			if len(entries) != 5 {
				t.Errorf("entries: got %d, want 5", len(entries))
			}
		})
	}
}

func TestFetchFromURLArchidektMalformedPath(t *testing.T) {
	for _, u := range []string{
		"https://archidekt.com/",
		"https://archidekt.com/decks/",
		"https://archidekt.com/users/alice",
	} {
		t.Run(u, func(t *testing.T) {
			_, _, err := FetchFromURL(context.Background(), DefaultClient(), u)
			if !errors.Is(err, ErrUnknownSource) {
				t.Errorf("got %v, want ErrUnknownSource", err)
			}
		})
	}
}

func TestFetchFromURLArchidektNotFound(t *testing.T) {
	withArchidektStub(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "gone", http.StatusNotFound)
	}))
	_, _, err := FetchFromURL(context.Background(), DefaultClient(), "https://archidekt.com/decks/999")
	if !errors.Is(err, ErrDeckNotFound) {
		t.Errorf("got %v, want ErrDeckNotFound", err)
	}
}
