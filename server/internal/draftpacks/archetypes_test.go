package draftpacks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// pairedSet is a made-up set whose first twelve uncommons are, in order,
// a gold card of each of the ten pairs, a mono-coloured card and a
// three-coloured one; the thirteenth is a gold card of WU again, and
// the fourteenth a double-faced card whose colours are on its front
// face only.
func pairedSet(t *testing.T) *cards.Index {
	t.Helper()
	idx := cards.NewIndex()
	synthSet(idx, "syn", 80, 100, 60, 20, true)
	colours := [][]string{
		{"W", "U"}, {"U", "B"}, {"B", "R"}, {"R", "G"}, {"G", "W"},
		{"W", "B"}, {"U", "R"}, {"B", "G"}, {"R", "W"}, {"G", "U"},
		{"W"}, {"W", "U", "B"}, {"U", "W"},
	}
	for i, cs := range colours {
		c := synthCard("syn", fmt.Sprintf("syn Uncommon %02d", i), "uncommon", fmt.Sprint(81+i))
		c.Colors = cs
		c.ManaCost = "{1}{" + cs[0] + "}"
		idx.Put(c)
	}
	dfc := synthCard("syn", "syn Uncommon 13", "uncommon", "94")
	dfc.Layout = "transform"
	dfc.CardFaces = []cards.CardFace{{Name: "front", Colors: []string{"B", "G"}}, {Name: "back", Colors: []string{"G"}}}
	idx.Put(dfc)
	return idx
}

func TestArchetypesAreTheSetsGoldUncommonsByPair(t *testing.T) {
	idx := pairedSet(t)
	l := NewListerWith(idx, allFullExcept())
	a, ok := l.Archetypes("syn")
	if !ok {
		t.Fatal("no archetypes for a draftable set")
	}
	if a.Set != "syn" || a.Name != "Synthetic syn" {
		t.Errorf("header %q %q", a.Set, a.Name)
	}
	var order []string
	for _, p := range a.Pairs {
		order = append(order, p.Colors)
	}
	if !reflect.DeepEqual(order, Pairs) || len(order) != 10 {
		t.Fatalf("pairs %v, want the ten in Pairs order", order)
	}
	names := func(p Pair) []string {
		var out []string
		for _, s := range p.Signposts {
			out = append(out, strings.TrimPrefix(s.Name, "syn Uncommon "))
		}
		return out
	}
	want := map[string][]string{
		"WU": {"00", "12"}, "UB": {"01"}, "BR": {"02"}, "RG": {"03"}, "WG": {"04"},
		"WB": {"05"}, "UR": {"06"}, "BG": {"07", "13"}, "WR": {"08"}, "UG": {"09"},
	}
	for _, p := range a.Pairs {
		if got := names(p); !reflect.DeepEqual(got, want[p.Colors]) {
			t.Errorf("%s: signposts %v, want %v", p.Colors, got, want[p.Colors])
		}
		for _, s := range p.Signposts {
			if s.ID.String() == "00000000-0000-0000-0000-000000000000" || s.ManaCost == "" && !strings.HasSuffix(s.Name, "13") {
				t.Errorf("%s: signpost %+v is missing its printing or cost", p.Colors, s)
			}
		}
	}
	// A mono-coloured or three-coloured uncommon is nobody's signpost.
	for _, p := range a.Pairs {
		for _, s := range p.Signposts {
			if strings.HasSuffix(s.Name, "10") || strings.HasSuffix(s.Name, "11") {
				t.Errorf("%s is a signpost of %s", s.Name, p.Colors)
			}
		}
	}
	if _, ok := l.Archetypes("nope"); ok {
		t.Error("an unknown set has archetypes")
	}
}

func TestAPairWithNoLabelHasNoLabelAndNoSignpostsIsAnEmptyList(t *testing.T) {
	idx := cards.NewIndex()
	synthSet(idx, "syn", 80, 100, 60, 20, true)
	pool := PoolFor(idx, "syn")
	a := archetypesWith(idx, pool, "Synthetic", map[string]string{"WU": "Birds and flyers"})
	for _, p := range a.Pairs {
		if p.Signposts == nil {
			t.Errorf("%s: signposts is nil, the client reads a list", p.Colors)
		}
		if (p.Colors == "WU") != (p.Label != "") {
			t.Errorf("%s: label %q", p.Colors, p.Label)
		}
	}
	raw, _ := json.Marshal(a.Pairs[1])
	if strings.Contains(string(raw), `"label"`) {
		t.Errorf("an unlabelled pair sends a label: %s", raw)
	}
}

func TestArchetypesEndpointIsPublicReadOnlyAndNotFoundForOtherSets(t *testing.T) {
	h := ArchetypesHandler(NewListerWith(pairedSet(t), allFullExcept()))
	get := func(method, path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
		return rec
	}
	rec := get(http.MethodGet, "/draft/sets/syn/archetypes")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET = %d: %s", rec.Code, rec.Body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
	var got Archetypes
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || len(got.Pairs) != 10 {
		t.Fatalf("decode: %v %s", err, rec.Body)
	}
	// The wire names the client mirrors.
	var raw struct {
		Pairs []map[string]json.RawMessage `json:"pairs"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &raw)
	for _, key := range []string{"colors", "signposts"} {
		if _, ok := raw.Pairs[0][key]; !ok {
			t.Errorf("a pair has no %q", key)
		}
	}
	for _, key := range []string{"id", "oracle_id", "name", "mana_cost"} {
		var sp []map[string]json.RawMessage
		_ = json.Unmarshal(raw.Pairs[0]["signposts"], &sp)
		if _, ok := sp[0][key]; !ok {
			t.Errorf("a signpost has no %q", key)
		}
	}
	if rec := get(http.MethodGet, "/draft/sets/nope/archetypes"); rec.Code != http.StatusNotFound {
		t.Errorf("an unknown set = %d, want 404", rec.Code)
	}
	if rec := get(http.MethodPost, "/draft/sets/syn/archetypes"); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST = %d, want 405", rec.Code)
	}
}

// The labels file is the owner's to review, so its shape is held: every
// pair is one of the ten, in WUBRG order, and a label is two to four
// words with no stray space.
func TestArchetypeLabelsFileIsWellFormed(t *testing.T) {
	valid := map[string]bool{}
	for _, p := range Pairs {
		valid[p] = true
	}
	for set, pairs := range archetypeLabels {
		if set != strings.ToLower(set) || set == "" {
			t.Errorf("set key %q is not a lowercase code", set)
		}
		for pair, label := range pairs {
			if !valid[pair] {
				t.Errorf("%s: %q is not one of the ten pairs", set, pair)
			}
			if label != strings.TrimSpace(label) || strings.Contains(label, "  ") {
				t.Errorf("%s %s: label %q has stray space", set, pair, label)
			}
			if n := len(strings.Fields(label)); n < 2 || n > 4 {
				t.Errorf("%s %s: label %q is %d words, want two to four", set, pair, label, n)
			}
		}
	}
}
