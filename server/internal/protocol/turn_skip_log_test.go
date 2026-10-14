package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// turn_skip_log_test.go — ADR 0188 Decision 1 and 9 (#3559, CR 614.10):
// the two log lines a skip debt writes, and PlayerView.turn_skips.

func logLinesOfKind(g *game.Game, kind LogKind) []LogEvent {
	var out []LogEvent
	for _, e := range ViewOfGame(g).Log {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

func TestTurnSkipAddedIsInTheLog(t *testing.T) {
	g := buildActiveGame(t)
	source := uuid.New()
	g.WithWriteLock(func() {
		g.Seats[1].Graveyard.PushTop(game.Card{
			InstanceID: source, Name: "Meditate", TypeLine: "Instant",
			Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
		})
		g.EmitEvent(game.Event{Kind: game.EventTurnSkipAdded, Actor: g.Seats[1].ID, Source: source, Amount: 1})
		g.EmitEvent(game.Event{Kind: game.EventTurnSkipAdded, Actor: g.Seats[1].ID, Source: source, Amount: 2})
	})
	lines := logLinesOfKind(g, LogTurnSkip)
	if len(lines) != 2 {
		t.Fatalf("%d turn_skip lines, want 2", len(lines))
	}
	if lines[0].Seat != 1 || lines[0].CardID != source.String() || lines[0].Amount != 1 {
		t.Errorf("line = %+v", lines[0])
	}
	if want := "P2 will skip their next turn (Meditate)"; lines[0].Text != want {
		t.Errorf("text = %q, want %q", lines[0].Text, want)
	}
	if want := "P2 will skip their next 2 turns (Meditate)"; lines[1].Text != want {
		t.Errorf("text = %q, want %q", lines[1].Text, want)
	}
}

func TestTurnSkippedIsInTheLog(t *testing.T) {
	g := buildActiveGame(t)
	source := uuid.New()
	g.WithWriteLock(func() {
		g.Seats[1].Graveyard.PushTop(game.Card{
			InstanceID: source, Name: "Magosi, the Waterveil", TypeLine: "Land",
			Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
		})
		g.EmitEvent(game.Event{Kind: game.EventTurnSkipped, Actor: g.Seats[1].ID, Source: source})
		g.EmitEvent(game.Event{Kind: game.EventTurnSkipped, Actor: g.Seats[1].ID, Source: source, Amount: 3, Label: "extra"})
	})
	lines := logLinesOfKind(g, LogTurnSkipped)
	if len(lines) != 2 {
		t.Fatalf("%d turn_skipped lines, want 2", len(lines))
	}
	if lines[0].Seat != 1 || lines[0].CardID != source.String() {
		t.Errorf("line = %+v", lines[0])
	}
	if want := "P2 skips their turn (Magosi, the Waterveil)"; lines[0].Text != want {
		t.Errorf("text = %q, want %q", lines[0].Text, want)
	}
	if want := "P2 skips their extra turn (Magosi, the Waterveil)"; lines[1].Text != want {
		t.Errorf("text = %q, want %q", lines[1].Text, want)
	}
}

// A debt made by the real engine call reaches the wire as a log line, and
// as turn_skips on the player's panel until it is spent.
func TestTurnSkipsReachTheViewUntilSpent(t *testing.T) {
	g := buildActiveGame(t)
	g.WithWriteLock(func() {
		if !g.SkipNextTurnsForEffect(g.Seats[1].ID, uuid.Nil, 2, "probe") {
			t.Fatal("no debt written")
		}
	})
	v := ViewOfGame(g)
	if v.Seats[1].TurnSkips != 2 || v.Seats[0].TurnSkips != 0 {
		t.Errorf("turn_skips = %d / %d, want 2 on seat 1 and 0 on seat 0", v.Seats[1].TurnSkips, v.Seats[0].TurnSkips)
	}
	if len(logLinesOfKind(g, LogTurnSkip)) != 1 {
		t.Errorf("no turn_skip line for a recorded debt")
	}
	// Public: another viewer sees the same number.
	for _, viewer := range g.Seats {
		fv := FilterViewFor(ViewOfGame(g), viewer.ID.String())
		if fv.Seats[1].TurnSkips != 2 {
			t.Errorf("viewer %s sees turn_skips %d, want 2", viewer.Name, fv.Seats[1].TurnSkips)
		}
	}

	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	// Two seats: each pass walks past seat 1 and spends one charge.
	if got := ViewOfGame(g).Seats[1].TurnSkips; got != 0 {
		t.Errorf("turn_skips = %d after two skipped turns, want 0", got)
	}
}
