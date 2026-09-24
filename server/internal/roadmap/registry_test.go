package roadmap

import (
	"encoding/json"
	"flag"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/coverage"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// registry_test.go — what CI holds the registry to. The registry is
// curated by hand; these are the parts of each judgement a program can
// check against the live catalog, so the public page cannot quietly
// stop being true.

var update = flag.Bool("update", false, "rewrite the generated open-seams table in docs/engine-seams.md")

var slugShape = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func TestItemsAreWellFormed(t *testing.T) {
	slugs := map[string]bool{}
	pins := map[int]string{}
	for _, it := range items {
		if !slugShape.MatchString(it.Slug) {
			t.Errorf("%q: slug is not lowercase-with-dashes", it.Slug)
		}
		if slugs[it.Slug] {
			t.Errorf("%q: duplicate slug", it.Slug)
		}
		slugs[it.Slug] = true
		if it.Name == "" {
			t.Errorf("%s: no name", it.Slug)
		}
		switch it.Kind {
		case KindKeyword, KindMechanic, KindSeam:
		default:
			t.Errorf("%s: unknown kind %q", it.Slug, it.Kind)
		}
		switch it.Status {
		case StatusImplemented, StatusPartial, StatusMissing:
		default:
			t.Errorf("%s: unknown status %q", it.Slug, it.Status)
		}
		if it.Pin < 0 {
			t.Errorf("%s: negative pin", it.Slug)
		}
		if it.Pin > 0 {
			if other, dup := pins[it.Pin]; dup {
				t.Errorf("%s: pin %d is already %s's", it.Slug, it.Pin, other)
			}
			pins[it.Pin] = it.Slug
		}
		if it.Printed != "" {
			if _, err := regexp.Compile(it.Printed); err != nil {
				t.Errorf("%s: Printed does not compile: %v", it.Slug, err)
			}
		}
		if it.ADR != "" {
			if _, err := os.Stat("../../../docs/decisions/" + it.ADR); err != nil {
				t.Errorf("%s: ADR %q does not exist", it.Slug, it.ADR)
			}
		}
	}
}

// Every keyword the engine enforces is on the page exactly once, and a
// keyword item names only tokens the engine enforces.
func TestEveryCanonicalKeywordHasOneItem(t *testing.T) {
	canonical := map[string]bool{}
	for _, kw := range game.CanonicalKeywordTable() {
		canonical[kw] = true
	}
	owner := map[string]string{}
	for _, it := range items {
		if it.Kind != KindKeyword {
			if len(it.Keywords) > 0 {
				t.Errorf("%s: only a keyword item may claim keyword tokens", it.Slug)
			}
			continue
		}
		if len(it.Keywords) == 0 {
			t.Errorf("%s: a keyword item must name the tokens it covers", it.Slug)
		}
		for _, kw := range it.Keywords {
			if !canonical[kw] {
				t.Errorf("%s: %q is not in game.CanonicalKeywordTable — a keyword joins the page when it joins the table", it.Slug, kw)
			}
			if other, dup := owner[kw]; dup {
				t.Errorf("%q is claimed by both %s and %s", kw, other, it.Slug)
			}
			owner[kw] = it.Slug
		}
	}
	for kw := range canonical {
		if _, ok := owner[kw]; !ok {
			t.Errorf("keyword %q is enforced by the engine and has no roadmap item", kw)
		}
	}
}

// Every row of the coverage package's curated mechanic table is on the
// page, and every item that borrows a row names one that exists.
func TestEveryCoverageMechanicHasAnItem(t *testing.T) {
	used := map[string]string{}
	for _, it := range items {
		if it.Mechanic == "" {
			continue
		}
		if _, ok := mechanicByName(it.Mechanic); !ok {
			t.Errorf("%s: Mechanic %q is not a row of coverage.Mechanics()", it.Slug, it.Mechanic)
		}
		if other, dup := used[it.Mechanic]; dup {
			t.Errorf("mechanic %q is borrowed by both %s and %s", it.Mechanic, other, it.Slug)
		}
		used[it.Mechanic] = it.Slug
	}
	for _, m := range coverage.Mechanics() {
		if _, ok := used[m.Name]; !ok {
			t.Errorf("coverage mechanic %q has no roadmap item", m.Name)
		}
	}
}

func entryBySlug(t *testing.T, r Roadmap) map[string]Entry {
	t.Helper()
	out := map[string]Entry{}
	for _, e := range r.Items {
		out[e.Slug] = e
	}
	return out
}

// An implemented or partial item has to point at a card that shows it
// working — or say why no such card can exist. And an excuse that has
// stopped being needed is stale, so it fails too.
func TestWorkingItemsHaveACompleteExample(t *testing.T) {
	entries := entryBySlug(t, Build())
	for _, it := range items {
		e := entries[it.Slug]
		if it.Status == StatusMissing {
			if it.NoCatalogExample != "" {
				t.Errorf("%s: a missing item needs no NoCatalogExample excuse", it.Slug)
			}
			continue
		}
		switch {
		case len(e.Examples) == 0 && it.NoCatalogExample == "":
			t.Errorf("%s is %s but no complete catalog card uses it.\n"+
				"Give it a probe (Probe, Printed or Mechanic) that finds one, curate one in Examples,\n"+
				"or, if no card file can show it (a vanilla keyword creature needs none), set NoCatalogExample to say why.",
				it.Slug, it.Status)
		case len(e.Examples) > 0 && it.NoCatalogExample != "":
			t.Errorf("%s: NoCatalogExample says %q, but %s now shows it — drop the excuse",
				it.Slug, it.NoCatalogExample, e.Examples[0].Name)
		}
	}
}

// A curated example must be a real, complete card that really uses the
// item, or the page would feature a card that does not show it.
func TestCuratedExamplesAreCompleteAndUseTheItem(t *testing.T) {
	cs := loadCards()
	for _, it := range items {
		m := compile(it)
		for _, name := range it.Examples {
			c := cs.byName[name]
			switch {
			case c == nil:
				t.Errorf("%s: curated example %q is not in the catalog", it.Slug, name)
			case !c.full():
				t.Errorf("%s: curated example %q is in the catalog as %s, not complete", it.Slug, name, c.verdict)
			case m.hasProbe() && !m.uses(c):
				t.Errorf("%s: curated example %q does not satisfy the item's own probe", it.Slug, name)
			}
		}
	}
}

// A card listed as waiting on an item must not already be registered
// as complete. When one is, the gap has probably closed.
func TestWaitingCardsAreNotComplete(t *testing.T) {
	cs := loadCards()
	for _, it := range items {
		if len(it.Waiting) > 0 && it.Status == StatusImplemented {
			t.Errorf("%s: an implemented item has nothing waiting on it — move Waiting to the item the cards still need", it.Slug)
		}
		for _, name := range it.Waiting {
			if c := cs.byName[name]; c != nil && c.full() {
				t.Errorf("%s: %q is waiting on this item, but it is registered as complete.\n"+
					"The seam may have closed. Check the code, then either take the card off\n"+
					"Waiting or mark the item implemented and move its prose to Closed seams.",
					it.Slug, name)
			}
		}
	}
}

func TestUnfinishedItemsSayWhatIsMissing(t *testing.T) {
	for _, it := range items {
		if it.Status == StatusImplemented {
			if it.Missing != "" {
				t.Errorf("%s: an implemented item has nothing missing — drop Missing, or mark it partial", it.Slug)
			}
			continue
		}
		if it.Issue <= 0 {
			t.Errorf("%s is %s and has no tracking Issue", it.Slug, it.Status)
		}
		if it.Missing == "" {
			t.Errorf("%s is %s and does not say what is missing", it.Slug, it.Status)
		}
		if it.Kind == KindSeam && it.EngineNotes == "" {
			t.Errorf("%s: an open seam needs EngineNotes for engine-seams.md", it.Slug)
		}
	}
}

// Everything a player reads is held to the Caveats tone rule, through
// the same function the catalog's caveats go through.
func TestPlayerFacingTextReadsAsSentences(t *testing.T) {
	for _, it := range items {
		for _, p := range effects.PlayerFacingProblems(it.Summary) {
			t.Errorf("%s: Summary %s: %q", it.Slug, p, it.Summary)
		}
		if it.Missing != "" {
			for _, p := range effects.PlayerFacingProblems(it.Missing) {
				t.Errorf("%s: Missing %s: %q", it.Slug, p, it.Missing)
			}
		}
		if it.NoCatalogExample != "" {
			for _, p := range effects.PlayerFacingProblems(it.NoCatalogExample) {
				t.Errorf("%s: NoCatalogExample %s: %q", it.Slug, p, it.NoCatalogExample)
			}
		}
	}
}

// Build is plain data a handler can serialise, and it publishes names
// and caveat sentences only.
func TestBuildIsPlainPublishableData(t *testing.T) {
	r := Build()
	if got, want := r.Counts.Implemented+r.Counts.Partial+r.Counts.Missing, len(items); got != want {
		t.Fatalf("counts add up to %d, registry has %d items", got, want)
	}
	if len(r.Items) != len(items) {
		t.Fatalf("Build published %d items, registry has %d", len(r.Items), len(items))
	}
	body, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, banned := range []string{`"image`, `"oracle_text"`, `"scryfall_id"`, `"engine_notes"`} {
		if strings.Contains(string(body), banned) {
			t.Errorf("the published roadmap carries %s", banned)
		}
	}
	for _, e := range r.Items {
		if len(e.Examples) > MaxExamples || len(e.PartialExamples) > MaxPartialExamples {
			t.Errorf("%s: over the example limits", e.Slug)
		}
	}
	if len(r.Next) > MaxNext {
		t.Errorf("up next has %d items, limit %d", len(r.Next), MaxNext)
	}
}

func TestNextPutsPinsFirstThenRanksByCardsUnblocked(t *testing.T) {
	reg := []Item{
		{Slug: "done", Status: StatusImplemented, Unblocks: 99},
		{Slug: "small", Status: StatusMissing, Unblocks: 1},
		{Slug: "big", Status: StatusPartial, Unblocks: 10, Waiting: []string{"A", "B"}},
		{Slug: "pinned-second", Status: StatusMissing, Pin: 2},
		{Slug: "pinned-first", Status: StatusImplemented, Pin: 1},
		{Slug: "nothing-waiting", Status: StatusMissing},
		{Slug: "mid", Status: StatusMissing, Waiting: []string{"A", "B", "C"}},
	}
	got := strings.Join(next(reg), ",")
	if want := "pinned-first,pinned-second,big,mid,small"; got != want {
		t.Fatalf("next = %s, want %s", got, want)
	}
}

// Not a failure: the caveat cards no registry item explains. Each is a
// gap the page does not yet name, so the list is where the next
// registry entry comes from.
func TestReportCaveatsNoItemExplains(t *testing.T) {
	cs := loadCards()
	r := Build()
	explained := map[string]bool{}
	for _, e := range r.Items {
		for _, pe := range e.PartialExamples {
			explained[pe.Name] = true
		}
	}
	matchers := make([]matcher, 0, len(items))
	for _, it := range items {
		matchers = append(matchers, compile(it))
		for _, w := range it.Waiting {
			explained[w] = true
		}
	}
	var loose []string
	for _, c := range cs.cards {
		if len(c.caveats) == 0 || explained[c.name] {
			continue
		}
		hit := false
		for _, m := range matchers {
			for _, cv := range c.caveats {
				if m.phrases != nil && m.phrases(cv) {
					hit = true
				}
			}
		}
		if !hit {
			loose = append(loose, c.name+": "+c.caveats[0])
		}
	}
	sort.Strings(loose)
	t.Logf("%d caveat cards name no roadmap item (report only):", len(loose))
	for i, l := range loose {
		if i == 40 {
			t.Logf("  … and %d more", len(loose)-40)
			break
		}
		t.Logf("  %s", l)
	}
}

// The open table in docs/engine-seams.md is generated from the
// registry. It depends on the registry alone, so a stale table always
// means the PR that edited registry.go did not regenerate it.
func TestSeamsTableIsCurrent(t *testing.T) {
	raw, err := os.ReadFile(SeamsDocPath)
	if err != nil {
		t.Fatalf("read %s: %v", SeamsDocPath, err)
	}
	doc := string(raw)
	want := SeamsBlock()
	got, ok := ExtractSeams(doc)
	if !ok {
		t.Fatalf("%s has lost its generated open-seams markers; nothing is checking the table.\n"+
			"Put the two marker lines back under \"## Open seams\" and run:\n\n"+
			"  go test ./internal/roadmap/ -update", SeamsDocPath)
	}
	if got == want {
		return
	}
	if *update {
		next, ok := SpliceSeams(doc, want)
		if !ok {
			t.Fatalf("markers vanished between read and splice in %s", SeamsDocPath)
		}
		if err := os.WriteFile(SeamsDocPath, []byte(next), 0o644); err != nil {
			t.Fatalf("write %s: %v", SeamsDocPath, err)
		}
		t.Logf("updated the open-seams table in %s", SeamsDocPath)
		return
	}
	t.Errorf("the open-seams table in %s is out of date with server/internal/roadmap/registry.go.\n\n"+
		"Regenerate it (this rewrites only the block between the markers):\n\n"+
		"  go test ./internal/roadmap/ -update", SeamsDocPath)
}
