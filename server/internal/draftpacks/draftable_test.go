package draftpacks

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// allFullExcept is a catalog verdict for synthetic sets: every card is
// Full but the named ones.
func allFullExcept(names ...string) IsFull {
	not := map[uuid.UUID]bool{}
	for _, n := range names {
		not[synthID("oracle", n)] = true
	}
	return func(id uuid.UUID) bool { return !not[id] }
}

func findSet(t *testing.T, resp Response, code string) SetStatus {
	t.Helper()
	for _, s := range resp.Sets {
		if s.Code == code {
			return s
		}
	}
	t.Fatalf("set %q is not in the list: %+v", code, resp.Sets)
	return SetStatus{}
}

func TestOneCardNotFullMakesASetNotDraftable(t *testing.T) {
	idx := cards.NewIndex()
	synthSet(idx, "syn", 12, 5, 3, 1, true)

	all := sets(idx, allFullExcept(), nil)
	s := findSet(t, all, "syn")
	if !s.Draftable || len(s.Blockers) != 0 {
		t.Fatalf("every card Full: %+v, want draftable with no blockers", s)
	}
	if s.Blockers == nil {
		t.Fatal("Blockers is nil; the client reads a list")
	}
	if s.Cards != 12+5+3+1+5 {
		t.Fatalf("Cards = %d, want %d", s.Cards, 12+5+3+1+5)
	}
	if all.Draftable != 1 {
		t.Fatalf("Response.Draftable = %d, want 1", all.Draftable)
	}

	// One card, of any rarity, is enough.
	for _, name := range []string{"syn Common 03", "syn Uncommon 00", "syn Rare 02", "syn Mythic 00"} {
		resp := sets(idx, allFullExcept(name), nil)
		s := findSet(t, resp, "syn")
		if s.Draftable {
			t.Errorf("%q is not Full and the set is still draftable", name)
		}
		if !reflect.DeepEqual(s.Blockers, []string{name}) {
			t.Errorf("blockers = %v, want [%s]", s.Blockers, name)
		}
		if resp.Draftable != 0 {
			t.Errorf("Response.Draftable = %d, want 0", resp.Draftable)
		}
		if PoolFor(idx, "syn").Draftable(allFullExcept(name)) {
			t.Errorf("Pool.Draftable is true with %q not Full", name)
		}
	}

	// Blockers are sorted by name.
	s = findSet(t, sets(idx, allFullExcept("syn Rare 01", "syn Common 07", "syn Mythic 00"), nil), "syn")
	if want := []string{"syn Common 07", "syn Mythic 00", "syn Rare 01"}; !reflect.DeepEqual(s.Blockers, want) {
		t.Fatalf("blockers = %v, want %v", s.Blockers, want)
	}
}

func TestBasicLandsNeverBlock(t *testing.T) {
	idx := cards.NewIndex()
	synthSet(idx, "syn", 12, 5, 3, 1, true)
	s := findSet(t, sets(idx, allFullExcept("Plains", "Island", "Swamp", "Mountain", "Forest"), nil), "syn")
	if !s.Draftable {
		t.Fatalf("basic lands block the set: %v", s.Blockers)
	}
}

func TestOnlyPackCardsBlock(t *testing.T) {
	idx := cards.NewIndex()
	n := synthSet(idx, "syn", 12, 5, 3, 1, true)
	out := synthCard("syn", "Not In Packs", "rare", fmt.Sprint(n))
	out.Booster = false
	idx.Put(out)
	s := findSet(t, sets(idx, allFullExcept("Not In Packs"), nil), "syn")
	if !s.Draftable {
		t.Fatalf("a card no pack contains blocks the set: %v", s.Blockers)
	}
}

