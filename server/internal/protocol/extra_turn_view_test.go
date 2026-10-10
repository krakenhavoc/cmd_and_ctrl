package protocol

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// extra_turn_view_test.go — ADR 0059 Decision 11 for CR 500.7 extra
// turns: the table sees the queue (`extra_turns`, next first), the mark
// on the extra turn itself (`extra`, with `number` still the round), and
// one log line per queued turn.

func TestExtraTurnsAreOnTheTurnViewAndInTheLog(t *testing.T) {
	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	active := g.Turn.ActiveSeat
	other := (active + 1) % len(g.Seats)
	source := uuid.New()
	g.WithWriteLock(func() {
		g.Seats[active].Graveyard.PushTop(game.Card{
			InstanceID: source, Name: "Time Warp", TypeLine: "Sorcery",
			Owner: g.Seats[active].ID, Controller: g.Seats[active].ID,
		})
		g.TakeExtraTurnsForEffect(g.Seats[other].ID, source, 1)
		g.TakeExtraTurnsForEffect(g.Seats[active].ID, source, 1)
	})

	v := ViewOfGame(g)
	if v.Turn.Extra {
		t.Error("the current turn is a normal one")
	}
	if want := []int{active, other}; !reflect.DeepEqual(v.Turn.ExtraTurns, want) {
		t.Errorf("extra_turns = %v, want %v (most recently created first)", v.Turn.ExtraTurns, want)
	}
	var lines []LogEvent
	for _, e := range v.Log {
		if e.Kind == LogExtraTurn {
			lines = append(lines, e)
		}
	}
	if len(lines) != 2 {
		t.Fatalf("want one extra_turn line per turn, got %d", len(lines))
	}
	if lines[0].Seat != other || lines[0].CardID != source.String() {
		t.Errorf("first line seat %d card %q, want seat %d and the source", lines[0].Seat, lines[0].CardID, other)
	}
	if !strings.Contains(lines[0].Text, "will take an extra turn") || !strings.Contains(lines[0].Text, "Time Warp") {
		t.Errorf("text = %q", lines[0].Text)
	}

	round := v.Turn.Number
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	v = ViewOfGame(g)
	if !v.Turn.Extra || v.Turn.ActiveSeat != active || v.Turn.Number != round {
		t.Errorf("turn view %+v: want seat %d's extra turn, still round %d", v.Turn, active, round)
	}
	if want := []int{other}; !reflect.DeepEqual(v.Turn.ExtraTurns, want) {
		t.Errorf("extra_turns = %v, want %v", v.Turn.ExtraTurns, want)
	}

	if err := g.Concede(g.Seats[other].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if v := ViewOfGame(g); len(v.Turn.ExtraTurns) != 0 {
		t.Errorf("a departed player's queued turn is still shown: %v", v.Turn.ExtraTurns)
	}
}
