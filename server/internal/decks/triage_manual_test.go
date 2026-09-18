package decks

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/catalog"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
)

// triage_manual_test.go — "how much of THIS deck does the engine
// actually play?", for any decklist, against the live registry.
//
// # Why this exists
//
// docs/decklists/ holds three triages of real decks, and all three
// say the same thing in their opening paragraph: *a deck is a better
// forcing function than a card count, because it says which gaps
// actually stop a game from being played.* That judgement is right and
// it is the reason those files were written.
//
// Every one of them was then produced by a THROWAWAY script — a probe
// test deleted before commit, plus a one-off Python pass over the
// dump. Nothing was left behind that could be run again. The
// predictable thing happened:
//
//   - aang-is-so-flashy.md reports "23 of 68 nonland in the catalog",
//     measured against a 289-spec registry. The registry now holds
//     more than 1,700. The number is not stale at the edges, it is
//     meaningless.
//   - hashaton-the-cheater.md reports 47 of 100, measured five days
//     later against a different registry again.
//   - The two cannot be compared with each other, and neither can be
//     compared with today.
//
// internal/cards/coverage exists for exactly this failure mode on the
// catalog's own numbers — "the documentation about it cannot quietly
// stop being true". This file is that idea pointed at a decklist.
//
// # What it measures, and with whose arithmetic
//
// The join is decks.coverage's, which is catalog.Build's: completeness
// is merged across faces, so a card whose back half is caveated is a
// caveat card, and an MDFC whose front face has no spec earns a caveat
// no card file declares. Recomputing that here would eventually
// disagree with /catalog about the same card — the mistake coverage.go
// already calls out — so this asks the same builder the same question.
//
// The one thing it adds is UNREGISTERED, broken out rather than folded
// into "imperfect". For a pre-built deck the distinction barely
// matters, because a curated deck has no unregistered cards by
// construction. For a real deck off Archidekt it is the whole answer:
// a caveated card plays and skips a clause, an unregistered card is a
// piece of cardboard the players move by hand.
//
// # Running it
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json \
//	  go test ./internal/decks/ -run TriageDecks -v
//
// With no CMDCTRL_TRIAGE_DECKS it profiles the four curated decks.
// Point it at a file or a directory of decklists to profile those
// instead — the format is the importer's own, so anything the lobby
// accepts works here:
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json \
//	CMDCTRL_TRIAGE_DECKS=/path/to/decks \
//	  go test ./internal/decks/ -run TriageDecks -v
//
// Gated on the dump for the same reason as realdump_manual_test.go:
// it is ~629MB and is not in CI, and a check that skips in CI is the
// same as a check that does not exist. This one is a REPORT rather
// than an assertion — it fails only when a deck cannot be read at
// all, because "42 of 100 cards are unregistered" is the finding, not
// an error.
func TestTriageDecks(t *testing.T) {
	dumpPath := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if dumpPath == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to run against the real dump")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(dumpPath); err != nil {
		t.Fatalf("Load: %v", err)
	}
	t.Logf("loaded %d printings", idx.Count())

	byOracle := make(map[string]catalog.Entry)
	for _, e := range catalog.Build(idx).Cards {
		byOracle[e.OracleID] = e
	}
	t.Logf("catalog holds %d entries", len(byOracle))

	for _, tc := range triageTargets(t) {
		t.Run(tc.name, func(t *testing.T) {
			list, err := tc.load(idx)
			if err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			report(t, byOracle, tc.name, list)
		})
	}
}

// triageTarget is one deck to profile: a display name and whatever it
// takes to turn it into a resolved list.
type triageTarget struct {
	name string
	load func(*cards.Index) (*deck.List, error)
}

