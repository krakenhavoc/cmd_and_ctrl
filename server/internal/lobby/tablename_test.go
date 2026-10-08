package lobby

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
)

func seededNames(seed uint64) *TableNameGenerator {
	return NewTableNameGenerator(rand.New(rand.NewPCG(seed, seed^0x9e3779b9)))
}

func TestTableNamesAreDeterministicUnderASeed(t *testing.T) {
	a, b := seededNames(7), seededNames(7)
	for i := 0; i < 50; i++ {
		x, y := a.Generate("Ada"), b.Generate("Ada")
		if x != y {
			t.Fatalf("draw %d: %q != %q", i, x, y)
		}
	}
	if seededNames(1).Generate("") == seededNames(2).Generate("") &&
		seededNames(1).Generate("") == seededNames(3).Generate("") {
		t.Error("different seeds all gave the same name")
	}
}

func TestTableNameWordListsAreBigEnoughAndDuplicateFree(t *testing.T) {
	for name, c := range map[string]struct {
		list []string
		min  int
	}{
		"adjectives": {nameAdjectives, 40},
		"nouns":      {nameNouns, 40},
		"plurals":    {namePlurals, 40},
		"groups":     {nameGroups, 8},
		"numbers":    {nameNumbers, 6},
		"troubles":   {nameTroubles, 6},
		"templates":  {nameTemplates, 12},
	} {
		if len(c.list) < c.min {
			t.Errorf("%s: %d entries, want at least %d", name, len(c.list), c.min)
		}
		seen := map[string]bool{}
		for _, w := range c.list {
			if strings.TrimSpace(w) != w || w == "" {
				t.Errorf("%s: %q is empty or has stray space", name, w)
			}
			if seen[strings.ToLower(w)] {
				t.Errorf("%s: %q appears twice", name, w)
			}
			seen[strings.ToLower(w)] = true
		}
	}
	if len(nameNouns) != len(namePlurals) {
		t.Fatalf("nouns (%d) and plurals (%d) must align", len(nameNouns), len(namePlurals))
	}
}

// Every template can fit, and whatever the generator returns fits, even
// for a display name of any length.
func TestEveryTemplateRendersWithinTheLimit(t *testing.T) {
	display := strings.Repeat("W", maxPersonalNameLen)
	for _, tpl := range nameTemplates {
		bare := strings.NewReplacer("{A}", "", "{N}", "", "{P}", "", "{G}", "", "{#}", "", "{T}", "", "{U}", display).Replace(tpl)
		if strings.ContainsAny(bare, "{}") {
			t.Errorf("template %q has an unknown marker", tpl)
		}
		if len(bare) > maxTableNameLen {
			t.Errorf("template %q cannot fit in %d", tpl, maxTableNameLen)
		}
	}
	g := seededNames(99)
	for i := 0; i < 5000; i++ {
		for _, d := range []string{"", "Ada", display, strings.Repeat("W", 200)} {
			if n := g.Generate(d); n == "" || len(n) > maxTableNameLen {
				t.Fatalf("Generate(%q) = %q (%d bytes)", d, n, len(n))
			}
		}
	}
}

func TestTableNamesUseTheDisplayNameOnlyWhenUsable(t *testing.T) {
	g := seededNames(3)
	used := 0
	for i := 0; i < 600; i++ {
		if strings.Contains(g.Generate("Zelda"), "Zelda") {
			used++
		}
	}
	if used == 0 || used > 400 {
		t.Errorf("a signed-in name was used %d times in 600, want sometimes", used)
	}
	for _, d := range []string{"", "   ", strings.Repeat("x", maxPersonalNameLen+1), "bad\nname"} {
		for i := 0; i < 300; i++ {
			n := g.Generate(d)
			if tr := strings.TrimSpace(d); tr != "" && strings.Contains(n, tr) {
				t.Fatalf("Generate(%q) used the name: %q", d, n)
			}
		}
	}
}

func TestSampleTableNames(t *testing.T) {
	g := seededNames(2630)
	for i := 0; i < 10; i++ {
		who := ""
		if i%2 == 0 {
			who = "Krakenhavoc"
		}
		t.Log(g.Generate(who))
	}
}

func personToken(t *testing.T, a auth.Authenticator, name string) string {
	t.Helper()
	tok, _, err := a.Issue(context.Background(), auth.Principal{Role: auth.RoleIdentified, UserID: uuid.New(), Name: name}, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	return tok
}

func TestBlankCreateGetsAGeneratedNameAndAnExplicitOneIsKept(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	admin := adminSession(t, a)
	for _, blank := range []string{"", "   "} {
		resp := postJSON(t, srv, "/games", admin, createGameRequest{Name: blank})
		var meta GameMeta
		_ = json.NewDecoder(resp.Body).Decode(&meta)
		resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("blank %q: status %d", blank, resp.StatusCode)
		}
		if strings.TrimSpace(meta.Name) == "" || len(meta.Name) > maxTableNameLen {
			t.Errorf("blank %q: name %q", blank, meta.Name)
		}
	}
	resp := postJSON(t, srv, "/games", admin, createGameRequest{Name: "  Friday Night  "})
	var meta GameMeta
	_ = json.NewDecoder(resp.Body).Decode(&meta)
	resp.Body.Close()
	if meta.Name != "Friday Night" {
		t.Errorf("explicit name = %q, want it kept (trimmed)", meta.Name)
	}

	// A signed-in person goes through CreateCapped.
	resp = postJSON(t, srv, "/games", personToken(t, a, "Ada"), createGameRequest{})
	meta = GameMeta{}
	_ = json.NewDecoder(resp.Body).Decode(&meta)
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated || strings.TrimSpace(meta.Name) == "" {
		t.Errorf("signed-in blank create: %d %q", resp.StatusCode, meta.Name)
	}
	if got := len(l.List()); got != 4 {
		t.Errorf("games = %d, want 4", got)
	}
}

func TestNameSuggestionRoute(t *testing.T) {
	srv, _, a := newTestHTTPStack(t)

	get := func(tok string) *http.Response {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/games/name-suggestion", nil)
		if tok != "" {
			req.Header.Set("Authorization", "Bearer "+tok)
		}
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("do: %v", err)
		}
		return resp
	}

	resp := get("")
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no credential: %d, want 401", resp.StatusCode)
	}

	// A guest seat (no user) may not create, so may not ask.
	guest, _, err := a.Issue(context.Background(), auth.Principal{Role: auth.RolePlayer, GameID: uuid.New(), PlayerID: uuid.New(), Name: "G"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	resp = get(guest)
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("guest: %d, want 403", resp.StatusCode)
	}

	for name, tok := range map[string]string{"admin": adminSession(t, a), "person": personToken(t, a, "Ada")} {
		resp = get(tok)
		var body map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s: %d", name, resp.StatusCode)
		}
		if len(body) != 1 || strings.TrimSpace(body["name"]) == "" || len(body["name"]) > maxTableNameLen {
			t.Errorf("%s: body %v", name, body)
		}
	}
}
