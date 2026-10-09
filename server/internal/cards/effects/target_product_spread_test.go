package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// target_product_spread_test.go — #2746. A spell of two target clauses
// is offered the product of their picks under one budget. Spent
// prefix-major, the budget could go entirely on the first clause's
// top-ranked candidate: with the caster's own seat ranked first and a
// second clause holding a budget's worth of picks, every Arc Trail move
// aimed its 2 damage at the caster. The budget now goes round the
// first clause's picks before any gets a second pairing.

// arcTrailFirstSlots is the first-clause target of every Arc Trail move
// offered under opts, with every move dispatched against a clone.
func arcTrailFirstSlots(t *testing.T, g *game.Game, me *game.Player, spell uuid.UUID, opts legal.Options) []uuid.UUID {
	t.Helper()
	var out []uuid.UUID
	for _, m := range legal.EnumerateForWithOptions(g, me.ID, opts) {
		if m.Type != legal.TypeCastSpell || m.Source != spell {
			continue
		}
		var p struct {
			Targets []struct {
				ID   string `json:"id"`
				Slot int    `json:"slot"`
			} `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if len(p.Targets) != 2 {
			t.Errorf("move %q: %d targets, want 2", m.Label, len(p.Targets))
			continue
		}
		for _, tg := range p.Targets {
			if tg.Slot == 0 {
				out = append(out, uuid.MustParse(tg.ID))
			}
		}
		if err := actions.Dispatch(g.Clone(), actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: me.ID, Params: m.Params}); err != nil {
			t.Errorf("move %q rejected: %v", m.Label, err)
		}
	}
	return out
}

func TestArcTrailFirstSlotIsSpreadUnderTheBudget(t *testing.T) {
	g, me, spell := modalTable(t, "Arc Trail", "{1}{R}", arcTrailOracle, "R", "C")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == spell {
				me.Hand.Cards[i].TypeLine = "Sorcery"
			}
		}
	})
	opponents := map[uuid.UUID]bool{}
	mine := map[uuid.UUID]bool{me.ID: true}
	g.ReadSnapshot(func() {
		for _, s := range g.Seats {
			if s.ID != me.ID {
				opponents[s.ID] = true
			}
		}
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID {
				mine[c.InstanceID] = true
			} else {
				opponents[c.InstanceID] = true
			}
		}
	})
	// The board this was found on: a scorer that ranks the caster's
	// own seat first, then the caster's creatures, then the rest.
	opts := legal.Options{OrderTargets: func(c legal.TargetCandidate) float64 {
		switch {
		case c.ID == me.ID:
			return 3
		case mine[c.ID]:
			return 2
		}
		return 1
	}}
	firsts := arcTrailFirstSlots(t, g, me, spell, opts)
	if len(firsts) == 0 {
		t.Fatal("Arc Trail was not offered")
	}
	if firsts[0] != me.ID {
		t.Errorf("the top-ranked candidate still leads: first move's slot 0 is %s, want %s", firsts[0], me.ID)
	}
	seen := map[uuid.UUID]bool{}
	hostile := 0
	for _, id := range firsts {
		if seen[id] {
			t.Errorf("%s takes a second pairing before every first pick has one", id)
		}
		seen[id] = true
		if opponents[id] {
			hostile++
		}
	}
	if hostile == 0 {
		t.Errorf("no Arc Trail move aims its 2 damage at an opponent or their creature: first slots %v", firsts)
	}
	// Twelve moves (the default budget), each with its own first pick:
	// the caster, its five creatures, and six of the opponents' nine.
	if len(firsts) != 12 || hostile != 6 {
		t.Errorf("offered %d moves, %d aimed at an opponent; want 12 and 6", len(firsts), hostile)
	}
}

// With no ranking hook (every human seat) and a budget wide enough for
// the whole product, the spread changes nothing: every ordered pair of
// distinct targets is offered.
func TestArcTrailFullProductUnchangedWithoutTheCap(t *testing.T) {
	g, me, spell := modalTable(t, "Arc Trail", "{1}{R}", arcTrailOracle, "R", "C")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == spell {
				me.Hand.Cards[i].TypeLine = "Sorcery"
			}
		}
	})
	// Four seats and ten creatures: 14 candidates, 14*13 ordered pairs.
	firsts := arcTrailFirstSlots(t, g, me, spell, legal.Options{MaxExpansionPerSource: 1000})
	if len(firsts) != 14*13 {
		t.Errorf("offered %d pairs, want all %d", len(firsts), 14*13)
	}
}
