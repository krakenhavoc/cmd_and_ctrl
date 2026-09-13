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

// oracleA / oracleB are fixed so a test can assert on a printing and
// the spec that claims it without either side guessing.
var (
	oracleA = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	oracleB = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	printA  = uuid.MustParse("aaaaaaaa-1111-1111-1111-111111111111")
	printB  = uuid.MustParse("bbbbbbbb-2222-2222-2222-222222222222")
)

func testIndex(t *testing.T, in ...cards.Card) *cards.Index {
	t.Helper()
	idx := cards.NewIndex()
	for _, c := range in {
		idx.Put(c)
	}
	return idx
}

func entryByName(t *testing.T, r Response, name string) Entry {
	t.Helper()
	for _, e := range r.Cards {
		if e.Name == name {
			return e
		}
	}
	t.Fatalf("no entry named %q in %d entries", name, len(r.Cards))
	return Entry{}
}

// A card whose spec says nothing about completeness must publish as
// unreviewed. This is the whole safety property: a card added on a
// branch that predates the field, or by someone who forgot it, must
// never be advertised as working in full.
func TestUnsetCompletenessPublishesAsUnreviewed(t *testing.T) {
	idx := testIndex(t, cards.Card{
		ID: printA, OracleID: oracleA, Name: "Silent Card",
		TypeLine: "Artifact", Layout: "normal",
	})
	got := build([]effects.Spec{{OracleID: oracleA.String(), Name: "Silent Card"}}, idx)

	e := entryByName(t, got, "Silent Card")
	if e.Completeness != "unreviewed" {
		t.Fatalf("completeness = %q, want unreviewed", e.Completeness)
	}
	if got.Counts.Full != 0 {
		t.Fatalf("an undeclared card counted as full: %+v", got.Counts)
	}
	if got.Counts.Unreviewed != 1 {
		t.Fatalf("unreviewed count = %d, want 1", got.Counts.Unreviewed)
	}
}

func TestFullAndCaveatsAreReportedSeparately(t *testing.T) {
	idx := testIndex(t,
		cards.Card{ID: printA, OracleID: oracleA, Name: "Whole Card", TypeLine: "Instant", Layout: "normal"},
		cards.Card{ID: printB, OracleID: oracleB, Name: "Partial Card", TypeLine: "Sorcery", Layout: "normal"},
	)
	got := build([]effects.Spec{
		{OracleID: oracleA.String(), Name: "Whole Card", Completeness: effects.CompletenessFull},
		{
			OracleID: oracleB.String(), Name: "Partial Card",
			Completeness: effects.CompletenessCaveats,
			Caveats:      []string{"Flashback is not implemented."},
		},
	}, idx)

	if got.Counts.Full != 1 || got.Counts.Caveats != 1 || got.Counts.Unreviewed != 0 {
		t.Fatalf("counts = %+v", got.Counts)
	}
	if cav := entryByName(t, got, "Partial Card").Caveats; len(cav) != 1 ||
		cav[0] != "Flashback is not implemented." {
		t.Fatalf("caveats = %v", cav)
	}
	if cav := entryByName(t, got, "Whole Card").Caveats; len(cav) != 0 {
		t.Fatalf("a full card carried caveats: %v", cav)
	}
}

// Sorted by name so the page has a stable order without sorting
// several hundred entries in the browser on every load.
func TestBuildSortsByName(t *testing.T) {
	idx := testIndex(t,
		cards.Card{ID: printA, OracleID: oracleA, Name: "Zombie Master", TypeLine: "Creature", Layout: "normal"},
		cards.Card{ID: printB, OracleID: oracleB, Name: "abrade", TypeLine: "Instant", Layout: "normal"},
	)
	got := build([]effects.Spec{
		{OracleID: oracleA.String(), Name: "Zombie Master"},
		{OracleID: oracleB.String(), Name: "abrade"},
	}, idx)
	if got.Cards[0].Name != "abrade" {
		t.Fatalf("sort is case-sensitive: %q came first", got.Cards[0].Name)
	}
}

