package ws

import (
	"bufio"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// deflate_bench_test.go — ADR 0205 delivery step 1 (#4200): what
// permessage-deflate would cost a broadcast, measured before it is
// turned on. Nothing here changes the hub: the upgrader still has
// EnableCompression unset.
//
// The benchmark runs the real broadcastToRoom (filter and marshal once
// per viewer ID, #3884), then plays the write pump for every connection
// onto a gorilla *websocket.Conn whose socket only counts bytes:
//
//   - off: today, one WriteMessage per connection, uncompressed.
//   - level1, level6: what step 2 would do, one websocket.PreparedMessage
//     per viewer frame, shared by every connection that was queued the
//     same bytes, on a connection that negotiated permessage-deflate
//     (no context takeover, the only mode gorilla has) at that level.
//     The PreparedMessage deflates once per viewer and copies the
//     result to each connection.
//
// ns/op is on one goroutine (production spreads the writes over the
// write pumps, but the CPU is the same). Each op also reports wire-B/op
// (bytes on the sockets, frame headers included) and frame-B/op (the
// marshalled frames, what cmdctrl_ws_sent_bytes_total counts).

// countingConn is the socket under a benchmark connection: writes are
// counted and dropped, and nothing is ever read (no read pump runs).
type countingConn struct{ n int64 }

func (c *countingConn) Write(p []byte) (int, error)      { c.n += int64(len(p)); return len(p), nil }
func (c *countingConn) Read([]byte) (int, error)         { return 0, io.EOF }
func (c *countingConn) Close() error                     { return nil }
func (c *countingConn) LocalAddr() net.Addr              { return &net.TCPAddr{} }
func (c *countingConn) RemoteAddr() net.Addr             { return &net.TCPAddr{} }
func (c *countingConn) SetDeadline(time.Time) error      { return nil }
func (c *countingConn) SetReadDeadline(time.Time) error  { return nil }
func (c *countingConn) SetWriteDeadline(time.Time) error { return nil }

// hijackRecorder is a ResponseWriter the upgrader can hijack onto a
// countingConn.
type hijackRecorder struct {
	*httptest.ResponseRecorder
	conn *countingConn
}

func (h hijackRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return h.conn, bufio.NewReadWriter(bufio.NewReader(h.conn), bufio.NewWriter(h.conn)), nil
}

// benchWire upgrades a server-side connection onto a countingConn, as
// the hub's upgrader would with EnableCompression set, and a browser
// that offers permessage-deflate. level 0 negotiates no compression.
// The handshake's bytes are not counted.
func benchWire(tb testing.TB, level int) (*websocket.Conn, *countingConn) {
	tb.Helper()
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	req.Header.Set("Sec-WebSocket-Extensions", "permessage-deflate; client_max_window_bits")
	u := baseUpgrader
	u.EnableCompression = level != 0
	u.CheckOrigin = func(*http.Request) bool { return true }
	cc := &countingConn{}
	conn, err := u.Upgrade(hijackRecorder{httptest.NewRecorder(), cc}, req, nil)
	if err != nil {
		tb.Fatalf("upgrade: %v", err)
	}
	if level != 0 {
		if err := conn.SetCompressionLevel(level); err != nil {
			tb.Fatalf("SetCompressionLevel(%d): %v", level, err)
		}
		conn.EnableWriteCompression(true)
	}
	cc.n = 0
	return conn, cc
}

// watchedTable puts BenchmarkBroadcastWithSpectators's audience on g:
// each seat once, and eight spectators.
func watchedTable(h *Hub, g *game.Game) []*Client {
	var clients []*Client
	for _, p := range g.Seats {
		clients = append(clients, broadcastClient(h, g.ID, p.ID, false))
	}
	for range 8 {
		clients = append(clients, broadcastClient(h, g.ID, uuid.Nil, false))
	}
	return clients
}

// deflateBroadcast is one commit's broadcast to clients, written to
// wires (one per client) at level (0 = off). It returns the bytes of
// the frames queued.
func deflateBroadcast(tb testing.TB, h *Hub, g *game.Game, view protocol.GameView, seq uint64,
	clients []*Client, wires []*websocket.Conn, level int) int64 {
	return writeFrames(tb, queuedFrames(h, g, view, seq, clients), wires, level)
}

// queuedFrames runs broadcastToRoom and takes each client's frame off
// its queue, in clients' order.
func queuedFrames(h *Hub, g *game.Game, view protocol.GameView, seq uint64, clients []*Client) [][]byte {
	h.broadcastToRoom(g.ID, seq, 0, view)
	frames := make([][]byte, len(clients))
	for i, c := range clients {
		frames[i] = <-c.send
	}
	return frames
}

// writeFrames is the write pumps' half of a commit: frames[i] written
// to wires[i] at level (0 = off), one PreparedMessage per distinct
// frame when compressed. It returns the bytes of the frames.
func writeFrames(tb testing.TB, frames [][]byte, wires []*websocket.Conn, level int) int64 {
	var queued int64
	prepared := make(map[*byte]*websocket.PreparedMessage, 6)
	for i, raw := range frames {
		queued += int64(len(raw))
		if level == 0 {
			if err := wires[i].WriteMessage(websocket.TextMessage, raw); err != nil {
				tb.Fatal(err)
			}
			continue
		}
		pm, ok := prepared[&raw[0]]
		if !ok {
			var err error
			if pm, err = websocket.NewPreparedMessage(websocket.TextMessage, raw); err != nil {
				tb.Fatal(err)
			}
			prepared[&raw[0]] = pm
		}
		if err := wires[i].WritePreparedMessage(pm); err != nil {
			tb.Fatal(err)
		}
	}
	return queued
}

// deflateBoard is a board the benchmark broadcasts.
type deflateBoard struct {
	name  string
	build func(tb testing.TB) *game.Game
}

// deflateBoards are BenchmarkBroadcastWithSpectators's board, the
// midgame board of #3823 (filler cards, so a floor: ADR 0205), and every
// scrubbed cmd-dev restore point in the snapshot corpus (real catalog
// cards), read and never written.
func deflateBoards(tb testing.TB) []deflateBoard {
	boards := []deflateBoard{
		{"spectators", func(tb testing.TB) *game.Game { return buildTestGame(tb, "bench") }},
		{"midgame", func(tb testing.TB) *game.Game { return midgameRestoreTable(tb, 5000) }},
	}
	files, err := filepath.Glob(filepath.Join("..", "game", "testdata", "snapshots", "real", "*.json"))
	if err != nil {
		tb.Fatal(err)
	}
	for _, f := range files {
		boards = append(boards, deflateBoard{
			name:  "real_" + strings.TrimSuffix(filepath.Base(f), ".json"),
			build: func(tb testing.TB) *game.Game { return restoreCorpusFile(tb, f) },
		})
	}
	return boards
}

// restoreCorpusFile restores one corpus restore point the way
// TestSnapshotCorpusRestores does.
func restoreCorpusFile(tb testing.TB, path string) *game.Game {
	tb.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		tb.Fatal(err)
	}
	var file struct {
		Snapshot *game.GameSnapshot `json:"snapshot"`
	}
	if err := json.Unmarshal(raw, &file); err != nil || file.Snapshot == nil {
		tb.Fatalf("%s: not a restore point: %v", path, err)
	}
	g, err := file.Snapshot.RestoreStrict()
	if err != nil {
		tb.Fatalf("%s: RestoreStrict: %v", path, err)
	}
	return g
}

