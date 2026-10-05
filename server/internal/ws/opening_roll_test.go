package ws

// opening_roll_test.go — ADR 0121 §3 at the WebSocket edge: the
// opening roll's verbs mint no undo entry (nor does anything sent while
// the roll is open), "Roll for everyone left" belongs to the host, and
// every other action waits for the roll to finish.

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// newOpeningRollServer is newE2EServerWithRoom for a four-seat table
// whose opening roll is open, on a key whose round 1 ties seats 2 and 3
// (7, 4, 8, 8) and whose round 2 seat 3 wins.
func newOpeningRollServer(t *testing.T) (string, *game.Game, *Room, func()) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	g := game.NewGame()
	for i := range 4 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Test Commander %d", i+1), uuid.Nil)}
		for j := range 20 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Test Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Player %d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithOpeningRoll(rand.New(rand.NewPCG(22, 2048))); err != nil {
		t.Fatalf("StartWithOpeningRoll: %v", err)
	}
	room := NewRoom(g, log, "")
	hub := NewHub(log)
	hub.SetRoom(room)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	return "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws", g, room, srv.Close
}

func undoDepth(r *Room) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.undoStack)
}

func TestOpeningRollOverTheWire(t *testing.T) {
	wsURL, g, room, cleanup := newOpeningRollServer(t)
	defer cleanup()
	room.SetHost(g.Seats[0].ID)

	conns := make([]interface {
		Close() error
	}, 0, 4)
	defer func() {
		for _, c := range conns {
			c.Close()
		}
	}()
	host := dialAs(t, wsURL, g.Seats[0].ID)
	conns = append(conns, host)
	readNextFrame(t, host)
	other := dialAs(t, wsURL, g.Seats[1].ID)
	conns = append(conns, other)
	readNextFrame(t, other)
	drain := func() { readNextFrame(t, other) } // the broadcast of the host's action

	// Nothing but the roll happens while it is open.
	ep := expectActionError(t, host, protocol.ActionPayload{Type: "keep_hand", Player: g.Seats[0].ID.String()})
	if ep.Code != protocol.CodeBadRequest || ep.Message != "the opening roll is not finished" {
		t.Fatalf("keep_hand during the roll: %+v", ep)
	}

	snap := sendActionAndWait(t, host, protocol.ActionPayload{Type: "roll_opening", Player: g.Seats[0].ID.String()})
	drain()
	or := snap.Game.OpeningRoll
	if or == nil || len(or.Rounds) != 1 || len(or.Rounds[0].Rolls) != 1 || or.Rounds[0].Rolls[0].Result != 7 {
		t.Fatalf("opening_roll after seat 0 rolled: %+v", or)
	}
	if !hasLogText(snap.Game.Log, "Player 1 rolled a d20: 7") {
		t.Fatalf("no roll line: %v", logTexts(snap.Game.Log))
	}

	// "Roll for everyone left" is the host's.
	ep = expectActionError(t, other, protocol.ActionPayload{Type: "host_roll_remaining"})
	if !strings.Contains(ep.Message, "host") {
		t.Fatalf("a non-host's button: %+v", ep)
	}
	snap = sendActionAndWait(t, host, protocol.ActionPayload{Type: "host_roll_remaining"})
	drain()
	or = snap.Game.OpeningRoll
	if len(or.Rounds) != 2 || or.Chooser != nil {
		t.Fatalf("after the host's button: %+v, want a tie round", or)
	}
	if by := or.Rounds[0].Rolls[1].By; by == nil || *by != 0 {
		t.Fatalf("round 1's second die by = %v, want the host's seat 0", by)
	}
	for _, want := range []string{"Player 1 rolled for Player 2, Player 3 and Player 4", "Player 3 and Player 4 tied with 8 and roll again"} {
		if !hasLogText(snap.Game.Log, want) {
			t.Fatalf("no line %q: %v", want, logTexts(snap.Game.Log))
		}
	}
	snap = sendActionAndWait(t, host, protocol.ActionPayload{Type: "host_roll_remaining"})
	drain()
	if c := snap.Game.OpeningRoll.Chooser; c == nil || *c != 3 {
		t.Fatalf("chooser = %v, want seat 3", c)
	}
	if !hasLogText(snap.Game.Log, "Player 4 won the opening roll with 18") {
		t.Fatalf("no won line: %v", logTexts(snap.Game.Log))
	}

	// A concession while the roll is open is not undoable either.
	quitter := dialAs(t, wsURL, g.Seats[2].ID)
	conns = append(conns, quitter)
	readNextFrame(t, quitter)
	sendActionAndWait(t, quitter, protocol.ActionPayload{Type: "concede", Player: g.Seats[2].ID.String()})
	readNextFrame(t, host)
	drain()

	chooser := dialAs(t, wsURL, g.Seats[3].ID)
	conns = append(conns, chooser)
	readNextFrame(t, chooser)
	snap = sendActionAndWait(t, chooser, protocol.ActionPayload{
		Type: "choose_starting_player", Player: g.Seats[3].ID.String(), Params: json.RawMessage(`{"seat":1}`),
	})
	if snap.Game.OpeningRoll != nil || snap.Game.StartingSeat != 1 || !snap.Game.MulligansOpen {
		t.Fatalf("after the choice: roll %+v, starting seat %d", snap.Game.OpeningRoll, snap.Game.StartingSeat)
	}
	if !hasLogText(snap.Game.Log, "Player 4 chose Player 2 to take the first turn") {
		t.Fatalf("no choice line: %v", logTexts(snap.Game.Log))
	}
	if n := undoDepth(room); n != 0 {
		t.Fatalf("the opening roll left %d undo entries, want none", n)
	}
	ep = expectActionError(t, chooser, protocol.ActionPayload{Type: "undo"})
	if ep.Message != "nothing to undo" {
		t.Fatalf("undo after the roll: %+v, want nothing to undo", ep)
	}

	// The first ordinary action after the deal mints the game's first
	// undo entry.
	// CR 103.5: seat 1 took the first turn, so it decides first; seat 3
	// is refused until the seats ahead of it have answered.
	ep = expectActionError(t, chooser, protocol.ActionPayload{Type: "keep_hand", Player: g.Seats[3].ID.String()})
	if !strings.Contains(ep.Message, "your turn to decide") {
		t.Fatalf("keep_hand out of turn: %+v", ep)
	}
	readNextFrame(t, other) // the broadcast of the choice
	sendActionAndWait(t, other, protocol.ActionPayload{Type: "keep_hand", Player: g.Seats[1].ID.String()})
	if n := undoDepth(room); n != 1 {
		t.Fatalf("keep_hand after the deal: %d undo entries, want 1", n)
	}
}

func hasLogText(log []protocol.LogEvent, text string) bool {
	for _, e := range log {
		if e.Text == text {
			return true
		}
	}
	return false
}

func logTexts(log []protocol.LogEvent) []string {
	out := make([]string, len(log))
	for i, e := range log {
		out[i] = e.Text
	}
	return out
}