func TestABonusSheetCardBlocks(t *testing.T) {
	idx := cards.NewIndex()
	synthSet(idx, "syn", 12, 5, 3, 1, true)
	bonus := synthCard("bon", "Bonus Card", "mythic", "1")
	bonus.SetType = "masterpiece"
	idx.Put(bonus)
	sheet := "bon"
	table := map[string]Override{"syn": {BonusSheet: &sheet}}

	s := findSet(t, sets(idx, allFullExcept("Bonus Card"), table), "syn")
	if s.Draftable || !reflect.DeepEqual(s.Blockers, []string{"Bonus Card"}) {
		t.Fatalf("with the bonus sheet in the recipe: %+v, want blocked by Bonus Card", s)
	}
	// Without the override the sheet is not in the set's packs.
	s = findSet(t, sets(idx, allFullExcept("Bonus Card"), nil), "syn")
	if !s.Draftable {
		t.Fatalf("without a bonus sheet the set is blocked by %v", s.Blockers)
	}
}

func TestListHoldsOnlySetsWithPacks(t *testing.T) {
	idx := cards.NewIndex()
	synthSet(idx, "old", 12, 5, 3, 1, true)
	synthSet(idx, "new", 12, 5, 3, 1, true)
	synthSet(idx, "norare", 12, 5, 0, 0, true)
	// A Commander precon: cards, but no booster to draft.
	n := synthSet(idx, "cmd", 12, 5, 3, 1, true)
	_ = n
	idx.PutSet(cards.SetInfo{Code: "old", Name: "Old Set", SetType: "expansion", ReleasedAt: "2025-01-01"})
	idx.PutSet(cards.SetInfo{Code: "new", Name: "New Set", SetType: "core", ReleasedAt: "2026-06-01"})
	idx.PutSet(cards.SetInfo{Code: "cmd", Name: "Precons", SetType: "commander", ReleasedAt: "2026-06-01"})

	resp := sets(idx, allFullExcept(), nil)
	var codes []string
	for _, s := range resp.Sets {
		codes = append(codes, s.Code)
	}
	if want := []string{"new", "old"}; !reflect.DeepEqual(codes, want) {
		t.Fatalf("listed sets = %v, want %v (newest first)", codes, want)
	}
	if s := findSet(t, resp, "new"); s.Name != "New Set" || s.Released != "2026-06-01" {
		t.Fatalf("set table not used for the name and date: %+v", s)
	}
}

func TestNilIndexListsNothing(t *testing.T) {
	if got := Sets(nil); got.Sets == nil || len(got.Sets) != 0 || got.Draftable != 0 {
		t.Fatalf("Sets(nil) = %+v, want an empty list", got)
	}
	if got := NewLister(nil).Sets(); got.Sets == nil || len(got.Sets) != 0 {
		t.Fatalf("NewLister(nil).Sets() = %+v, want an empty list", got)
	}
}

// --- the fixture dump, the real catalog, and the ledgers ---

const (
	fixtureDump    = "testdata/fixture_dump.json"
	fixtureLedger  = "testdata/draftable_fixture.txt"
	realDumpLedger = "testdata/draftable_sets.txt"
)

var updateDraftable = flag.Bool("update-draftable", false,
	"rewrite the draftable-set ledger: "+realDumpLedger+" when CMDCTRL_SCRYFALL_DUMP is set, "+fixtureLedger+" otherwise")

func loadFixture(t *testing.T) *cards.Index {
	t.Helper()
	idx := cards.NewIndex()
	if _, err := idx.Load(fixtureDump); err != nil {
		t.Fatalf("load %s: %v", fixtureDump, err)
	}
	return idx
}

// TestLoaderKeepsTheBoosterFields pins the three fields this package
// added to cards.Card to the dump's JSON names.
func TestLoaderKeepsTheBoosterFields(t *testing.T) {
	idx := loadFixture(t)
	var showcase, plain cards.Card
	idx.EachPrinting(func(c cards.Card) {
		if c.SetCode != "c2a" || c.Name != "Lightning Bolt" {
			return
		}
		if len(c.FrameEffects) > 0 {
			showcase = c
		} else {
			plain = c
		}
	})
	if plain.Rarity != "common" || !plain.Booster {
		t.Fatalf("plain Lightning Bolt: rarity %q booster %v, want common and true", plain.Rarity, plain.Booster)
	}
	if !reflect.DeepEqual(showcase.PromoTypes, []string{"boosterfun"}) {
		t.Fatalf("showcase Lightning Bolt promo_types = %v, want [boosterfun]", showcase.PromoTypes)
	}
}

