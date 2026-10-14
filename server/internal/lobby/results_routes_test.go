package lobby

// results_routes_test.go: the HTTP half of ADR 0174 §1 and §6.

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

func TestPutResultsRoute(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	meta, _ := l.Create("FNM")
	admin := adminSession(t, a)
	path := "/games/" + meta.ID.String() + "/results"
	put := func(token string, body any) (int, GameMeta, string) {
		t.Helper()
		resp := doJSON(t, srv, http.MethodPut, path, token, body)
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		var m GameMeta
		_ = json.Unmarshal(raw, &m)
		return resp.StatusCode, m, string(raw)
	}
	on := map[string]bool{"enabled": true}

	if code, _, body := put(admin, on); code != http.StatusServiceUnavailable || !strings.Contains(body, discord.ResultsChannelEnv) {
		t.Errorf("feed off: %d %s, want a 503 naming the variable", code, body)
	}
	feed, _ := quietFeed(newFakeResults())
	l.SetResultsFeed(feed)
	if code, _, _ := put(admin, map[string]any{}); code != http.StatusBadRequest {
		t.Errorf("no enabled field: %d, want 400", code)
	}
	if code, _, _ := put("", on); code != http.StatusUnauthorized {
		t.Errorf("anonymous: %d, want 401", code)
	}
	// A guest host may opt in: the post names seats, not accounts.
	gus := joinAs(t, srv, meta, "Gus")
	if code, m, _ := put(gus.Token, on); code != http.StatusOK || !m.Results {
		t.Errorf("guest host: %d results=%v", code, m.Results)
	}
	bob := joinAs(t, srv, meta, "Bob")
	if code, _, _ := put(bob.Token, map[string]bool{"enabled": false}); code != http.StatusForbidden {
		t.Errorf("non-host: %d, want 403", code)
	}
	if code, m, _ := put(admin, map[string]bool{"enabled": false}); code != http.StatusOK || m.Results {
		t.Errorf("admin opts out: %d results=%v", code, m.Results)
	}
	if mustGet(t, l, meta.ID).Results {
		t.Error("the opt-out did not stick")
	}
}

func TestCreateWithResults(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	admin := adminSession(t, a)
	create := func(body createGameRequest) GameMeta {
		t.Helper()
		resp := postJSON(t, srv, "/games", admin, body)
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create: %d %s", resp.StatusCode, raw)
		}
		var m GameMeta
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		return m
	}
	req := createGameRequest{Name: "recorded", Results: true}
	// Feed off: the ask is ignored, the table is still made.
	if m := create(req); m.Results || mustGet(t, l, m.ID).Results {
		t.Error("feed off, but the table was opted in")
	}
	feed, _ := quietFeed(newFakeResults())
	l.SetResultsFeed(feed)
	if m := create(req); !m.Results || !mustGet(t, l, m.ID).Results {
		t.Error("feed on, but the table was not opted in")
	}
	if m := create(createGameRequest{Name: "private"}); m.Results {
		t.Error("a table that did not ask was opted in")
	}
}

func TestCommunityReportsResults(t *testing.T) {
	srv, l, _ := newTestHTTPStack(t)
	read := func() map[string]any {
		t.Helper()
		resp := doGet(t, srv, "/community", "")
		defer resp.Body.Close()
		var body map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		return body
	}
	if b := read(); b["results"] != false {
		t.Errorf("feed off: results = %v", b["results"])
	}
	feed, _ := quietFeed(newFakeResults())
	l.SetResultsFeed(feed)
	if b := read(); b["results"] != true {
		t.Errorf("feed on: results = %v", b["results"])
	}
}
