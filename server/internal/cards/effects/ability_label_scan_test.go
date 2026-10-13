package effects

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ability_label_scan_test.go — #2410: how many static ◆ rows on the
// art tiles still read as a kind of effect ("Replacement effect")
// instead of as wording a player can use.

// genericStaticRows lists, per card name, the static rows the catalog
// gave no wording.
func genericStaticRows() map[string][]string {
	out := map[string][]string{}
	for _, s := range All() {
		for _, r := range game.AbilityRowsOf(game.Card{Name: s.Name, OracleID: s.OracleID}) {
			if r.Kind == game.AbilityRowStatic && r.Generic {
				out[s.Name] = append(out[s.Name], r.Label)
			}
		}
	}
	return out
}

// TestGenericStaticRowCount prints the remaining count (the figure the
// PR body quotes) and, with ABILITY_LABEL_SCAN=1, the card list.
func TestGenericStaticRowCount(t *testing.T) {
	rows := genericStaticRows()
	total := 0
	kinds := map[string]int{}
	var names []string
	for n, ls := range rows {
		total += len(ls)
		names = append(names, n)
		for _, l := range ls {
			kinds[l]++
		}
	}
	sort.Strings(names)
	t.Logf("generic static rows: %d on %d cards (of %d catalogued)", total, len(names), len(All()))
	if os.Getenv("ABILITY_LABEL_SCAN") != "" {
		var b strings.Builder
		for _, n := range names {
			fmt.Fprintf(&b, "%s\t%s\n", n, strings.Join(rows[n], " | "))
		}
		fmt.Print(b.String())
		for l, c := range kinds {
			fmt.Printf("KIND\t%d\t%s\n", c, l)
		}
	}
}
