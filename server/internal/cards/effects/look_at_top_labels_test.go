package effects

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// An agent seat playing Ponder was offered "leave the order alone", "put
// Middling on top" and "put Best on top", and never told which card was
// already on top. Every look-at-top move now names the whole pile, top
// first, so the three cards are all on the page.
func TestPonderMovesNameTheWholePile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Worst", "Middling", "Best")
	castCatalogSpell(t, g, "Ponder", "Sorcery", ponderOracle, nil)
	passPriorityAroundTable(t, g)
	if lookAtTopChoiceFor(g, me.ID) == nil {
		t.Fatal("Ponder queued no look-at prompt")
	}

	var labels []string
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Kind == legal.KindChoice {
			labels = append(labels, m.Label)
		}
	}
	if len(labels) != 3 {
		t.Fatalf("%d look-at moves, want 3: %q", len(labels), labels)
	}
	for _, l := range labels {
		for _, name := range []string{"Worst", "Middling", "Best"} {
			if !strings.Contains(l, name) {
				t.Errorf("move %q does not name %s", l, name)
			}
		}
		if !strings.Contains(l, "top to bottom: ") {
			t.Errorf("move %q does not say which end is the top", l)
		}
	}
}
