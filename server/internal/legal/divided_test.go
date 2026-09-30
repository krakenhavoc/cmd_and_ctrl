package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// divided_test.go — #1563, CR 601.2d: the enumerator half of "divided
// as you choose". Every divided move a bot is offered announces a
// division the gate accepts (#544): the even split, remainder to the
// earliest targets, and never more targets than the amount can give
// at least 1 each.

const (
	oracleShatterskullSmashing = "78301998-fd9b-4cd5-afad-dbcb43cac2a7"
	oracleFury                 = "fbf9f8c5-849f-45d5-8129-5fc683c21a04"
)

type dividedMove struct {
	Targets      []map[string]any `json:"targets"`
	XValue       int              `json:"x_value"`
	Distribution map[string]int   `json:"distribution"`
}

func decodeDivided(t *testing.T, m legal.Move) dividedMove {
	t.Helper()
	var d dividedMove
	if err := json.Unmarshal(m.Params, &d); err != nil {
		t.Fatalf("decode %q: %v", m.Label, err)
	}
	return d
}

// checkEvenDivision asserts a divided move's shares: one per target,
// each at least 1, summing to `total`, and differing by at most one
// with the larger shares first.
func checkEvenDivision(t *testing.T, label string, d dividedMove, total int) {
	t.Helper()
	if len(d.Targets) == 0 {
		if len(d.Distribution) != 0 {
			t.Errorf("%q: a division with no targets: %v", label, d.Distribution)
		}
		return
	}
	if len(d.Distribution) != len(d.Targets) {
		t.Errorf("%q: %d targets, %d shares", label, len(d.Targets), len(d.Distribution))
		return
	}
	sum, prev := 0, total+1
	for _, tw := range d.Targets {
		v := d.Distribution[tw["id"].(string)]
		if v < 1 {
			t.Errorf("%q: a target was assigned %d", label, v)
		}
		if v > prev {
			t.Errorf("%q: the remainder goes to the earliest targets: %v", label, d.Distribution)
		}
		prev = v
		sum += v
	}
	if sum != total {
		t.Errorf("%q: shares sum to %d, want %d", label, sum, total)
	}
}

func TestShatterskullSmashingMovesAnnounceALegalDivision(t *testing.T) {
	for _, tc := range []struct {
		mountains int
		wantX     int
		wantPairs bool
	}{
		{5, 3, true},  // X=3: two targets take 2/1
		{3, 1, false}, // X=1: a second target would get 0
	} {
		g := newTable(t)
		active := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		clearHand(active)
		spell := handCard(active, game.Card{
			Name: "Shatterskull Smashing", TypeLine: "Sorcery",
			OracleID: oracleShatterskullSmashing, ManaCost: "{X}{R}{R}",
		})
		lands(g, active, "Mountain", "Mountain", tc.mountains)
		battlefieldCard(g, opp, creature("Ox", "{3}{G}", 4, 4))
		battlefieldCard(g, opp, creature("Elk", "{3}{G}", 4, 4))
		advanceTo(t, g, game.StepPrecombatMain)

		moves := castMovesFor(legal.EnumerateFor(g, active.ID), spell)
		if len(moves) == 0 {
			t.Fatalf("%d Mountains: Shatterskull Smashing not offered", tc.mountains)
		}
		pairs := false
		for _, m := range moves {
			d := decodeDivided(t, m)
			if d.XValue != tc.wantX {
				t.Errorf("%d Mountains: %q announces X=%d, want %d", tc.mountains, m.Label, d.XValue, tc.wantX)
			}
			checkEvenDivision(t, m.Label, d, d.XValue)
			if len(d.Targets) == 2 {
				pairs = true
			}
		}
		if pairs != tc.wantPairs {
			t.Errorf("%d Mountains: two-target casts offered = %v, want %v", tc.mountains, pairs, tc.wantPairs)
		}
		// #544: every offer is one the engine accepts.
		dispatchAll(t, g, active.ID, moves)
	}
}

func TestFuryPickTargetAnswersAnnounceALegalDivision(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	fury := handCard(active, game.Card{
		Name: "Fury", TypeLine: "Creature — Elemental Incarnation",
		OracleID: oracleFury, ManaCost: "{3}{R}{R}", Power: 3, Toughness: 3,
	})
	lands(g, active, "Mountain", "Mountain", 5)
	for _, name := range []string{"A", "B", "C", "D", "E"} {
		battlefieldCard(g, opp, creature(name, "{1}{G}", 2, 2))
	}
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(active.ID, fury, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast Fury: %v", err)
	}
	for i := 0; i < 8 && len(g.PendingChoices) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if len(g.PendingChoices) == 0 || g.PendingChoices[0].Kind != game.PendingChoicePickTarget {
		t.Fatal("expected Fury's pick_target prompt")
	}

	moves := legal.EnumerateFor(g, active.ID)
	if len(moves) == 0 {
		t.Fatal("no answers to Fury's prompt — the seat is stuck")
	}
	for _, m := range moves {
		var p struct {
			Targets      []map[string]any `json:"targets"`
			Target       map[string]any   `json:"target"`
			Distribution map[string]int   `json:"distribution"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if len(p.Targets) > 4 {
			t.Errorf("%q: %d targets cannot divide 4", m.Label, len(p.Targets))
		}
		checkEvenDivision(t, m.Label, dividedMove{Targets: p.Targets, Distribution: p.Distribution}, 4)
	}
	dispatchAll(t, g, active.ID, moves)

	// And the trigger resolves with the announced shares: dispatch a
	// two-target answer for real.
	var two *legal.Move
	for i := range moves {
		var p struct {
			Targets []map[string]any `json:"targets"`
		}
		_ = json.Unmarshal(moves[i].Params, &p)
		if len(p.Targets) == 2 {
			two = &moves[i]
			break
		}
	}
	if two == nil {
		t.Fatal("no two-target answer offered")
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(two.Type), Player: active.ID, Caller: active.ID, Params: two.Params}); err != nil {
		t.Fatal(err)
	}
	// The trigger drains onto the stack as soon as its walk finishes.
	var item *game.StackItem
	for _, it := range g.StackMeta {
		if it.Kind == game.StackItemTriggered {
			item = it
		}
	}
	if item == nil {
		t.Fatal("Fury's trigger was not put on the stack")
	}
	total := 0
	for id, v := range item.Distribution {
		if v != 2 {
			t.Errorf("even split of 4 over two: %s got %d", id, v)
		}
		total += v
	}
	if total != 4 || len(item.Distribution) != 2 {
		t.Errorf("the item carries the division: %v", item.Distribution)
	}
}
