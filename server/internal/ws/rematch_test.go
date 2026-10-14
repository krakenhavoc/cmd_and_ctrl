package ws

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// Tests for rematch.go (#3285): the room keeps its own copy of the
// rematch offer and stamps it on every capture, and the per-seat filter
// reduces it.

func TestRoomStampsRematchOnEveryCapture(t *testing.T) {
	room := NewRoom(seedTestGame(t), slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	if room.Rematch() != nil {
		t.Fatal("a new room has a rematch stamp")
	}
	in := &protocol.RematchView{GameID: "new", Host: "alice", Open: true, Accepted: []string{"bob"}}
	room.SetRematch(in)
	in.Accepted[0] = "changed" // the room keeps its own copy
	in.Open = false

	view, _, err := room.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if view.Rematch == nil || !view.Rematch.Open || view.Rematch.Accepted[0] != "bob" {
		t.Errorf("snapshot stamp = %+v", view.Rematch)
	}
	if got := room.ViewFor("alice").Rematch; got == nil || got.You != protocol.RematchHost {
		t.Errorf("alice's view = %+v, want host", got)
	}
	if got := room.ViewFor("bob").Rematch; got == nil || got.You != protocol.RematchAccepted || got.Accepted != nil {
		t.Errorf("bob's view = %+v, want accepted, with no other seat's answer", got)
	}
	if got := room.ViewFor("carol").Rematch; got == nil || got.You != protocol.RematchUndecided {
		t.Errorf("carol's view = %+v, want undecided", got)
	}
	if got := room.ViewFor(protocol.SpectatorViewerID).Rematch; got != nil {
		t.Errorf("a spectator sees the offer: %+v", got)
	}

	room.SetRematch(nil)
	if view, _, _ := room.Snapshot(); view.Rematch != nil {
		t.Errorf("cleared, but the snapshot still has %+v", view.Rematch)
	}
}

// ConnectedSeats is who is still on the table's page: seats with a socket
// on that game, not spectators, not other games, one entry per seat.
func TestHubConnectedSeats(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := NewRoomManager(log, "")
	g1, g2 := seedTestGame(t), seedTestGame(t)
	mgr.Create(g1)
	mgr.Create(g2)
	alice, bob, carol := g1.Seats[0].ID, g1.Seats[1].ID, g2.Seats[0].ID

	hub := NewHub(log)
	hub.SetManager(mgr)
	hub.SetAuthorizer(UpgradeAuthorizerFunc(func(r *http.Request) (Binding, error) {
		q := r.URL.Query()
		b := Binding{ReadOnly: q.Get("ro") == "1"}
		if raw := q.Get("game"); raw != "" {
			b.GameID = uuid.MustParse(raw)
		}
		if raw := q.Get("player"); raw != "" {
			b.PlayerID = uuid.MustParse(raw)
		}
		return b, nil
	}))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	base := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	for _, q := range []string{
		"?game=" + g1.ID.String() + "&player=" + alice.String(),
		"?game=" + g1.ID.String() + "&player=" + alice.String(), // a second tab
		"?game=" + g1.ID.String() + "&player=" + bob.String(),
		"?game=" + g1.ID.String() + "&ro=1",                     // a spectator
		"?game=" + g2.ID.String() + "&player=" + carol.String(), // another table
	} {
		c := dial(t, base+q)
		t.Cleanup(func() { _ = c.Close() })
	}
	waitForHubCount(t, hub, 5)

	got := hub.ConnectedSeats(g1.ID)
	if len(got) != 2 || !got[alice] || !got[bob] || got[carol] {
		t.Errorf("ConnectedSeats = %v, want alice and bob", got)
	}
	if got := hub.ConnectedSeats(uuid.Nil); len(got) != 0 {
		t.Errorf("ConnectedSeats(nil) = %v", got)
	}
}