// A modal double-faced card registers ONLY its land back, under the
// "<oracle_id>#1" key. It must publish as one card, not two, and the
// unautomated front face must be declared as a caveat — no card file
// can declare it, because no card file owns that face.
func TestFaceOnlyRegistrationBecomesOneCardWithACaveat(t *testing.T) {
	idx := testIndex(t, cards.Card{
		ID: printA, OracleID: oracleA, Layout: "modal_dfc",
		Name: "Akoum Warrior // Akoum Teeth", TypeLine: "Creature — Minotaur Warrior // Land",
		CardFaces: []cards.CardFace{
			{Name: "Akoum Warrior", TypeLine: "Creature — Minotaur Warrior"},
			{Name: "Akoum Teeth", TypeLine: "Land"},
		},
	})
	got := build([]effects.Spec{{
		OracleID:     oracleA.String() + "#1",
		Name:         "Akoum Teeth",
		Completeness: effects.CompletenessFull,
	}}, idx)

	if len(got.Cards) != 1 {
		t.Fatalf("got %d entries, want the two faces collapsed into 1", len(got.Cards))
	}
	e := got.Cards[0]
	if e.OracleID != oracleA.String() {
		t.Fatalf("oracle id kept the face suffix: %q", e.OracleID)
	}
	if e.Name != "Akoum Warrior // Akoum Teeth" {
		t.Fatalf("name = %q, want the printing's full name", e.Name)
	}
	if e.Completeness != "caveats" {
		t.Fatalf("completeness = %q — a half-automated card must not read as full", e.Completeness)
	}
	if len(e.Faces) != 2 || e.Faces[0].Automated || !e.Faces[1].Automated {
		t.Fatalf("face automation wrong: %+v", e.Faces)
	}
	if len(e.Caveats) != 1 {
		t.Fatalf("caveats = %v, want one naming the unautomated front", e.Caveats)
	}
}

// One unreviewed half makes the whole card unreviewed: a player
// reading both halves would conclude the same.
func TestCompletenessMergesPessimistically(t *testing.T) {
	idx := testIndex(t, cards.Card{
		ID: printA, OracleID: oracleA, Layout: "transform", Name: "Front // Back",
		CardFaces: []cards.CardFace{{Name: "Front"}, {Name: "Back"}},
	})
	got := build([]effects.Spec{
		{OracleID: oracleA.String(), Name: "Front", Completeness: effects.CompletenessFull},
		{OracleID: oracleA.String() + "#1", Name: "Back"},
	}, idx)
	if got.Cards[0].Completeness != "unreviewed" {
		t.Fatalf("completeness = %q, want unreviewed", got.Cards[0].Completeness)
	}
}

// A spec the loaded dump has no printing for still appears — the
// engine really does automate it — but is flagged rather than hidden,
// because a non-zero count means the dump is stale.
func TestMissingPrintingIsPublishedAndCounted(t *testing.T) {
	got := build([]effects.Spec{
		{OracleID: oracleA.String(), Name: "Unprinted Card", Completeness: effects.CompletenessFull},
	}, testIndex(t))

	e := entryByName(t, got, "Unprinted Card")
	if !e.Missing {
		t.Fatal("entry not flagged Missing")
	}
	if e.ScryfallID != "" {
		t.Fatalf("scryfall id = %q, want empty", e.ScryfallID)
	}
	if got.Counts.Missing != 1 || got.Counts.Total != 1 {
		t.Fatalf("counts = %+v", got.Counts)
	}
}

func TestBuildWithNilIndexStillPublishesNames(t *testing.T) {
	got := build([]effects.Spec{{OracleID: oracleA.String(), Name: "Sol Ring"}}, nil)
	if len(got.Cards) != 1 || got.Cards[0].Name != "Sol Ring" {
		t.Fatalf("entries = %+v", got.Cards)
	}
}