// triageTargets is the curated catalog by default, or the decklist
// files under CMDCTRL_TRIAGE_DECKS when that is set.
func triageTargets(t *testing.T) []triageTarget {
	t.Helper()
	root := os.Getenv("CMDCTRL_TRIAGE_DECKS")
	if root == "" {
		out := make([]triageTarget, 0, len(all))
		for _, d := range all {
			id := d.ID
			out = append(out, triageTarget{
				name: id,
				load: func(idx *cards.Index) (*deck.List, error) { return Load(idx, id) },
			})
		}
		return out
	}

	var files []string
	info, err := os.Stat(root)
	if err != nil {
		t.Fatalf("CMDCTRL_TRIAGE_DECKS: %v", err)
	}
	if info.IsDir() {
		entries, err := os.ReadDir(root)
		if err != nil {
			t.Fatalf("CMDCTRL_TRIAGE_DECKS: %v", err)
		}
		for _, e := range entries {
			if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			switch strings.ToLower(filepath.Ext(e.Name())) {
			case ".txt", ".dec", ".dek", ".list":
				files = append(files, filepath.Join(root, e.Name()))
			}
		}
		sort.Strings(files)
	} else {
		files = []string{root}
	}
	if len(files) == 0 {
		t.Fatalf("CMDCTRL_TRIAGE_DECKS=%s holds no decklist files", root)
	}

	out := make([]triageTarget, 0, len(files))
	for _, f := range files {
		path := f
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		out = append(out, triageTarget{
			name: name,
			load: func(idx *cards.Index) (*deck.List, error) { return loadFile(idx, path, name) },
		})
	}
	return out
}

// loadFile parses and resolves one decklist file.
//
// deck.Validate is REPORTED, not enforced. A deck someone actually
// plays can be 99 cards because they are still tuning it, and refusing
// to say anything about its coverage until it is legal would make the
// tool useless exactly when it is most wanted.
func loadFile(idx *cards.Index, path, name string) (*deck.List, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	entries, err := deck.ParseText(string(src))
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	list, err := deck.Resolve(idx, name, entries)
	if err != nil {
		return nil, fmt.Errorf("resolve: %w", err)
	}
	return list, nil
}

// report prints one deck's profile: the counts, then the cards that
// are not `full`, worst first — unregistered before caveated, because
// those are the two different problems and they need different work.
func report(t *testing.T, byOracle map[string]catalog.Entry, name string, list *deck.List) {
	t.Helper()

	type row struct {
		name    string
		grade   string
		caveats []string
	}
	var (
		basics                             int
		full, caveated, unreviewed, unregd int
		rows                               []row
		seen                               = map[string]bool{}
		copies                             = map[string]int{}
		order                              []string
	)

	cardsIn := append(append([]cards.Card{}, list.Commanders...), list.Mainboard...)
	for _, c := range cardsIn {
		if strings.Contains(c.TypeLine, "Basic Land") {
			basics++
			continue
		}
		// Per distinct card, like decks.Coverage: a deck's four
		// Treasure-makers are four rows, its three Islands are one.
		key := c.OracleID.String()
		copies[key]++
		if seen[key] {
			continue
		}
		seen[key] = true
		order = append(order, key)

		e, ok := byOracle[key]
		switch {
		case !ok:
			unregd++
			rows = append(rows, row{c.Name, "UNREGISTERED", nil})
		case e.Completeness == "full":
			full++
		case e.Completeness == "caveats":
			caveated++
			rows = append(rows, row{c.Name, "caveats", e.Caveats})
		default:
			unreviewed++
			rows = append(rows, row{c.Name, "unreviewed", nil})
		}
	}

	graded := full + caveated + unreviewed + unregd
	denom := graded
	if denom == 0 {
		denom = 1
	}
	played := float64(full) / float64(denom) * 100

	t.Logf("")
	t.Logf("=== %s ===", name)
	t.Logf("  %d physical cards — %d basic-land copies, %d distinct non-basic cards",
		len(cardsIn), basics, graded)
	t.Logf("  full          %4d   (%.0f%% of non-basics — played exactly as printed)", full, played)
	t.Logf("  caveats       %4d   (registered, skips at least one printed clause)", caveated)
	t.Logf("  unreviewed    %4d   (registered, nobody has graded it)", unreviewed)
	t.Logf("  UNREGISTERED  %4d   (no implementation — moved by hand)", unregd)
	if err := deck.Validate(list); err != nil {
		t.Logf("  note: not a legal Commander deck as written: %v", err)
	}

	sort.Slice(rows, func(a, b int) bool {
		rank := map[string]int{"UNREGISTERED": 0, "unreviewed": 1, "caveats": 2}
		if rank[rows[a].grade] != rank[rows[b].grade] {
			return rank[rows[a].grade] < rank[rows[b].grade]
		}
		return rows[a].name < rows[b].name
	})
	for _, r := range rows {
		switch {
		case len(r.caveats) == 0:
			t.Logf("  [%-12s] %s", r.grade, r.name)
		default:
			t.Logf("  [%-12s] %s — %s", r.grade, r.name, strings.Join(r.caveats, " / "))
		}
	}
}
