package catalog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fixed ids for the set fixtures. Printing ids are only there to keep
// the index's printings apart.
var (
	oracleC = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	oracleD = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	oracleE = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	oracleF = uuid.MustParse("66666666-6666-6666-6666-666666666666")
)

func printing(n int, oracle uuid.UUID, name, set, setName, released string) cards.Card {
	return cards.Card{
		ID:         uuid.MustParse("00000000-0000-0000-0000-0000000000" + twoDigits(n)),
		OracleID:   oracle,
		Name:       name,
		SetCode:    set,
		SetName:    setName,
		SetType:    "expansion",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		ReleasedAt: released,
	}
}

func twoDigits(n int) string {
	return string([]byte{byte('0' + n/10), byte('0' + n%10)})
}

func setByCode(t *testing.T, r Response, code string) SetSummary {
	t.Helper()
	for _, s := range r.Sets {
		if s.Code == code {
			return s
		}
	}
	t.Fatalf("no set %q in %+v", code, r.Sets)
	return SetSummary{}
}

// The per-set tally counts unique cards, splits the catalogued ones by
// completeness, and calls the rest unsupported.
func TestSetCountsSplitByCompleteness(t *testing.T) {
	idx := testIndex(t,
		// spm: A (full), B (caveats), C (unreviewed), D (not catalogued)
		printing(1, oracleA, "Card A", "spm", "Spider Set", "2026-09-26"),
		printing(2, oracleB, "Card B", "spm", "Spider Set", "2026-09-26"),
		printing(3, oracleC, "Card C", "spm", "Spider Set", "2026-09-26"),
		printing(4, oracleD, "Card D", "spm", "Spider Set", "2026-09-26"),
		// A is reprinted in an older set.
		printing(5, oracleA, "Card A", "old", "Old Set", "2020-01-01"),
	)
	got := buildWithSets([]effects.Spec{
		{OracleID: oracleA.String(), Name: "Card A", Completeness: effects.CompletenessFull},
		{
			OracleID: oracleB.String(), Name: "Card B",
			Completeness: effects.CompletenessCaveats, Caveats: []string{"Not modelled."},
		},
		{OracleID: oracleC.String(), Name: "Card C"},
	}, idx)

	spm := setByCode(t, got, "spm")
	want := SetSummary{
		Code: "spm", Name: "Spider Set", Released: "2026-09-26",
		Total: 4, Full: 1, Caveats: 1, Unreviewed: 1, Unsupported: 1,
	}
	if spm != want {
		t.Fatalf("spm = %+v, want %+v", spm, want)
	}
	old := setByCode(t, got, "old")
	if old.Total != 1 || old.Full != 1 || old.Unsupported != 0 {
		t.Fatalf("old = %+v", old)
	}
	for _, s := range got.Sets {
		if s.Full+s.Caveats+s.Unreviewed+s.Unsupported != s.Total {
			t.Fatalf("%s does not partition its total: %+v", s.Code, s)
		}
	}
}

// Several printings of one card in one set are one card.
func TestSetCountsAreByOracleNotPrinting(t *testing.T) {
	idx := testIndex(t,
		printing(1, oracleA, "Card A", "spm", "Spider Set", "2026-09-26"),
		printing(2, oracleA, "Card A", "spm", "Spider Set", "2026-09-26"),
		printing(3, oracleA, "Card A", "spm", "Spider Set", "2026-09-26"),
	)
	got := buildWithSets([]effects.Spec{
		{OracleID: oracleA.String(), Name: "Card A", Completeness: effects.CompletenessFull},
	}, idx)
	if s := setByCode(t, got, "spm"); s.Total != 1 || s.Full != 1 {
		t.Fatalf("spm = %+v, want one card", s)
	}
}