func TestFixtureAgainstTheRealCatalog(t *testing.T) {
	idx := loadFixture(t)
	resp := Sets(idx)

	alpha := findSet(t, resp, "c2a")
	if !alpha.Draftable {
		t.Fatalf("c2a is not draftable; these cards are not Full in the catalog: %v", alpha.Blockers)
	}
	// 12 commons, 5 uncommons, 3 rares, 1 mythic, 5 basic lands. The
	// showcase reprint, the promo, the box topper and the token add none.
	if alpha.Cards != 26 {
		t.Fatalf("c2a has %d pack cards, want 26", alpha.Cards)
	}
	if alpha.Name != "C2 Fixture Alpha" || alpha.Released != "2026-01-01" {
		t.Fatalf("c2a name and date: %+v", alpha)
	}
	if alpha.Booster != KindPlay {
		t.Fatalf("c2a, released in 2026, uses the %q booster, want play", alpha.Booster)
	}

	beta := findSet(t, resp, "c2b")
	if beta.Draftable || !reflect.DeepEqual(beta.Blockers, []string{"C2 Fixture Unbuilt Card"}) {
		t.Fatalf("c2b: %+v, want blocked by exactly the one card the catalog does not hold", beta)
	}
	if resp.Draftable != 1 {
		t.Fatalf("Response.Draftable = %d, want 1", resp.Draftable)
	}

	// A pack from the draftable set holds only cards that are Full, or
	// basic lands, and the plain printing of each.
	isFull := CatalogIsFull(idx)
	pool := PoolFor(idx, "c2a")
	for seed := range uint64(300) {
		pack, err := pool.Pack(seeded(seed))
		if err != nil {
			t.Fatal(err)
		}
		if len(pack.Cards) != PlayBoosterSize {
			t.Fatalf("seed %d: %d cards", seed, len(pack.Cards))
		}
		for _, c := range pack.Cards {
			if !c.Basic && !isFull(c.OracleID) {
				t.Fatalf("seed %d: the pack holds %q, which is not Full", seed, c.Name)
			}
			if p, ok := idx.Get(c.ID); !ok || len(p.FrameEffects) > 0 || p.Promo {
				t.Fatalf("seed %d: %q is printing %s, not the plain booster printing", seed, c.Name, c.ID)
			}
		}
	}
}

