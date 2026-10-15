package ws

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// broadcastClient is a hub client with no socket: broadcastToRoom only
// queues onto send, which is buffered here so nothing blocks.
func broadcastClient(h *Hub, gameID, playerID uuid.UUID, admin bool) *Client {
	c := &Client{
		hub:      h,
		send:     make(chan []byte, 4),
		log:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		gameID:   gameID,
		playerID: playerID,
		admin:    admin,
	}
	h.clients[c] = struct{}{}
	return c
}

func onlyFrame(t *testing.T, c *Client) []byte {
	t.Helper()
	select {
	case raw := <-c.send:
		select {
		case extra := <-c.send:
			t.Fatalf("client got a second frame: %.80s", extra)
		default:
		}
		return raw
	default:
		t.Fatal("client got no frame")
		return nil
	}
}

// TestBroadcastBuildsOneFramePerViewer (#3823): clients that share a
// viewer ID share one frame, and every client's frame is exactly the
// one its own viewer ID filters to, so no seat is ever sent another
// seat's (or the admin's) view. A client of another game gets nothing.
func TestBroadcastBuildsOneFramePerViewer(t *testing.T) {
	g := buildTestGame(t, "share")
	h := NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
	seat0, seat1 := g.Seats[0].ID, g.Seats[1].ID

	p0a := broadcastClient(h, g.ID, seat0, false)
	p0b := broadcastClient(h, g.ID, seat0, false) // the same player's second tab
	p1 := broadcastClient(h, g.ID, seat1, false)
	specA := broadcastClient(h, g.ID, uuid.Nil, false)
	specB := broadcastClient(h, g.ID, uuid.Nil, false)
	admin := broadcastClient(h, g.ID, uuid.Nil, true)
	elsewhere := broadcastClient(h, uuid.New(), seat0, false)

	view := protocol.ViewOfGame(g)
	h.broadcastToRoom(g.ID, 7, 2, view)

	want := func(viewer string) []byte {
		raw, err := marshalSnapshotFrame(7, 2, protocol.FilterViewFor(view, viewer))
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	got := map[*Client][]byte{}
	for c, viewer := range map[*Client]string{
		p0a: seat0.String(), p0b: seat0.String(), p1: seat1.String(),
		specA: protocol.SpectatorViewerID, specB: protocol.SpectatorViewerID, admin: "",
	} {
		got[c] = onlyFrame(t, c)
		if !bytes.Equal(got[c], want(viewer)) {
			t.Errorf("viewer %q: the frame is not its own filtered view", viewer)
		}
	}
	select {
	case raw := <-elsewhere.send:
		t.Fatalf("a client of another game got a frame: %.80s", raw)
	default:
	}

	if &got[p0a][0] != &got[p0b][0] || &got[specA][0] != &got[specB][0] {
		t.Error("clients with one viewer ID were each sent a frame of their own; want one shared frame")
	}
	for _, pair := range [][2]*Client{{p0a, p1}, {p0a, specA}, {p1, specA}, {p0a, admin}} {
		if bytes.Equal(got[pair[0]], got[pair[1]]) {
			t.Error("two different viewers were sent identical frames; the fixture no longer tells them apart")
		}
	}
}

// BenchmarkBroadcastWithSpectators is a watched table's broadcast: each
// seat once, and eight spectators.
func BenchmarkBroadcastWithSpectators(b *testing.B) {
	g := buildTestGame(b, "bench")
	h := NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
	var clients []*Client
	for _, p := range g.Seats {
		clients = append(clients, broadcastClient(h, g.ID, p.ID, false))
	}
	for range 8 {
		clients = append(clients, broadcastClient(h, g.ID, uuid.Nil, false))
	}
	view := protocol.ViewOfGame(g)
	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		h.broadcastToRoom(g.ID, uint64(i), 0, view)
		for _, c := range clients {
			<-c.send
		}
	}
}
