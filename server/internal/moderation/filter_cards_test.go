package moderation_test

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/moderation"
)

// A denylist word hiding inside a real card name would refuse a player
// who types the name, or mask it in chat. These two tests are what lets
// the list grow without anyone reading thirty thousand names.

// checkNames fails once per name the filter flags, in both of the ways it
// is used: refusing a name, and masking a chat line. A split card is
// checked face by face as well as whole, because a player types either.
func checkNames(t *testing.T, names []string) {
	t.Helper()
	seen := map[string]bool{}
	var bad []string
	check := func(n string) {
		if n == "" || seen[n] {
			return
		}
		seen[n] = true
		if err := moderation.Name(n); err != nil {
			bad = append(bad, n)
			return
		}
		// A sentence around the name must not make it match either.
		if masked, changed := moderation.Chat("I cast " + n + " on your board"); changed {
			bad = append(bad, n+" (in chat: "+masked+")")
		}
	}
	for _, n := range names {
		check(n)
		for _, face := range strings.Split(n, " // ") {
			check(face)
		}
	}
	sort.Strings(bad)
	if len(bad) > 0 {
		shown := bad
		if len(shown) > 40 {
			shown = shown[:40]
		}
		t.Errorf("%d card names are flagged by the denylist; edit internal/moderation/denylist.txt (first %d):\n  %s",
			len(bad), len(shown), strings.Join(shown, "\n  "))
	}
	t.Logf("checked %d distinct names", len(seen))
}

// TestNoCatalogCardNameIsFlagged runs on every build: the catalog is in
// the binary, so it needs no dump.
func TestNoCatalogCardNameIsFlagged(t *testing.T) {
	var names []string
	for _, s := range effects.All() {
		names = append(names, s.Name)
	}
	if len(names) < 100 {
		t.Fatalf("only %d catalog names; did the registry load?", len(names))
	}
	checkNames(t, names)
}

// TestNoScryfallCardNameIsFlagged is the same check over every printing in
// the Scryfall dump, the whole card pool a deck can hold. It skips without
// CMDCTRL_SCRYFALL_DUMP, like the other dump-gated tests.
func TestNoScryfallCardNameIsFlagged(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to the Scryfall default-cards.json to check every card name")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	if _, err := dec.Token(); err != nil { // the opening [
		t.Fatal(err)
	}
	var names []string
	for dec.More() {
		var c struct {
			Name string `json:"name"`
		}
		if err := dec.Decode(&c); err != nil {
			t.Fatal(err)
		}
		names = append(names, c.Name)
	}
	if len(names) < 10000 {
		t.Fatalf("only %d names in the dump", len(names))
	}
	checkNames(t, names)
}
