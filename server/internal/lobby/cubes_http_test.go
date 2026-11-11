package lobby

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cubes"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draft"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/drafthistory"
)

// Saved cubes and the cube draft as the lobby serves them (ADR 0222 §8,
// #5441). The check's own rules are tested in internal/cubes and the
// pack source's in internal/draft.

// cubeStack is a draft stack with a database, the cube store over it
// and a deck fetcher that never reaches the network.
type cubeStack struct {
	*draftStack
	d       *db.DB
	fetched int
}

func newCubeStack(t *testing.T) *cubeStack {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	cs := &cubeStack{d: d}
	cs.draftStack = newDraftStackWith(t, func(c *Config) {
		c.Cubes = cubes.NewSQLStore(d)
		c.FetchDeck = func(context.Context, string) (string, []deck.Entry, error) {
			cs.fetched++
			entries := []deck.Entry{{Name: "syn Rare 00", Count: 1, IsCommander: true}, {Name: "syn Rare 01", Count: 1, IsSideboard: true}}
			for i := range 30 {
				entries = append(entries, deck.Entry{Name: fmt.Sprintf("syn Common %02d", i), Count: 1})
			}
			return "A cube from a link", entries, nil
		}
	})
	return cs
}

// person signs someone in and gives them a users row, which a cube's
// owner must have.
func (s *cubeStack) person(t *testing.T, name string) (token string, user uuid.UUID) {
	t.Helper()
	token, user = s.signIn(t, name)
	if _, err := s.d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, ?, 1, 1)`, user.String(), name); err != nil {
		t.Fatal(err)
	}
	return token, user
}

// cubeCall makes one request and decodes whichever body came back.
func (s *cubeStack) cubeCall(t *testing.T, method, path, token string, body any) (int, cubeAnswer) {
	t.Helper()
	var rd io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		rd = strings.NewReader(string(raw))
	}
	req, err := http.NewRequest(method, s.srv.URL+path, rd)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := s.srv.Client().Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out cubeAnswer
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatalf("%s %s: %d, body is not JSON: %s", method, path, resp.StatusCode, raw)
		}
	}
	return resp.StatusCode, out
}

type cubeAnswer struct {
	Cube     cubeInfo   `json:"cube"`
	Cubes    []cubeInfo `json:"cubes"`
	MaxCards int        `json:"max_cards"`
	Error    string     `json:"error"`
	Code     string     `json:"code"`
}

// synList is n different cards of the test index's draftable set, one a
// row. The set has 260.
func synList(n int) string {
	var b strings.Builder
	kinds := []struct {
		prefix string
		count  int
	}{{"Common", 80}, {"Uncommon", 100}, {"Rare", 60}, {"Mythic", 20}}
	for _, k := range kinds {
		for i := 0; i < k.count && n > 0; i++ {
			fmt.Fprintf(&b, "1 syn %s %02d\n", k.prefix, i)
			n--
		}
	}
	return b.String()
}

func (s *cubeStack) saveCube(t *testing.T, token, name, text string) cubeInfo {
	t.Helper()
	code, out := s.cubeCall(t, http.MethodPost, "/me/cubes", token, map[string]any{"name": name, "text": text})
	if code != http.StatusCreated {
		t.Fatalf("POST /me/cubes: %d %s", code, out.Error)
	}
	return out.Cube
}

// Create, check and delete: a cube is saved even when cards block it,
// its check names them, and dropping them makes it ready. It is its
// owner's alone.
func TestMyCubesCreateCheckAndDelete(t *testing.T) {
	s := newCubeStack(t)
	ann, _ := s.person(t, "Ann")
	bob, _ := s.person(t, "Bob")

	// "wip Common 07" is the one card the test catalog does not hold as
	// Full; "Nonesuch" is not a card.
	text := synList(100) + "wip Common 07\n1 Nonesuch\n2 syn Common 00\n"
	cube := s.saveCube(t, ann, "  Friday   cube ", text)
	if cube.Name != "Friday cube" || !cube.Singleton {
		t.Fatalf("saved cube = %q, singleton %v; want the trimmed name and singleton by default", cube.Name, cube.Singleton)
	}
	ck := cube.Check
	if ck.Ready || ck.Cards != 104 ||
		fmt.Sprint(ck.NotFull) != "[wip Common 07]" || fmt.Sprint(ck.Unknown) != "[Nonesuch]" || fmt.Sprint(ck.Repeated) != "[syn Common 00]" {
		t.Fatalf("check = %+v", ck)
	}
	if !strings.Contains(cube.List, "1 wip Common 07\n") || !strings.Contains(cube.List, "3 syn Common 00\n") {
		t.Fatalf("the stored list does not hold what was pasted, merged by card:\n%s", cube.List)
	}

	// The list of cubes carries the check and not the cards.
	code, list := s.cubeCall(t, http.MethodGet, "/me/cubes", ann, nil)
	if code != http.StatusOK || len(list.Cubes) != 1 || list.Cubes[0].List != "" || list.Cubes[0].Check.Ready || list.MaxCards != cubes.MaxCards {
		t.Fatalf("GET /me/cubes: %d %+v", code, list)
	}
	// Nobody else can see, change or delete it.
	path := "/me/cubes/" + cube.ID.String()
	if code, out := s.cubeCall(t, http.MethodGet, "/me/cubes", bob, nil); code != http.StatusOK || len(out.Cubes) != 0 {
		t.Fatalf("Bob's list: %d, %d cubes", code, len(out.Cubes))
	}
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		var body any
		if method == http.MethodPut {
			body = map[string]any{"name": "Mine now"}
		}
		if code, _ := s.cubeCall(t, method, path, bob, body); code != http.StatusNotFound {
			t.Fatalf("%s by another person: %d, want 404", method, code)
		}
	}

	// The owner drops the rows the check named, and it is ready.
	code, out := s.cubeCall(t, http.MethodPut, path, ann, map[string]any{"text": synList(100)})
	if code != http.StatusOK || !out.Cube.Check.Ready || out.Cube.Check.Cards != 100 || out.Cube.Check.MaxSeats != 2 {
		t.Fatalf("PUT a clean list: %d %+v", code, out.Cube.Check)
	}
	// Singleton waived: the repeated card no longer blocks.
	code, out = s.cubeCall(t, http.MethodPut, path, ann, map[string]any{"text": synList(100) + "2 syn Common 00\n", "singleton": false})
	if code != http.StatusOK || !out.Cube.Check.Ready || out.Cube.Singleton || out.Cube.Check.Cards != 102 {
		t.Fatalf("PUT with singleton off: %d %+v", code, out.Cube)
	}
	// And enforced again by the setting alone, with the list unchanged.
	code, out = s.cubeCall(t, http.MethodPut, path, ann, map[string]any{"singleton": true})
	if code != http.StatusOK || out.Cube.Check.Ready || fmt.Sprint(out.Cube.Check.Repeated) != "[syn Common 00]" {
		t.Fatalf("PUT singleton on: %d %+v", code, out.Cube.Check)
	}

	if code, _ := s.cubeCall(t, http.MethodDelete, path, ann, nil); code != http.StatusNoContent {
		t.Fatalf("DELETE: %d", code)
	}
	if code, _ := s.cubeCall(t, http.MethodGet, path, ann, nil); code != http.StatusNotFound {
		t.Fatalf("GET a deleted cube: %d, want 404", code)
	}
}

// What a save refuses: a filtered name, a name already used, a list
// with no card or too many, both sources at once, and a caller who is
// not signed in.
func TestACubeSaveIsRefused(t *testing.T) {
	s := newCubeStack(t)
	ann, _ := s.person(t, "Ann")
	s.saveCube(t, ann, "Taken", synList(10))

	for name, tc := range map[string]struct {
		body map[string]any
		want int
		code string
	}{
		"a filtered name":        {map[string]any{"name": "kys everyone", "text": synList(10)}, http.StatusBadRequest, CodeNameNotAllowed},
		"a name of another case": {map[string]any{"name": "taken", "text": synList(10)}, http.StatusCreated, ""}, // names differ by case
		"the same name":          {map[string]any{"name": "Taken", "text": synList(10)}, http.StatusConflict, CodeCubeNameTaken},
		"no name":                {map[string]any{"text": synList(10)}, http.StatusBadRequest, ""},
		"a long name":            {map[string]any{"name": strings.Repeat("a", cubes.MaxNameLen+1), "text": synList(10)}, http.StatusBadRequest, ""},
		"no list":                {map[string]any{"name": "Empty"}, http.StatusBadRequest, ""},
		"a list of nothing":      {map[string]any{"name": "Empty", "text": "# only a comment"}, http.StatusBadRequest, ""},
		"both sources":           {map[string]any{"name": "Both", "text": synList(10), "url": "https://archidekt.com/decks/1"}, http.StatusBadRequest, ""},
		"too many cards":         {map[string]any{"name": "Huge", "text": fmt.Sprintf("%d syn Common 00\n", cubes.MaxCards+1), "singleton": false}, http.StatusBadRequest, CodeCubeTooLarge},
	} {
		code, out := s.cubeCall(t, http.MethodPost, "/me/cubes", ann, tc.body)
		if code != tc.want || out.Code != tc.code {
			t.Errorf("%s: %d code %q (%s), want %d code %q", name, code, out.Code, out.Error, tc.want, tc.code)
		}
	}
	// Exactly the cap is saved.
	if code, out := s.cubeCall(t, http.MethodPost, "/me/cubes", ann, map[string]any{
		"name": "At the cap", "text": fmt.Sprintf("%d syn Common 00\n", cubes.MaxCards), "singleton": false,
	}); code != http.StatusCreated || !out.Cube.Check.Ready || out.Cube.Check.MaxSeats != draft.MaxSeats {
		t.Errorf("a cube of %d cards: %d %+v", cubes.MaxCards, code, out.Cube.Check)
	}

	// A visitor, with no session at all, and a guest's table seat are
	// both turned away; nothing is saved for either.
	for _, path := range []string{"/me/cubes", "/me/cubes/" + uuid.NewString()} {
		if code, _ := s.cubeCall(t, http.MethodGet, path, "", nil); code != http.StatusUnauthorized {
			t.Errorf("GET %s with no session: %d, want 401", path, code)
		}
	}
	if code, _ := s.cubeCall(t, http.MethodPost, "/me/cubes", "", map[string]any{"name": "Visitor's", "text": synList(10)}); code != http.StatusUnauthorized {
		t.Errorf("POST /me/cubes with no session: %d, want 401", code)
	}
	var n int
	if err := s.d.QueryRow(`SELECT COUNT(*) FROM cubes`).Scan(&n); err != nil || n != 3 {
		t.Errorf("%d cubes stored (%v), want Ann's three", n, err)
	}
}

// A cube from a deck-site link: every row of the fetched deck is a card
// of the cube, whichever board it was on, and the link's name stands in
// for a name.
func TestACubeFromALink(t *testing.T) {
	s := newCubeStack(t)
	ann, _ := s.person(t, "Ann")
	code, out := s.cubeCall(t, http.MethodPost, "/me/cubes", ann, map[string]any{"url": "https://archidekt.com/decks/123456/a-cube"})
	if code != http.StatusCreated {
		t.Fatalf("POST /me/cubes with a link: %d %s", code, out.Error)
	}
	if out.Cube.Name != "A cube from a link" || out.Cube.Check.Cards != 32 || !out.Cube.Check.Ready || s.fetched != 1 {
		t.Fatalf("cube from a link = %q, %+v, fetched %d times", out.Cube.Name, out.Cube.Check, s.fetched)
	}
	for _, row := range []string{"1 syn Rare 00\n", "1 syn Rare 01\n", "1 syn Common 29\n"} {
		if !strings.Contains(out.Cube.List, row) {
			t.Errorf("the cube is missing %q: commander and sideboard rows are cards of a cube", strings.TrimSpace(row))
		}
	}
	if code, out := s.cubeCall(t, http.MethodPost, "/me/cubes", ann, map[string]any{"name": "Bad link", "url": "https://example.test/not-a-deck"}); code == http.StatusCreated {
		t.Fatalf("a link to no deck site was saved: %+v", out.Cube)
	}
}

// A four-seat cube draft, one person and three bots, through the picks
// and the build: the packs come from the cube and nowhere else, no card
// is dealt twice, and what follows is the draft as it is for a set.
func TestACubeDraftWithBotsThroughDeckbuilding(t *testing.T) {
	s := newCubeStack(t)
	store := drafthistory.NewSQLStore(s.d)
	s.lobby.SetDraftRecorder(store)
	ann, user := s.person(t, "Ann")
	cube := s.saveCube(t, ann, "Friday cube", synList(200))
	if !cube.Check.Ready || cube.Check.MaxSeats != 4 {
		t.Fatalf("check = %+v, want ready for four seats", cube.Check)
	}

	body := fmt.Sprintf(`{"cube":%q,"seats":4,"name":"Cube night"}`, cube.ID)
	host := s.create(t, ann, body)
	v := host.Draft
	if !v.Cube || v.Set != "" || v.SetName != "Friday cube" || v.Booster != CubeBooster || v.Seats != 4 || v.Solo {
		t.Fatalf("draft view = cube %v, set %q, set_name %q, booster %q, seats %d, solo %v", v.Cube, v.Set, v.SetName, v.Booster, v.Seats, v.Solo)
	}
	id := v.ID
	base := "/draft/" + id.String()
	seat := s.draftSeatOf(t, host.Token)
	if resp, out := s.call(t, http.MethodPost, base+"/start", host.Token, ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("start: %d %s", resp.StatusCode, out.Error)
	}
	room := s.lobby.DraftRoom(id)
	first, _ := room.View(seat)
	if len(first.Pack) != draft.DefaultCubePackSize || first.PicksInRound != draft.DefaultCubePackSize || first.Rounds != draft.Rounds {
		t.Fatalf("first pack has %d cards, %d picks a round, %d rounds", len(first.Pack), first.PicksInRound, first.Rounds)
	}
	pickEverything(t, room, seat)

	pools, done := room.Pools()
	if !done || len(pools) != 4 {
		t.Fatalf("pools: done %v, %d seats", done, len(pools))
	}
	inCube := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(cube.List), "\n") {
		inCube[strings.SplitN(line, " ", 2)[1]] = true
	}
	seen := map[string]bool{}
	for i, pool := range pools {
		if len(pool) != draft.Rounds*draft.DefaultCubePackSize {
			t.Fatalf("seat %d drafted %d cards, want %d", i, len(pool), draft.Rounds*draft.DefaultCubePackSize)
		}
		for _, c := range pool {
			if !inCube[c.Name] {
				t.Fatalf("seat %d drafted %q, which is not in the cube", i, c.Name)
			}
			if seen[c.Name] {
				t.Fatalf("%q was dealt twice from a singleton cube", c.Name)
			}
			seen[c.Name] = true
		}
	}

	// Deckbuilding, with free basic lands, then the rounds begin.
	built, _ := room.View(seat)
	if built.Phase != draft.PhaseBuilding || built.Build == nil {
		t.Fatalf("after the picks: phase %q", built.Phase)
	}
	if err := room.SubmitDeck(seat, deckOf(t, room, seat)); err != nil {
		t.Fatalf("SubmitDeck: %v", err)
	}
	after, _ := room.View(seat)
	if after.Phase != draft.PhaseRounds || after.Deck == nil || len(after.Deck.Main) != 23 {
		t.Fatalf("after the build: phase %q, deck %+v", after.Phase, after.Deck)
	}

	// The look back names the cube where a set's draft names the set.
	if resp, _ := s.call(t, http.MethodPost, base+"/leave", host.Token, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("leave: %d", resp.StatusCode)
	}
	rec, err := store.Get(context.Background(), user, id)
	if err != nil {
		t.Fatalf("the draft was not kept: %v", err)
	}
	if rec.Set != "" || rec.SetName != "Friday cube" || len(rec.Picks) != 45 {
		t.Fatalf("stored draft: set %q, set_name %q, %d picks", rec.Set, rec.SetName, len(rec.Picks))
	}
}

// What a cube draft refuses: a visitor, a cube that is not the
// caller's, one with a card that blocks, one too small for the seats
// and pack size, and a pack size out of bounds. A set's draft is
// untouched: a visitor still drafts one.
func TestACubeDraftIsRefused(t *testing.T) {
	s := newCubeStack(t)
	ann, _ := s.person(t, "Ann")
	bob, _ := s.person(t, "Bob")
	ready := s.saveCube(t, ann, "Ready", synList(200))
	blocked := s.saveCube(t, ann, "Blocked", synList(200)+"1 wip Common 07\n")
	q := func(c cubeInfo, rest string) string { return fmt.Sprintf(`{"cube":%q%s}`, c.ID, rest) }

	for name, tc := range map[string]struct {
		token, body string
		want        int
		code        string
	}{
		"a visitor":                 {"", q(ready, `,"seats":4`), http.StatusForbidden, CodeCubeNeedsSignIn},
		"another person's cube":     {bob, q(ready, `,"seats":4`), http.StatusNotFound, ""},
		"no such cube":              {ann, fmt.Sprintf(`{"cube":%q,"seats":4}`, uuid.NewString()), http.StatusNotFound, ""},
		"not a cube id":             {ann, `{"cube":"friday","seats":4}`, http.StatusNotFound, ""},
		"a card that is not Full":   {ann, q(blocked, `,"seats":4`), http.StatusUnprocessableEntity, CodeCubeNotReady},
		"too small for eight seats": {ann, q(ready, ``), http.StatusUnprocessableEntity, CodeCubeTooSmall},
		"too small for five":        {ann, q(ready, `,"seats":5`), http.StatusUnprocessableEntity, CodeCubeTooSmall},
		"too small for big packs":   {ann, q(ready, `,"seats":4,"pack_size":17`), http.StatusUnprocessableEntity, CodeCubeTooSmall},
		"a pack too small":          {ann, q(ready, `,"seats":4,"pack_size":8`), http.StatusBadRequest, ""},
		"a pack too large":          {ann, q(ready, `,"seats":4,"pack_size":21`), http.StatusBadRequest, ""},
		"a set and a cube":          {ann, q(ready, `,"set":"syn","seats":4`), http.StatusBadRequest, ""},
		"a pack size for a set":     {ann, `{"set":"syn","pack_size":15}`, http.StatusBadRequest, ""},
	} {
		resp, out := s.call(t, http.MethodPost, "/draft", tc.token, tc.body)
		if resp.StatusCode != tc.want || out.Code != tc.code {
			t.Errorf("%s: %d code %q (%s), want %d code %q", name, resp.StatusCode, out.Code, out.Error, tc.want, tc.code)
		}
	}
	if s.lobby.DraftCount() != 0 {
		t.Fatalf("%d drafts were made by refused requests", s.lobby.DraftCount())
	}
	// The message names the card that blocks.
	_, out := s.call(t, http.MethodPost, "/draft", ann, q(blocked, `,"seats":4`))
	if !strings.Contains(out.Error, "wip Common 07") {
		t.Errorf("the refusal does not name the card: %q", out.Error)
	}

	// Smaller packs fit more seats: 200 cards deal eight seats of nine.
	small := s.create(t, ann, q(ready, `,"seats":7,"pack_size":9`))
	if !small.Draft.Cube || small.Draft.Seats != 7 {
		t.Fatalf("a seven-seat draft of nine-card packs: %+v", small.Draft)
	}
	// A visitor still drafts a set, with no cube and no account.
	visitor := s.create(t, "", `{"set":"syn"}`)
	if visitor.Draft.Cube || visitor.Draft.Set != "syn" || !visitor.Draft.Solo {
		t.Fatalf("a visitor's set draft = %+v", visitor.Draft)
	}
}
