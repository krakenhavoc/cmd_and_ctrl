package deckcoverage

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func testingDeck(tc TestingCards) []deck.Entry {
	return []deck.Entry{
		{Name: tc.Commander, Count: 1, IsCommander: true},
		{Name: tc.Vanilla, Count: 1},
		{Name: tc.Basic, Count: 30},
		{Name: tc.Basic, Count: 2}, // a second row of the same card
		{Name: tc.Manual, Count: 1},
		{Name: tc.Automated, Count: 1},
		{Name: tc.Caveated, Count: 1},
		{Name: tc.Unreviewed, Count: 1},
		{Name: "No Such Card Anywhere", Count: 1},
		{Name: "no such card anywhere", Count: 1},
		{Name: "Sideboard Only Thing", Count: 1, IsSideboard: true},
	}
}

func TestBuildPutsEveryCardInItsBucket(t *testing.T) {
	idx, tc := TestingIndex(t)
	r, err := Build(idx, Deck{Name: "  Test Deck ", Source: "text", Entries: testingDeck(tc)})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if r.DeckName != "Test Deck" || r.Source != "text" || r.DeckKey != "" || r.SourceURL != "" {
		t.Errorf("header = %q %q %q %q", r.DeckName, r.Source, r.DeckKey, r.SourceURL)
	}

	want := map[string]Bucket{
		tc.Commander:  NoEffect,
		tc.Vanilla:    NoEffect,
		tc.Basic:      NoEffect,
		tc.Manual:     Manual,
		tc.Automated:  Automated,
		tc.Caveated:   Caveats,
		tc.Unreviewed: Unreviewed,
	}
	got := map[string]Bucket{}
	for _, c := range r.Cards {
		got[c.Name] = c.Bucket
		if c.OracleID == "" {
			t.Errorf("%s has no oracle id", c.Name)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("buckets = %v\nwant %v", got, want)
	}

	wantCounts := map[Bucket]int{Manual: 1, Unreviewed: 1, Caveats: 1, Automated: 1, NoEffect: 3}
	if !reflect.DeepEqual(r.Counts, wantCounts) {
		t.Errorf("counts = %v, want %v (by distinct card)", r.Counts, wantCounts)
	}
	if !r.Needed() {
		t.Error("a deck with a manual card should need a request")
	}

	for _, c := range r.Cards {
		switch c.Name {
		case tc.Basic:
			if c.Count != 32 {
				t.Errorf("Forest count = %d, want 32 across both rows", c.Count)
			}
		case tc.Caveated:
			if len(c.Caveats) == 0 || c.Caveats[0] != tc.CaveatText {
				t.Errorf("caveats = %v, want the catalogue's %q", c.Caveats, tc.CaveatText)
			}
		default:
			if len(c.Caveats) != 0 {
				t.Errorf("%s (%s) carries caveats %v", c.Name, c.Bucket, c.Caveats)
			}
		}
	}

	if !reflect.DeepEqual(r.Unknown, []string{"No Such Card Anywhere", "Sideboard Only Thing"}) {
		t.Errorf("unknown = %v", r.Unknown)
	}
	if !reflect.DeepEqual(r.Commanders, []string{tc.Commander}) {
		t.Errorf("commanders = %v", r.Commanders)
	}
}

// manual is game.Unimplemented, exactly: ADR 0095 adds no second
// definition of "unimplemented".
func TestManualIsExactlyUnimplemented(t *testing.T) {
	idx, tc := TestingIndex(t)
	r, err := Build(idx, Deck{Source: "text", Entries: testingDeck(tc)})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	for _, c := range r.Cards {
		printed, ok := idx.FindByName(c.Name)
		if !ok {
			t.Fatalf("%s not in the index", c.Name)
		}
		unimpl := game.Unimplemented(deck.ToGameCard(printed, false))
		if unimpl != (c.Bucket == Manual) {
			t.Errorf("%s: bucket %s but game.Unimplemented = %v", c.Name, c.Bucket, unimpl)
		}
	}
}

func TestBuildSortsByBucketThenName(t *testing.T) {
	idx, tc := TestingIndex(t)
	r, err := Build(idx, Deck{Source: "text", Entries: testingDeck(tc)})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	rank := map[Bucket]int{}
	for i, b := range Buckets {
		rank[b] = i
	}
	for i := 1; i < len(r.Cards); i++ {
		a, b := r.Cards[i-1], r.Cards[i]
		if rank[a.Bucket] > rank[b.Bucket] ||
			(a.Bucket == b.Bucket && strings.ToLower(a.Name) > strings.ToLower(b.Name)) {
			t.Errorf("out of order: %s/%s before %s/%s", a.Bucket, a.Name, b.Bucket, b.Name)
		}
	}
	if r.Cards[0].Bucket != Manual {
		t.Errorf("first card is %s, want the manual one first", r.Cards[0].Bucket)
	}
}

// Validation is reported, never blocking: a 40-card list with no
// commander still gets its buckets.
func TestViolationsAreReportedNotFatal(t *testing.T) {
	idx, tc := TestingIndex(t)
	r, err := Build(idx, Deck{Source: "text", Entries: []deck.Entry{
		{Name: tc.Basic, Count: 38},
		{Name: tc.Manual, Count: 2},
	}})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	codes := map[string]bool{}
	for _, v := range r.Violations {
		codes[v.Code] = true
	}
	for _, want := range []string{deck.CodeMissingCommander, deck.CodeWrongCardCount, deck.CodeSingleton} {
		if !codes[want] {
			t.Errorf("violations %v lack %s", r.Violations, want)
		}
	}
	if r.Counts[Manual] != 1 || r.Counts[NoEffect] != 1 {
		t.Errorf("counts = %v", r.Counts)
	}
}

func TestNothingToAdd(t *testing.T) {
	idx, tc := TestingIndex(t)
	r, err := Build(idx, Deck{Source: "text", Entries: []deck.Entry{
		{Name: tc.Commander, Count: 1, IsCommander: true},
		{Name: tc.Automated, Count: 1},
		{Name: tc.Caveated, Count: 1},
	}})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if r.Needed() {
		t.Errorf("a deck of automated, caveated and vanilla cards needs nothing: %v", r.Counts)
	}
	if got := r.CardsIn(Caveats); len(got) != 1 || got[0].Name != tc.Caveated {
		t.Errorf("CardsIn(caveats) = %v", got)
	}
}

// The JSON is what the public route serves: every bucket always
// present, arrays never null, and no art or oracle text anywhere.
func TestReportJSONShape(t *testing.T) {
	idx, tc := TestingIndex(t)
	r, err := Build(idx, Deck{Source: "text", Entries: []deck.Entry{{Name: tc.Automated, Count: 1}}})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	counts, _ := m["counts"].(map[string]any)
	for _, b := range Buckets {
		if _, ok := counts[string(b)]; !ok {
			t.Errorf("counts lacks %s: %s", b, raw)
		}
	}
	for _, k := range []string{"commanders", "cards", "unknown", "violations"} {
		if _, ok := m[k].([]any); !ok {
			t.Errorf("%s is not an array: %s", k, raw)
		}
	}
	for _, banned := range []string{"oracle_text", "image", "scryfall", "Do what this card does"} {
		if strings.Contains(string(raw), banned) {
			t.Errorf("report JSON contains %q: %s", banned, raw)
		}
	}
}

func TestBuildWithoutAnIndex(t *testing.T) {
	if _, err := Build(nil, Deck{}); err != ErrNoIndex {
		t.Errorf("err = %v, want ErrNoIndex", err)
	}
}
