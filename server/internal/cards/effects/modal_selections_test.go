package effects

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// modal_selections_test.go — #2681 (ADR 0065 §6). A modal spell is
// offered every kind of mode selection under the default budget, on a
// board wide enough that one selection's targets alone would spend it:
// the budget goes round the selections before any gets a second
// target set, a repeatable spec offers its mixed multisets (CR
// 700.2d), and the label names the modes.

// modalTable is a main phase on seat 0's turn with ten creatures on
// the battlefield, five each for seat 0 and seat 1, no artifacts and
// nothing on the stack, and the named spell in seat 0's hand with the
// mana to cast it floating.
func modalTable(t *testing.T, name, cost, oracle string, mana ...string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := discardTable(t)
	g.WithWriteLock(func() {
		for i := 0; i < 10; i++ {
			owner := me
			if i%2 == 1 {
				owner = opp
			}
			c := game.Card{
				InstanceID: uuid.New(), Name: fmt.Sprintf("Bear %d", i), TypeLine: "Creature — Bear",
				Power: 2, Toughness: 2, Owner: owner.ID, Controller: owner.ID,
				KnownBy: map[uuid.UUID]bool{},
			}
			for _, s := range g.Seats {
				c.KnownBy[s.ID] = true
			}
			g.Battlefield.PushTop(c)
		}
		for _, m := range mana {
			me.ManaPool.AddMana(game.ManaToken{Color: m})
		}
	})
	spell := handCardOf(g, me, name, "Instant", cost, oracle)
	return g, me, spell
}

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
	g, me, spell := modalTable(t, "Mystic Confluence", "{3}{U}{U}", mysticConfluenceOracle, "U", "U", "C", "C", "C")
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
	g, me, spell := modalTable(t, "Prismari Command", "{1}{U}{R}", prismariCommandOracle, "U", "R", "C")
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