// Tokens, emblems, art-series cards, basic lands, planes, digital-only
// printings and memorabilia are not cards a set "has" for this tally.
func TestSetCountsLeaveOutNonCards(t *testing.T) {
	real := printing(1, oracleA, "Real Card", "spm", "Spider Set", "2026-09-26")

	token := printing(2, oracleB, "A Token", "spm", "Spider Set", "2026-09-26")
	token.Layout = "token"

	dfcToken := printing(3, oracleC, "Incubator", "spm", "Spider Set", "2026-09-26")
	dfcToken.Layout = "double_faced_token"

	emblem := printing(4, oracleD, "An Emblem", "spm", "Spider Set", "2026-09-26")
	emblem.Layout = "emblem"
	emblem.TypeLine = "Emblem"

	art := printing(5, oracleE, "Art Card", "spm", "Spider Set", "2026-09-26")
	art.Layout = "art_series"
	art.SetType = "art_series"

	basic := printing(6, oracleF, "Forest", "spm", "Spider Set", "2026-09-26")
	basic.TypeLine = "Basic Land — Forest"

	snow := printing(7, uuid.MustParse("77777777-7777-7777-7777-777777777777"), "Snow-Covered Forest", "spm", "Spider Set", "2026-09-26")
	snow.TypeLine = "Basic Snow Land — Forest"

	plane := printing(8, uuid.MustParse("88888888-8888-8888-8888-888888888888"), "A Plane", "spm", "Spider Set", "2026-09-26")
	plane.Layout = "planar"

	digital := printing(9, uuid.MustParse("99999999-9999-9999-9999-999999999999"), "A-Digital", "spm", "Spider Set", "2026-09-26")
	digital.Digital = true

	memo := printing(10, uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), "Memo", "spm", "Spider Set", "2026-09-26")
	memo.SetType = "memorabilia"

	// A set made only of tokens is not listed at all.
	tokenSet := printing(11, uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"), "Soldier", "tspm", "Spider Tokens", "2026-09-26")
	tokenSet.Layout = "token"
	tokenSet.SetType = "token"

	// A card whose NAME starts with "Basic" is not a basic land: only
	// the type line decides.
	almost := printing(12, uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"), "Basic Instinct", "spm", "Spider Set", "2026-09-26")
	almost.TypeLine = "Sorcery"

	idx := testIndex(t, real, token, dfcToken, emblem, art, basic, snow, plane, digital, memo, tokenSet, almost)
	got := buildWithSets([]effects.Spec{
		{OracleID: oracleA.String(), Name: "Real Card", Completeness: effects.CompletenessFull},
	}, idx)

	if s := setByCode(t, got, "spm"); s.Total != 2 || s.Full != 1 || s.Unsupported != 1 {
		t.Fatalf("spm = %+v, want 2 cards (Real Card, Basic Instinct)", s)
	}
	for _, s := range got.Sets {
		if s.Code == "tspm" {
			t.Fatalf("a set of only tokens was listed: %+v", s)
		}
	}
}

// Newest set first, and a card lists its sets in that same order.
func TestSetsAreNewestFirstAndEntriesListThem(t *testing.T) {
	idx := testIndex(t,
		printing(1, oracleA, "Card A", "mid", "Middle", "2022-05-05"),
		printing(2, oracleA, "Card A", "new", "Newest", "2026-09-26"),
		printing(3, oracleA, "Card A", "old", "Oldest", "2015-01-01"),
		printing(4, oracleB, "Card B", "new", "Newest", "2026-09-26"),
	)
	got := buildWithSets([]effects.Spec{
		{OracleID: oracleA.String(), Name: "Card A", Completeness: effects.CompletenessFull},
	}, idx)

	var codes []string
	for _, s := range got.Sets {
		codes = append(codes, s.Code)
	}
	if want := []string{"new", "mid", "old"}; !reflect.DeepEqual(codes, want) {
		t.Fatalf("set order = %v, want %v", codes, want)
	}
	if sets := entryByName(t, got, "Card A").Sets; !reflect.DeepEqual(sets, []string{"new", "mid", "old"}) {
		t.Fatalf("Card A sets = %v", sets)
	}
}

// A stray promo date on one card does not move the set's release date.
func TestSetReleaseDateIsTheCommonOne(t *testing.T) {
	early := printing(1, oracleA, "Prerelease", "spm", "Spider Set", "2026-09-19")
	idx := testIndex(t,
		early,
		printing(2, oracleB, "Card B", "spm", "Spider Set", "2026-09-26"),
		printing(3, oracleC, "Card C", "spm", "Spider Set", "2026-09-26"),
	)
	got := buildWithSets(nil, idx)
	if s := setByCode(t, got, "spm"); s.Released != "2026-09-26" {
		t.Fatalf("released = %q, want the common date", s.Released)
	}
}

// With no dump there are no sets, and build's own output is untouched.
func TestNoIndexNoSets(t *testing.T) {
	got := buildWithSets([]effects.Spec{{OracleID: oracleA.String(), Name: "Card A"}}, nil)
	if got.Sets != nil || got.Cards[0].Sets != nil {
		t.Fatalf("sets without a dump: %+v", got)
	}
}

// Build alone stays set-free, so its other callers do not pay the pass.
func TestBuildAloneHasNoSets(t *testing.T) {
	idx := testIndex(t, printing(1, oracleA, "Card A", "spm", "Spider Set", "2026-09-26"))
	got := build([]effects.Spec{{OracleID: oracleA.String(), Name: "Card A"}}, idx)
	if got.Sets != nil || got.Cards[0].Sets != nil {
		t.Fatalf("Build carried sets: %+v", got)
	}
}

// GET /catalog serves the sets, and they survive the JSON round trip
// under the names the client reads.
func TestCatalogRouteServesSets(t *testing.T) {
	idx := testIndex(t, printing(1, oracleA, "Card A", "spm", "Spider Set", "2026-09-26"))
	srv := httptest.NewServer(Handler(idx, nil))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/catalog")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body struct {
		Sets []map[string]any `json:"sets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Sets) != 1 {
		t.Fatalf("sets = %v, want one", body.Sets)
	}
	for _, key := range []string{"code", "name", "released", "total", "full", "caveats", "unreviewed", "unsupported"} {
		if _, ok := body.Sets[0][key]; !ok {
			t.Errorf("set is missing %q: %v", key, body.Sets[0])
		}
	}
}
