package testsm

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// modeSelectionsOffered is the sorted mode multisets the cast moves for
// spell carry, keyed like "1,1,2", with every move dispatched against a
// clone and every label checked to be unique.
func modeSelectionsOffered(t *testing.T, g *game.Game, me *game.Player, spell uuid.UUID) map[string]int {
	t.Helper()
	out := map[string]int{}
	labels := map[string]bool{}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != spell {
			continue
		}
		var p struct {
			Modes []int `json:"modes"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		sel := slices.Clone(p.Modes)
		slices.Sort(sel)
		out[fmt.Sprint(sel)]++
		if labels[m.Label] {
			t.Errorf("two moves share the label %q", m.Label)
		}
		labels[m.Label] = true
		if err := actions.Dispatch(g.Clone(), actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: me.ID, Params: m.Params}); err != nil {
			t.Errorf("move %q rejected: %v", m.Label, err)
		}
	}
	return out
}

func TestMysticConfluenceIsOfferedEverySelection(t *testing.T) {
	g, me, spell := effectstest.ModalTable(t, "Mystic Confluence", "{3}{U}{U}", effectstest.MysticConfluenceOracle, "U", "U", "C", "C", "C")
	got := modeSelectionsOffered(t, g, me, spell)
	// Nothing on the stack, so the counter bullet (0) has no target and
	// is never offered. Bounce is 1, draw is 2.
	for _, want := range []string{"[1 1 1]", "[2 2 2]", "[1 1 2]", "[1 2 2]"} {
		if got[want] == 0 {
			t.Errorf("selection %s not offered; got %v", want, got)
		}
	}
	if len(got) != 4 {
		t.Errorf("offered %v, want exactly the four bounce/draw multisets", got)
	}
}

func TestPrismariCommandIsOfferedEveryPair(t *testing.T) {
	g, me, spell := effectstest.ModalTable(t, "Prismari Command", "{1}{U}{R}", effectstest.PrismariCommandOracle, "U", "R", "C")
	got := modeSelectionsOffered(t, g, me, spell)
	// No artifact, so "destroy target artifact" (3) is never offered.
	for _, want := range []string{"[0 1]", "[0 2]", "[1 2]"} {
		if got[want] == 0 {
			t.Errorf("selection %s not offered; got %v", want, got)
		}
	}
	if len(got) != 3 {
		t.Errorf("offered %v, want exactly the three pairs without the artifact bullet", got)
	}
}