func TestPrimaryTypes(t *testing.T) {
	for _, tc := range []struct {
		line string
		want []string
	}{
		{"Legendary Artifact Creature — Golem", []string{"artifact", "creature"}},
		{"Instant", []string{"instant"}},
		{"Basic Land — Forest", []string{"land"}},
		// The subtype half must not leak in: "Legendary Creature —
		// Human Artificer" is not an artifact.
		{"Legendary Creature — Human Artificer", []string{"creature"}},
		// Split / modal lines answer for both halves, which is what
		// someone filtering for "land" expects of an MDFC.
		{"Creature — Minotaur Warrior // Land", []string{"creature", "land"}},
		{"Enchantment — Aura", []string{"enchantment"}},
		{"", nil},
	} {
		if got := primaryTypes(tc.line); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("primaryTypes(%q) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestSplitCatalogKey(t *testing.T) {
	for _, tc := range []struct {
		key      string
		wantID   string
		wantFace int
	}{
		{"abc", "abc", 0},
		{"abc#1", "abc", 1},
		{"abc#12", "abc", 12},
		// Not a face suffix; treated as part of the ID rather than
		// guessed at, so a malformed key shows as a missing printing
		// instead of silently merging into another card.
		{"abc#x", "abc#x", 0},
		{"abc#", "abc#", 0},
	} {
		id, face := splitCatalogKey(tc.key)
		if id != tc.wantID || face != tc.wantFace {
			t.Errorf("splitCatalogKey(%q) = (%q, %d), want (%q, %d)",
				tc.key, id, face, tc.wantID, tc.wantFace)
		}
	}
}

// The route must answer an anonymous request. It is a public
// showcase; a 401 here is the bug.
func TestCatalogRouteNeedsNoSession(t *testing.T) {
	srv := httptest.NewServer(Handler(cards.NewIndex(), nil))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/catalog")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 with no credential", resp.StatusCode)
	}
	var body Response
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// The real registry is linked into this test binary, so the live
	// catalog is what comes back. Assert on the invariants rather
	// than a count that changes every time a card lands.
	if body.Counts.Total == 0 {
		t.Fatal("catalog is empty — the effects blank import is missing")
	}
	if body.Counts.Full+body.Counts.Caveats+body.Counts.Unreviewed != body.Counts.Total {
		t.Fatalf("buckets do not partition the catalog: %+v", body.Counts)
	}
	for _, e := range body.Cards {
		if e.Completeness == "" {
			t.Fatalf("%q published with no completeness", e.Name)
		}
		if e.Completeness == "caveats" && len(e.Caveats) == 0 {
			t.Fatalf("%q says it has caveats but names none", e.Name)
		}
		if e.Completeness != "caveats" && len(e.Caveats) > 0 {
			t.Fatalf("%q lists caveats but reads as %q", e.Name, e.Completeness)
		}
	}
}

// The public image route is scoped to catalog cards. Anything else in
// the dump — and the dump is ~35k cards — must 404, which is what
// stops an unauthenticated route from being a general image proxy.
func TestImageRouteRefusesNonCatalogCards(t *testing.T) {
	idx := testIndex(t, cards.Card{
		ID: printA, OracleID: oracleA, Name: "Not In The Catalog",
		TypeLine: "Instant", Layout: "normal",
		ImageURIs: map[string]string{"normal": "https://cards.scryfall.io/normal/x.jpg"},
	})
	srv := httptest.NewServer(Handler(idx, nil))
	defer srv.Close()

	// The scope check runs before the availability check, so an
	// out-of-scope ID 404s rather than reporting on the cache.
	resp, err := http.Get(srv.URL + "/catalog/image/" + printA.String())
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 for a card the catalog does not register", resp.StatusCode)
	}
}

func TestImageRouteRejectsGarbageID(t *testing.T) {
	srv := httptest.NewServer(Handler(cards.NewIndex(), nil))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/catalog/image/not-a-uuid")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

// IsCatalogPrinting is the scope check itself. A real registered
// oracle ID must pass; an invented one must not.
func TestIsCatalogPrinting(t *testing.T) {
	all := effects.All()
	if len(all) == 0 {
		t.Skip("registry empty")
	}
	var oracle uuid.UUID
	for _, s := range all {
		id, _ := splitCatalogKey(s.OracleID)
		if parsed, err := uuid.Parse(id); err == nil {
			oracle = parsed
			break
		}
	}
	real := uuid.MustParse("cccccccc-3333-3333-3333-333333333333")
	fake := uuid.MustParse("dddddddd-4444-4444-4444-444444444444")
	idx := testIndex(t,
		cards.Card{ID: real, OracleID: oracle, Name: "Real", TypeLine: "Instant", Layout: "normal"},
		cards.Card{ID: fake, OracleID: oracleB, Name: "Fake", TypeLine: "Instant", Layout: "normal"},
	)
	if !IsCatalogPrinting(idx, real) {
		t.Error("a registered card's printing was refused")
	}
	if IsCatalogPrinting(idx, fake) {
		t.Error("an unregistered card's printing was allowed")
	}
	if IsCatalogPrinting(nil, real) {
		t.Error("nil index allowed a printing")
	}
}
