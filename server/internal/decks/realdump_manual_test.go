package decks

import (
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// realdump_manual_test.go — the half of the deck checks that needs real
// Scryfall data, gated on CMDCTRL_SCRYFALL_DUMP so CI never touches it.
// Same pattern, same env var and same reason as
// internal/deck/realdump_manual_test.go: the dump is ~629MB and is not
// in the repo.
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json \
//	  go test ./internal/decks/ -run RealDump -v
//
// decks_test.go pins everything that can be checked from the effects
// registry alone: coverage, count, singleton, and colour identity as
// DECLARED. This file pins that the declarations are true — that each
// name resolves, that it resolves to the oracle ID claimed for it, that
// its real colour identity is the one written down, and that
// deck.Validate passes on the resulting list with no violations at all.
//
// Run it after touching a decklist. It is the only thing standing
// between a plausible-looking card name and a bot seat that fails to
// start.
func TestRealDumpDecksAreLegalCommanderDecks(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to run against the real dump")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	t.Logf("loaded %d printings", idx.Count())

	for _, d := range All() {
		t.Run(d.ID, func(t *testing.T) {
			// Load runs ParseText → Resolve → oracle-ID check →
			// Validate, which is the whole pipeline the lobby will run.
			list, err := Load(idx, d.ID)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if got := len(list.Commanders) + len(list.Mainboard); got != 100 {
				t.Errorf("resolved to %d cards, want 100", got)
			}

			// Declared colour identity vs Scryfall's. The offline test
			// checks the subset rule against these declarations; this
			// checks the declarations themselves.
			byName := make(map[string][]string, len(list.Mainboard)+1)
			for _, c := range list.Commanders {
				byName[c.Name] = c.ColorIdentity
			}
			for _, c := range list.Mainboard {
				byName[c.Name] = c.ColorIdentity
			}
			var drift []string
			for _, c := range d.Cards() {
				if c.Basic {
					continue
				}
				actual, ok := byName[c.Name]
				if !ok {
					drift = append(drift, c.Name+" did not resolve")
					continue
				}
				if want, got := c.Identity, identityString(actual); want != got {
					drift = append(drift, c.Name+": declared "+quote(want)+", Scryfall says "+quote(got))
				}
			}
			if len(drift) > 0 {
				sort.Strings(drift)
				t.Errorf("%d colour-identity declaration(s) wrong:\n\t%s", len(drift), strings.Join(drift, "\n\t"))
			}
		})
	}
}

// The coverage profile the picker shows must not depend on whether a
// Scryfall dump is loaded. GET /decks is served either way and says so
// in docs/lobby.md; this is the check behind that sentence.
//
// It is real rather than obvious: CoverageOf joins against
// catalog.Build(idx), and catalog.Build DOES change with an index — it
// fills in printings, parses type lines, and reads printed faces,
// which is what decides an MDFC's "only part of this card is
// automated" caveat. So a dump could in principle move a card between
// buckets, and if it ever does, a player would see one number in
// staging and another in production.
func TestRealDumpCoverageMatchesTheOfflineProfile(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to run against the real dump")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	for _, d := range All() {
		t.Run(d.ID, func(t *testing.T) {
			offline := CoverageOf(nil, d)
			loaded := CoverageOf(idx, d)
			if offline.Cards != loaded.Cards || offline.Full != loaded.Full ||
				offline.Caveats != loaded.Caveats || offline.Unreviewed != loaded.Unreviewed ||
				offline.Unregistered != loaded.Unregistered {
				t.Errorf("coverage differs with a dump loaded:\n\toffline %+v\n\tloaded  %+v", offline, loaded)
			}
			if len(offline.Imperfect) != len(loaded.Imperfect) {
				t.Errorf("%d imperfect cards offline, %d with a dump", len(offline.Imperfect), len(loaded.Imperfect))
			}
			t.Logf("%s: %d cards — %d full, %d caveats, %d unreviewed, %d basic land rows",
				d.ID, loaded.Cards, loaded.Full, loaded.Caveats, loaded.Unreviewed, loaded.Basics)
		})
	}
}

// identityString normalises a Scryfall color_identity array into the
// WUBRG-ordered string this package declares.
func identityString(syms []string) string {
	const order = "WUBRG"
	present := map[rune]bool{}
	for _, s := range syms {
		for _, r := range strings.ToUpper(s) {
			present[r] = true
		}
	}
	var b strings.Builder
	for _, r := range order {
		if present[r] {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func quote(s string) string {
	if s == "" {
		return `"" (colourless)`
	}
	return `"` + s + `"`
}