func readLedger(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open the ledger: %v", err)
	}
	defer func() { _ = f.Close() }()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, strings.Fields(line)[0])
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func writeLedger(t *testing.T, path, source string, resp Response) {
	t.Helper()
	var b strings.Builder
	b.WriteString("# Sets recorded as draftable (ADR 0222 §4, #5437). GENERATED: do not edit.\n")
	b.WriteString("# Source: " + source + "\n")
	b.WriteString("# TestDraftableLedger fails when a set listed here stops being draftable.\n")
	b.WriteString("# Regenerate on purpose only; see testdata/README.md.\n")
	b.WriteString("# <set code> <pack cards> <name>\n")
	var lines []string
	for _, s := range resp.Sets {
		if s.Draftable {
			lines = append(lines, fmt.Sprintf("%s %d %s", s.Code, s.Cards, s.Name))
		}
	}
	sort.Strings(lines)
	for _, l := range lines {
		b.WriteString(l + "\n")
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %d draftable sets to %s", len(lines), path)
}

// TestDraftableLedger fails when a set recorded as draftable is no
// longer: a card its packs can contain stopped being Full in the
// catalog, or left the catalog. A draft of that set would hand someone
// a card that does not work.
//
// CI has no Scryfall dump, so it checks the fixture's ledger against the
// real catalog. With CMDCTRL_SCRYFALL_DUMP set (e2e-nightly.yml) it
// checks the real sets' ledger instead. A newly draftable set is logged,
// not failed: the ledger is a floor, and grows only when someone
// regenerates it with -update-draftable.
func TestDraftableLedger(t *testing.T) {
	idx, ledger, source := loadFixture(t), fixtureLedger, "the made-up sets of "+filepath.Base(fixtureDump)
	if dump := os.Getenv("CMDCTRL_SCRYFALL_DUMP"); dump != "" {
		idx = cards.NewIndex()
		if _, err := idx.Load(dump); err != nil {
			t.Fatalf("load %s: %v", dump, err)
		}
		// The set table, when it sits beside the dump as the server
		// expects, names the sets and their types.
		if _, err := idx.LoadSets(filepath.Join(filepath.Dir(dump), "sets.json")); err != nil && !os.IsNotExist(err) {
			t.Fatalf("load sets.json: %v", err)
		}
		ledger, source = realDumpLedger, "the Scryfall default-cards dump"
	}
	resp := Sets(idx)

	if *updateDraftable {
		writeLedger(t, ledger, source, resp)
		return
	}

	byCode := map[string]SetStatus{}
	for _, s := range resp.Sets {
		byCode[s.Code] = s
	}
	recorded := map[string]bool{}
	for _, code := range readLedger(t, ledger) {
		recorded[code] = true
		s, ok := byCode[code]
		switch {
		case !ok:
			t.Errorf("%s is recorded as draftable in %s, and the data no longer makes packs for it", code, ledger)
		case !s.Draftable:
			t.Errorf("%s (%s) is recorded as draftable in %s and has lost %d card(s) that are no longer Full in the catalog:\n  %s",
				code, s.Name, ledger, len(s.Blockers), strings.Join(s.Blockers, "\n  "))
		}
	}
	for _, s := range resp.Sets {
		if s.Draftable && !recorded[s.Code] {
			t.Logf("%s (%s) is draftable and not in %s yet; add it with -update-draftable", s.Code, s.Name, ledger)
		}
	}
}

// TestLedgerCatchesALostCard proves the ledger check bites: the fixture
// records c2a, and c2b, which is c2a plus one unbuilt card, is exactly
// what c2a would look like after losing a card.
func TestLedgerCatchesALostCard(t *testing.T) {
	recorded := readLedger(t, fixtureLedger)
	if !reflect.DeepEqual(recorded, []string{"c2a"}) {
		t.Fatalf("%s records %v, want [c2a]", fixtureLedger, recorded)
	}
	idx := loadFixture(t)
	full := CatalogIsFull(idx)
	bolt := uuid.MustParse("4457ed35-7c10-48c8-9776-456485fdf070")
	if !full(bolt) {
		t.Fatal("Lightning Bolt is not Full in the catalog; pick another fixture card")
	}
	lost := func(id uuid.UUID) bool { return id != bolt && full(id) }
	s := findSet(t, sets(idx, lost, overrides), "c2a")
	if s.Draftable || !reflect.DeepEqual(s.Blockers, []string{"Lightning Bolt"}) {
		t.Fatalf("with Lightning Bolt no longer Full, c2a is %+v, want blocked by it", s)
	}
}

// --- the endpoint ---

func TestDraftSetsEndpoint(t *testing.T) {
	h := Handler(loadFixture(t))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/draft/sets", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /draft/sets = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=300" {
		t.Errorf("Cache-Control = %q", cc)
	}
	var got Response
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v\n%s", err, rec.Body.String())
	}
	if !reflect.DeepEqual(got, Sets(loadFixture(t))) {
		t.Fatalf("body = %+v, want Sets()", got)
	}
	// The wire names the client mirrors.
	var raw struct {
		Sets []map[string]json.RawMessage `json:"sets"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"code", "name", "released", "booster", "cards", "draftable", "blockers"} {
		if _, ok := raw.Sets[0][key]; !ok {
			t.Errorf("a set has no %q field", key)
		}
	}

	// Read-only: nothing but GET (and HEAD) is answered.
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(method, "/draft/sets", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s /draft/sets = %d, want 405", method, rec.Code)
		}
	}
}

func TestDraftSetsEndpointWithNoDump(t *testing.T) {
	for name, idx := range map[string]*cards.Index{"nil index": nil, "empty index": cards.NewIndex()} {
		rec := httptest.NewRecorder()
		Handler(idx).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/draft/sets", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d", name, rec.Code)
		}
		if body := strings.TrimSpace(rec.Body.String()); body != `{"sets":[],"draftable":0}` {
			t.Errorf("%s: body = %s", name, body)
		}
	}
}

func TestListerRecomputesOnlyWhenTheIndexChanges(t *testing.T) {
	idx := cards.NewIndex()
	synthSet(idx, "one", 12, 5, 3, 1, true)
	l := NewLister(idx)
	first := l.Sets()
	if len(first.Sets) != 1 {
		t.Fatalf("%d sets, want 1", len(first.Sets))
	}
	if again := l.Sets(); &again.Sets[0] != &first.Sets[0] {
		t.Fatal("an unchanged index was listed twice")
	}
	synthSet(idx, "two", 12, 5, 3, 1, true)
	if got := l.Sets(); len(got.Sets) != 2 {
		t.Fatalf("after the index grew: %d sets, want 2", len(got.Sets))
	}
}

// The Lister hands a draft the pool of a draftable set only, with the
// verdict GET /draft/sets shows, and knows which printings those pools
// hold (#5438).
func TestListerHandsOutDraftablePoolsOnly(t *testing.T) {
	idx := cards.NewIndex()
	synthSet(idx, "aaa", 20, 20, 10, 5, true)
	synthSet(idx, "bbb", 20, 20, 10, 5, false)
	l := NewListerWith(idx, allFullExcept("bbb Common 03"))

	pool, status, ok := l.DraftablePool(" AAA ")
	if !ok || !status.Draftable || status.Code != "aaa" || pool.Set != "aaa" || !pool.CanFill() {
		t.Fatalf("a draftable set: ok %v, status %+v", ok, status)
	}
	if _, _, ok := l.DraftablePool("bbb"); ok {
		t.Fatal("a set with a card that is not Full was handed out")
	}
	if _, _, ok := l.DraftablePool("zzz"); ok {
		t.Fatal("an unknown set was handed out")
	}
	var nilLister *Lister
	if _, _, ok := nilLister.DraftablePool("aaa"); ok || nilLister.IsDraftPrinting(uuid.New()) {
		t.Fatal("a nil lister hands something out")
	}

	// Every card of the draftable pool is a draft printing, basic lands
	// included; no card of the blocked set is.
	for _, c := range pool.Contents() {
		if !l.IsDraftPrinting(c.ID) {
			t.Fatalf("%s is in a draftable pool but is not a draft printing", c.Name)
		}
	}
	for _, c := range PoolFor(idx, "bbb").Contents() {
		if l.IsDraftPrinting(c.ID) {
			t.Fatalf("%s is a draft printing, but its set cannot be drafted", c.Name)
		}
	}
	if l.IsDraftPrinting(uuid.New()) {
		t.Fatal("an unknown id is a draft printing")
	}
}

// GET /draft/image/{id} is public, so it is scoped: a printing outside
// every draftable pool is a 404 whether or not the dump holds it, and
// the scope is checked before the image cache is.
func TestDraftImageRouteServesOnlyDraftPrintings(t *testing.T) {
	idx := cards.NewIndex()
	synthSet(idx, "aaa", 20, 20, 10, 5, true)
	synthSet(idx, "bbb", 20, 20, 10, 5, false)
	l := NewListerWith(idx, allFullExcept("bbb Common 03"))
	h := ImageHandler(l, idx, nil)
	get := func(path string) int {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec.Code
	}
	in := PoolFor(idx, "aaa").Contents()[0].ID
	out := PoolFor(idx, "bbb").Contents()[0].ID
	if got := get("/draft/image/" + out.String()); got != http.StatusNotFound {
		t.Errorf("a card of a set that cannot be drafted: %d, want 404", got)
	}
	if got := get("/draft/image/" + uuid.NewString()); got != http.StatusNotFound {
		t.Errorf("an unknown card: %d, want 404", got)
	}
	if got := get("/draft/image/not-an-id"); got != http.StatusBadRequest {
		t.Errorf("a malformed id: %d, want 400", got)
	}
	// In scope, with no image cache on this deployment: 503, not 404.
	if got := get("/draft/image/" + in.String()); got != http.StatusServiceUnavailable {
		t.Errorf("a draft printing with no cache: %d, want 503", got)
	}
}