var deflateLevels = []struct {
	name  string
	level int
}{{"off", 0}, {"level1", 1}, {"level6", 6}}

// BenchmarkBroadcastDeflate is ADR 0205's deflate CPU benchmark: a
// watched table's broadcast (each seat once, eight spectators) written
// uncompressed and at compression levels 1 and 6.
//
//   - commit: the whole broadcast, filter and marshal included, then
//     the writes.
//   - writes: the writes alone, of one commit's frames built once
//     before the timer. A commit's frames are new bytes every time, so
//     each op makes new PreparedMessages and deflates them again. The
//     deflate cost is a level's ns/op less off's, without the filter's
//     and marshal's noise.
func BenchmarkBroadcastDeflate(b *testing.B) {
	for _, board := range deflateBoards(b) {
		b.Run(board.name, func(b *testing.B) {
			g := board.build(b)
			view := protocol.ViewOfGame(g)
			for _, part := range []string{"commit", "writes"} {
				for _, lv := range deflateLevels {
					b.Run(part+"/"+lv.name, func(b *testing.B) {
						h := NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
						clients := watchedTable(h, g)
						wires := make([]*websocket.Conn, len(clients))
						counts := make([]*countingConn, len(clients))
						for i := range clients {
							wires[i], counts[i] = benchWire(b, lv.level)
						}
						frames := queuedFrames(h, g, view, 0, clients)
						var queued int64
						b.ReportAllocs()
						b.ResetTimer()
						for i := range b.N {
							if part == "commit" {
								queued += deflateBroadcast(b, h, g, view, uint64(i), clients, wires, lv.level)
							} else {
								queued += writeFrames(b, frames, wires, lv.level)
							}
						}
						b.StopTimer()
						var wire int64
						for _, c := range counts {
							wire += c.n
						}
						b.ReportMetric(float64(wire)/float64(b.N), "wire-B/op")
						b.ReportMetric(float64(queued)/float64(b.N), "frame-B/op")
					})
				}
			}
		})
	}
}

// TestBroadcastDeflateHarness keeps the benchmark honest: off puts each
// frame on the wire as it was queued (plus a frame header), and a
// compressed level negotiates permessage-deflate and puts fewer bytes
// on the wire than were queued, so the benchmark is measuring deflate
// and not a connection that quietly sent plain frames.
func TestBroadcastDeflateHarness(t *testing.T) {
	g := buildTestGame(t, "harness")
	view := protocol.ViewOfGame(g)
	for _, lv := range deflateLevels {
		h := NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
		clients := watchedTable(h, g)
		wires := make([]*websocket.Conn, len(clients))
		counts := make([]*countingConn, len(clients))
		for i := range clients {
			wires[i], counts[i] = benchWire(t, lv.level)
		}
		queued := deflateBroadcast(t, h, g, view, 1, clients, wires, lv.level)
		var wire int64
		for _, c := range counts {
			wire += c.n
		}
		// A server frame's header is 2 to 10 bytes.
		headers := int64(len(clients))
		switch {
		case lv.level == 0 && (wire < queued+2*headers || wire > queued+10*headers):
			t.Errorf("off: %d bytes on the wire for %d queued; want the frames plus a header each", wire, queued)
		case lv.level != 0 && wire*2 > queued:
			t.Errorf("%s: %d bytes on the wire for %d queued; want under half (is deflate negotiated?)", lv.name, wire, queued)
		}
	}
}
