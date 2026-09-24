package roadmap

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// The route answers a request that carries no cookie and no token at
// all: it is the public half of ADR 0092 Decision 1.
func TestHandlerServesTheRoadmapWithoutASession(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/roadmap", nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=300" {
		t.Errorf("Cache-Control = %q, want public, max-age=300", got)
	}

	var got Roadmap
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body does not parse: %v", err)
	}
	if want := Build(); !reflect.DeepEqual(got, want) {
		t.Errorf("served roadmap differs from Build()")
	}
	if len(got.Items) == 0 {
		t.Fatal("served roadmap has no items")
	}
	if got.Counts.Cards.Total == 0 || got.Counts.Cards.Full == 0 {
		t.Errorf("catalog counts missing: %+v", got.Counts.Cards)
	}
	if sum := got.Counts.Cards.Full + got.Counts.Cards.Caveats + got.Counts.Cards.Unreviewed; sum != got.Counts.Cards.Total {
		t.Errorf("catalog counts add up to %d, total is %d", sum, got.Counts.Cards.Total)
	}
}

// Walks every key in the served body, at any depth, so a field added
// later under a nested object cannot slip an image or oracle text
// through.
func TestHandlerBodyCarriesNoArtOrOracleText(t *testing.T) {
	rec := httptest.NewRecorder()
	Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/roadmap", nil))

	var doc any
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatalf("body does not parse: %v", err)
	}
	banned := []string{"image", "oracle_text", "scryfall_id", "engine_notes"}
	var walk func(path string, v any)
	walk = func(path string, v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, child := range x {
				for _, b := range banned {
					if strings.Contains(strings.ToLower(k), b) {
						t.Errorf("%s.%s: banned key in the public roadmap", path, k)
					}
				}
				walk(path+"."+k, child)
			}
		case []any:
			for _, child := range x {
				walk(path+"[]", child)
			}
		case string:
			if strings.Contains(x, "scryfall.io") || strings.Contains(x, "/image") {
				t.Errorf("%s: an image URL in the public roadmap: %q", path, x)
			}
		}
	}
	walk("$", doc)
}

func TestHandlerAnswersHeadAndRefusesWrites(t *testing.T) {
	rec := httptest.NewRecorder()
	Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodHead, "/roadmap", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("HEAD status = %d, want 200", rec.Code)
	}
	for _, m := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		rec := httptest.NewRecorder()
		Handler().ServeHTTP(rec, httptest.NewRequest(m, "/roadmap", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s status = %d, want 405", m, rec.Code)
		}
	}
}
