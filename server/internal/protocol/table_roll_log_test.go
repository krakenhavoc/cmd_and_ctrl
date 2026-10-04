package protocol

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// table_roll_log_test.go — ADR 0121 §3 and §5: a table roll is a
// `table_roll` line with its sides, result or face and roll_id, one
// line per roll, and the same line keeps its roll_id when an undo
// re-emits it under a new seq.

func tableRollLines(entries []LogEvent) []LogEvent {
	var out []LogEvent
	for _, e := range entries {
		if e.Kind == LogTableRoll {
			out = append(out, e)
		}
	}
	return out
}

func TestTableRollLogLines(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventTableRoll, Actor: me.ID, Sides: 20, Amount: 14, RollID: 1})
		g.EmitEvent(game.Event{Kind: game.EventTableRoll, Actor: me.ID, Sides: 6, Amount: 3, RollID: 2})
		g.EmitEvent(game.Event{Kind: game.EventTableRoll, Actor: them.ID, Label: "heads", RollID: 3})
	})
	view := FilterViewFor(ViewOfGame(g), them.ID.String())
	lines := tableRollLines(view.Log)
	if len(lines) != 3 {
		t.Fatalf("%d table_roll lines, want one per roll: %+v", len(lines), lines)
	}
	want := []struct {
		text    string
		sides   int
		results []int
		faces   []string
		rollID  uint64
		seat    int
	}{
		{me.Name + " rolled a d20 at the table: 14", 20, []int{14}, nil, 1, 0},
		{me.Name + " rolled a d6 at the table: 3", 6, []int{3}, nil, 2, 0},
		{them.Name + " flipped a coin at the table: heads", 0, nil, []string{"heads"}, 3, 1},
	}
	for i, w := range want {
		l := lines[i]
		if l.Text != w.text || l.Sides != w.sides || !reflect.DeepEqual(l.Results, w.results) ||
			!reflect.DeepEqual(l.Faces, w.faces) || l.RollID != w.rollID || l.Seat != w.seat || l.CardID != "" {
			t.Errorf("line %d = %+v, want %+v", i, l, w)
		}
	}
	raw, err := json.Marshal(lines[0])
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	if wire["kind"] != "table_roll" || wire["roll_id"] != float64(1) || wire["sides"] != float64(20) {
		t.Errorf("wire = %v", wire)
	}
}

// An undo of an earlier action re-emits the roll: its line moves to
// a new seq and keeps its roll_id.
func TestTableRollLineSurvivesAnUndo(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	before := g.Clone()
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	ev, err := g.RollTableDie(them.ID, game.TableDieD20)
	if err != nil {
		t.Fatal(err)
	}
	was := tableRollLines(ViewOfGame(g).Log)
	g.WithWriteLock(func() { g.RestoreFrom(before) })
	now := tableRollLines(ViewOfGame(g).Log)
	if len(was) != 1 || len(now) != 1 {
		t.Fatalf("lines before %d, after %d, want 1 and 1", len(was), len(now))
	}
	if now[0].RollID != ev.RollID || now[0].Text != was[0].Text || now[0].Seq == was[0].Seq {
		t.Errorf("after the undo %+v, before %+v: want the same roll_id and text at a new seq", now[0], was[0])
	}
}
