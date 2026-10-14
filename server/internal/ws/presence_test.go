package ws

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0162 §2: the `visibility` frame, Hub.SeatAway (away is no seat
// connection, or every one hidden; a client that never sends the frame
// is visible), and Room.LastActed.

// presenceServer is newE2EServerWithRoom with the hub handed back and
// its clock under the test's control.
func presenceServer(t *testing.T) (string, *game.Game, *Room, *Hub, *testClock) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	g := game.NewGame()
	for i := range 2 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := range 20 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Player %d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(42, 42))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	room := NewRoom(g, log, "")
	hub := NewHub(log)
	clk := &testClock{at: time.Date(2026, 10, 13, 20, 0, 0, 0, time.UTC)}
	hub.clock = clk.now
	hub.SetRoom(room)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws", g, room, hub, clk
}

type testClock struct {
	mu sync.Mutex
	at time.Time
}

func (c *testClock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.at
}

func (c *testClock) set(at time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.at = at
}

func sendVisibility(t *testing.T, conn *websocket.Conn, payload string) {
	t.Helper()
	sendFrame(t, conn, protocol.Frame{V: protocol.Version, Kind: protocol.KindVisibility, ID: uuid.NewString(), Payload: json.RawMessage(payload)})
}

// awaitAway polls SeatAway until it answers want: a visibility frame is
// read on the connection's own goroutine, after the write returns.
func awaitAway(t *testing.T, hub *Hub, gameID, seat uuid.UUID, want bool) time.Time {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		away, since := hub.SeatAway(gameID, seat)
		if away == want {
			return since
		}
		if time.Now().After(deadline) {
			t.Fatalf("SeatAway = %v, want %v", away, want)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestSeatAwayWithNoConnectionHasNoDepartureTime(t *testing.T) {
	_, g, _, hub, _ := presenceServer(t)
	away, since := hub.SeatAway(g.ID, g.Seats[0].ID)
	if !away || !since.IsZero() {
		t.Errorf("a seat nobody connected to: away %v since %v, want away with no departure time", away, since)
	}
}

func TestVisibilityFrameDecidesAway(t *testing.T) {
	wsURL, g, _, hub, clk := presenceServer(t)
	seat := g.Seats[0].ID
	conn := dialAs(t, wsURL, seat)
	defer conn.Close()
	readSnapshotFrame(t, conn)

	// Connected and never said otherwise: visible, so present.
	awaitAway(t, hub, g.ID, seat, false)

	hidAt := clk.now().Add(time.Minute)
	clk.set(hidAt)
	sendVisibility(t, conn, `{"visible": false}`)
	if since := awaitAway(t, hub, g.ID, seat, true); !since.Equal(hidAt) {
		t.Errorf("hidden: away since %v, want %v", since, hidAt)
	}

	sendVisibility(t, conn, `{"visible": true}`)
	awaitAway(t, hub, g.ID, seat, false)

	// The other seat is unaffected throughout.
	if away, _ := hub.SeatAway(g.ID, g.Seats[1].ID); !away {
		t.Error("seat 1 has no connection and reads present")
	}
}

func TestOneVisibleTabKeepsTheSeatPresent(t *testing.T) {
	wsURL, g, _, hub, _ := presenceServer(t)
	seat := g.Seats[0].ID
	a := dialAs(t, wsURL, seat)
	defer a.Close()
	readSnapshotFrame(t, a)
	b := dialAs(t, wsURL, seat)
	defer b.Close()
	readSnapshotFrame(t, b)

	sendVisibility(t, a, `{"visible": false}`)
	// Give a's frame time to land; b has said nothing and is visible.
	time.Sleep(50 * time.Millisecond)
	if away, _ := hub.SeatAway(g.ID, seat); away {
		t.Fatal("one hidden tab and one visible one read as away")
	}
	sendVisibility(t, b, `{"visible": false}`)
	awaitAway(t, hub, g.ID, seat, true)
}

func TestClosingTheLastConnectionStampsTheDeparture(t *testing.T) {
	wsURL, g, _, hub, clk := presenceServer(t)
	seat := g.Seats[0].ID
	conn := dialAs(t, wsURL, seat)
	readSnapshotFrame(t, conn)
	awaitAway(t, hub, g.ID, seat, false)

	leftAt := clk.now().Add(3 * time.Minute)
	clk.set(leftAt)
	_ = conn.Close()
	if since := awaitAway(t, hub, g.ID, seat, true); !since.Equal(leftAt) {
		t.Errorf("after closing: away since %v, want %v", since, leftAt)
	}
}

func TestMalformedVisibilityIsRefusedAndChangesNothing(t *testing.T) {
	wsURL, g, _, hub, _ := presenceServer(t)
	seat := g.Seats[0].ID
	conn := dialAs(t, wsURL, seat)
	defer conn.Close()
	readSnapshotFrame(t, conn)
	for _, payload := range []string{`{}`, `{"visible": "no"}`, `null`} {
		sendVisibility(t, conn, payload)
		f := readRawFrame(t, conn)
		if f.Kind != protocol.KindError {
			t.Fatalf("visibility %s answered %q, want an error", payload, f.Kind)
		}
	}
	if away, _ := hub.SeatAway(g.ID, seat); away {
		t.Error("a refused visibility frame hid the seat")
	}
}

func TestLastActedStampsOnlyTheSeatThatActed(t *testing.T) {
	wsURL, g, room, _, _ := presenceServer(t)
	seat0, seat1 := g.Seats[0].ID, g.Seats[1].ID
	if !room.LastActed(seat0).IsZero() {
		t.Fatal("LastActed before any action is not zero")
	}
	conn := dialAs(t, wsURL, seat0)
	defer conn.Close()
	readSnapshotFrame(t, conn)
	before := time.Now()
	sendActionAndWait(t, conn, protocol.ActionPayload{Type: "keep_hand", Player: seat0.String()})
	if at := room.LastActed(seat0); at.Before(before) {
		t.Errorf("after keep_hand, LastActed = %v, want at or after %v", at, before)
	}
	if !room.LastActed(seat1).IsZero() {
		t.Error("the seat that did not act was stamped")
	}
}
