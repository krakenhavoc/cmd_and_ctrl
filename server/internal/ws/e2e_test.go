package ws

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// updateGolden, when set, causes the e2e test to overwrite its golden
// file with the current output. Run with `go test -update ./...` to
// regenerate after an intentional schema change.
var updateGolden = flag.Bool("update", false, "update golden files instead of diffing")

// normalizeView walks a protocol.GameView in a fixed, field-by-field
// traversal order, collects every UUID it encounters, assigns
// deterministic <uuid-NNN> placeholders to each unique UUID, and
// returns a copy of the view with placeholders substituted.
//
// An earlier version of this function normalized the marshalled JSON
// text with a regex. That broke subtly for views containing
// UUID-keyed maps (PlayerView.CommanderDamage, CardView.Counters):
// encoding/json sorts map keys alphabetically, so the first-
// occurrence order of UUIDs in the emitted bytes depended on the
// alphabetical sort of the raw UUID strings — which are random per
// test run. Walking the struct directly gives us explicit control
// over traversal order, so placeholder assignment is stable.
//
// Traversal order: top-level ID, then each seat in seat order (player
// ID → private zones → CommanderDamage lookups), then shared zones
// (battlefield, stack, exile). UUIDs that first appear as map KEYS
// (CommanderDamage keys are opponent player IDs) have ALREADY been
// assigned placeholders during the preceding seat walk, so the
// non-deterministic map iteration order in Go never allocates a new
// placeholder — it only looks up an existing one.
func normalizeView(v protocol.GameView) protocol.GameView {
	n := &normalizer{ids: make(map[string]string)}
	return n.game(v)
}

type normalizer struct {
	ids  map[string]string
	next int
}

// id returns the stable placeholder for s. Empty string passes
// through unchanged (omitempty on the wire).
func (n *normalizer) id(s string) string {
	if s == "" {
		return ""
	}
	if v, ok := n.ids[s]; ok {
		return v
	}
	n.next++
	v := fmt.Sprintf("<uuid-%03d>", n.next)
	n.ids[s] = v
	return v
}

func (n *normalizer) game(v protocol.GameView) protocol.GameView {
	out := protocol.GameView{
		ID:            n.id(v.ID),
		State:         v.State,
		Turn:          v.Turn,
		MulligansOpen: v.MulligansOpen,
	}
	out.Seats = make([]protocol.PlayerView, len(v.Seats))
	// First pass: assign placeholders to every player's ID so that
	// CommanderDamage lookups on the second pass resolve to existing
	// placeholders rather than allocating new (nondeterministic) ones.
	for i, p := range v.Seats {
		_ = n.id(p.ID)
		out.Seats[i].ID = n.ids[p.ID]
	}
	for i, p := range v.Seats {
		out.Seats[i] = n.player(p, out.Seats[i].ID)
	}
	out.Battlefield = n.zone(v.Battlefield)
	out.Stack = n.zone(v.Stack)
	out.Exile = n.zone(v.Exile)
	return out
}

func (n *normalizer) player(p protocol.PlayerView, id string) protocol.PlayerView {
	out := protocol.PlayerView{
		ID:             id,
		Name:           p.Name,
		Seat:           p.Seat,
		Life:           p.Life,
		Poison:         p.Poison,
		Energy:         p.Energy,
		Library:        n.zone(p.Library),
		Hand:           n.zone(p.Hand),
		Graveyard:      n.zone(p.Graveyard),
		Command:        n.zone(p.Command),
		Eliminated:     p.Eliminated,
		HandKept:       p.HandKept,
		MulligansTaken: p.MulligansTaken,
	}
	// Always initialise (possibly empty) — matches the wire shape
	// emitted by protocol.viewOfPlayer, which always allocates the
	// map. Leaving it nil here would make the golden file show
	// `null` while the real wire shows `{}`.
	out.CommanderDamage = make(map[string]int, len(p.CommanderDamage))
	for k, v := range p.CommanderDamage {
		out.CommanderDamage[n.id(k)] = v
	}
	// Same shape concern as CommanderDamage: viewOfPlayer always
	// allocates the slice. For determinism the timestamp is replaced
	// with a stable placeholder — wall-clock time would diff every
	// run.
	out.LifeHistory = make([]protocol.LifeChangeView, len(p.LifeHistory))
	for i, c := range p.LifeHistory {
		out.LifeHistory[i] = protocol.LifeChangeView{
			Delta:    c.Delta,
			NewTotal: c.NewTotal,
			At:       "<timestamp>",
		}
	}
	return out
}

func (n *normalizer) zone(z protocol.ZoneView) protocol.ZoneView {
	out := protocol.ZoneView{
		Kind:  z.Kind,
		Owner: n.id(z.Owner),
		Count: z.Count,
	}
	// Always allocate, even when empty, to match protocol.viewOfZone
	// which always does `make([]CardView, len(z.Cards))`. An empty
	// but non-nil slice marshals to `[]`; a nil slice marshals to
	// `null`. The wire always emits `[]`.
	out.Cards = make([]protocol.CardView, len(z.Cards))
	for i, c := range z.Cards {
		out.Cards[i] = n.card(c)
	}
	return out
}

func (n *normalizer) card(c protocol.CardView) protocol.CardView {
	return protocol.CardView{
		InstanceID:  n.id(c.InstanceID),
		Name:        c.Name,
		Owner:       n.id(c.Owner),
		Controller:  n.id(c.Controller),
		Tapped:      c.Tapped,
		Counters:    c.Counters,
		IsCommander: c.IsCommander,
	}
}

// newE2EServer stands up a fully-wired httptest server with a room
// whose game has been seeded from a deterministic RNG. Returns the
// ws URL (without query params), the game (for player ID lookups),
// and a cleanup func. Callers append `?game=<id>&player=<id>` as
// needed to bind a specific seat.
func newE2EServer(t *testing.T) (string, *game.Game, func()) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	g := game.NewGame()
	// Fixed-shape synthetic decks for reproducibility.
	for i := range 2 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Test Commander %d", i+1), uuid.Nil)}
		for j := range 20 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Test Filler %d", j+1), uuid.Nil))
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
	hub.SetRoom(room)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	return wsURL, g, srv.Close
}

// dialAs opens a WS connection bound to the given player ID. Passing
// uuid.Nil yields a spectator connection (no player query param),
// which has every opponent hand filtered away in the received
// snapshots.
func dialAs(t *testing.T, wsURL string, playerID uuid.UUID) *websocket.Conn {
	t.Helper()
	u := wsURL
	if playerID != uuid.Nil {
		u += "?player=" + playerID.String()
	}
	return dial(t, u)
}

// sendActionAndWait sends a scripted action and returns the resulting
// snapshot payload. Fails the test on any protocol error.
func sendActionAndWait(t *testing.T, conn *websocket.Conn, a protocol.ActionPayload) protocol.SnapshotPayload {
	t.Helper()
	payload, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal action: %v", err)
	}
	frame, err := json.Marshal(protocol.Frame{
		V:       protocol.Version,
		Kind:    protocol.KindAction,
		ID:      uuid.New().String(),
		Payload: payload,
	})
	if err != nil {
		t.Fatalf("marshal frame: %v", err)
	}
	if err := conn.WriteMessage(websocket.TextMessage, frame); err != nil {
		t.Fatalf("write: %v", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	var f protocol.Frame
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("unmarshal frame: %v", err)
	}
	if f.Kind != protocol.KindSnapshot {
		// Probably an error frame — surface it.
		var ep protocol.ErrorPayload
		_ = json.Unmarshal(f.Payload, &ep)
		t.Fatalf("expected snapshot after %q, got %q (code=%s message=%s)", a.Type, f.Kind, ep.Code, ep.Message)
	}
	var snap protocol.SnapshotPayload
	if err := json.Unmarshal(f.Payload, &snap); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	return snap
}

// readChatFrame reads the next frame from conn and asserts it is a
// chat frame, returning the decoded payload. Used by the S07 chat
// broadcast test.
func readChatFrame(t *testing.T, conn *websocket.Conn) protocol.ChatPayload {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read chat frame: %v", err)
	}
	var f protocol.Frame
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("unmarshal frame: %v", err)
	}
	if f.Kind != protocol.KindChat {
		t.Fatalf("kind: got %q, want %q", f.Kind, protocol.KindChat)
	}
	var p protocol.ChatPayload
	if err := json.Unmarshal(f.Payload, &p); err != nil {
		t.Fatalf("unmarshal chat payload: %v", err)
	}
	return p
}

// TestChatBroadcast covers the S07 chat path: a client-sent text
// frame is server-stamped (author ID, author name, RFC3339 timestamp)
// and broadcast to every client bound to the same game, including the
// originator. Snapshot state is not touched.
func TestChatBroadcast(t *testing.T) {
	wsURL, g, cleanup := newE2EServer(t)
	defer cleanup()

	seat0 := g.Seats[0]
	seat1 := g.Seats[1]

	connA := dialAs(t, wsURL, seat0.ID)
	defer connA.Close()
	connB := dialAs(t, wsURL, seat1.ID)
	defer connB.Close()

	// Both clients consume their initial snapshots so subsequent reads
	// see only the chat broadcast.
	_ = readSnapshotFrame(t, connA)
	_ = readSnapshotFrame(t, connB)

	// Client A sends a chat frame. The author_id/name/timestamp values
	// it puts on the wire are intentionally bogus — the server must
	// overwrite all three.
	out := protocol.ChatPayload{
		AuthorID:   "client-supplied-bogus",
		AuthorName: "client-supplied-bogus",
		Text:       "hello table",
		Timestamp:  "client-supplied-bogus",
	}
	payload, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal chat payload: %v", err)
	}
	frame, err := json.Marshal(protocol.Frame{
		V:       protocol.Version,
		Kind:    protocol.KindChat,
		ID:      uuid.New().String(),
		Payload: payload,
	})
	if err != nil {
		t.Fatalf("marshal chat frame: %v", err)
	}
	if err := connA.WriteMessage(websocket.TextMessage, frame); err != nil {
		t.Fatalf("write: %v", err)
	}

	for label, conn := range map[string]*websocket.Conn{"A": connA, "B": connB} {
		got := readChatFrame(t, conn)
		if got.Text != "hello table" {
			t.Errorf("conn %s: text=%q, want %q", label, got.Text, "hello table")
		}
		if got.AuthorID != seat0.ID.String() {
			t.Errorf("conn %s: author_id=%q, want %q (server should re-stamp from connection's player)",
				label, got.AuthorID, seat0.ID)
		}
		if got.AuthorName != seat0.Name {
			t.Errorf("conn %s: author_name=%q, want %q", label, got.AuthorName, seat0.Name)
		}
		if got.Timestamp == "" || got.Timestamp == "client-supplied-bogus" {
			t.Errorf("conn %s: server did not stamp timestamp (got %q)", label, got.Timestamp)
		}
	}
}

// TestE2EScriptedTurn is the S03 exit-criteria test: drive a scripted
// turn (draw, play, tap, pass_turn) through the real hub via an
// actual WebSocket dial, then diff the final normalized snapshot
// against a golden file.
func TestE2EScriptedTurn(t *testing.T) {
	wsURL, g, cleanup := newE2EServer(t)
	defer cleanup()

	// Drive the scripted turn from seat 0's perspective. Binding the
	// viewer matters now that S04 filters opponent hand + library
	// cards out of every broadcast — a spectator connection would
	// never see the drawn card and the test's `Hand.Cards[0]` lookup
	// would panic on an empty slice.
	seat0ID := g.Seats[0].ID
	conn := dialAs(t, wsURL, seat0ID)
	defer conn.Close()

	// Consume initial snapshot. As of S08, Start deals an opening
	// hand of 7 to each seat — see game.OpeningHandSize.
	initial := readSnapshotFrame(t, conn)
	if initial.Game.Seats[0].Hand.Count != 7 {
		t.Fatalf("initial hand count: got %d, want 7 (opening hand)", initial.Game.Seats[0].Hand.Count)
	}
	if !initial.Game.MulligansOpen {
		t.Fatalf("initial mulligans_open: got false, want true (window opens at Start)")
	}

	seat0 := seat0ID.String()

	// 1) Draw a card. Hand: 7 → 8, library: 13 (20 - 7 dealt) → 12.
	afterDraw := sendActionAndWait(t, conn, protocol.ActionPayload{
		Type:   "draw_card",
		Player: seat0,
	})
	if afterDraw.Game.Seats[0].Hand.Count != 8 {
		t.Errorf("after draw: hand=%d, want 8", afterDraw.Game.Seats[0].Hand.Count)
	}
	if afterDraw.Game.Seats[0].Library.Count != 12 {
		t.Errorf("after draw: library=%d, want 12", afterDraw.Game.Seats[0].Library.Count)
	}

	// 2) Play the first card in hand onto the battlefield. (No longer
	// necessarily the just-drawn card now that the hand starts non-
	// empty — but the play mechanics are identical.)
	drawnCardID := afterDraw.Game.Seats[0].Hand.Cards[0].InstanceID
	params, _ := json.Marshal(map[string]string{"instance_id": drawnCardID})
	afterPlay := sendActionAndWait(t, conn, protocol.ActionPayload{
		Type:   "play_card",
		Player: seat0,
		Params: params,
	})
	if afterPlay.Game.Seats[0].Hand.Count != 7 {
		t.Errorf("after play: hand=%d, want 7", afterPlay.Game.Seats[0].Hand.Count)
	}
	if afterPlay.Game.Battlefield.Count != 1 {
		t.Errorf("after play: battlefield=%d, want 1", afterPlay.Game.Battlefield.Count)
	}

	// 3) Tap the played card ("attack").
	tapParams, _ := json.Marshal(map[string]string{"instance_id": drawnCardID})
	afterTap := sendActionAndWait(t, conn, protocol.ActionPayload{
		Type:   "tap",
		Params: tapParams,
	})
	if !afterTap.Game.Battlefield.Cards[0].Tapped {
		t.Error("after tap: card should be tapped")
	}

	// 4) Pass turn to seat 1.
	afterPass := sendActionAndWait(t, conn, protocol.ActionPayload{
		Type: "pass_turn",
	})
	if afterPass.Game.Turn.ActiveSeat != 1 {
		t.Errorf("after pass_turn: seat=%d, want 1", afterPass.Game.Turn.ActiveSeat)
	}
	if afterPass.Game.Turn.Step != "untap" {
		t.Errorf("after pass_turn: step=%q, want untap", afterPass.Game.Turn.Step)
	}

	// Final state must match the golden file after UUID normalization.
	// normalizeView walks the struct in a deterministic order and
	// replaces every UUID with a stable placeholder; encoding/json
	// then produces identical bytes across runs for the same inputs.
	normalizedPayload := protocol.SnapshotPayload{
		Seq:  afterPass.Seq,
		Game: normalizeView(afterPass.Game),
	}
	normalized, err := json.MarshalIndent(normalizedPayload, "", "  ")
	if err != nil {
		t.Fatalf("marshal normalized snapshot: %v", err)
	}

	goldenPath := filepath.Join("testdata", "e2e_scripted_turn.golden.json")
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(goldenPath, normalized, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("wrote golden file: %s", goldenPath)
		return
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden (run with -update to create): %v", err)
	}
	if !bytes.Equal(normalized, want) {
		t.Errorf("snapshot diverged from golden\n\n--- want (%s) ---\n%s\n--- got ---\n%s",
			goldenPath, want, normalized)
	}
}

// readNextFrame reads either a snapshot or error frame and returns
// the raw protocol.Frame so the caller can assert on Kind. Used by
// the S11 undo authorization tests where the server's response can
// be either an error (rejected) or a snapshot (accepted).
func readNextFrame(t *testing.T, conn *websocket.Conn) protocol.Frame {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read frame: %v", err)
	}
	var f protocol.Frame
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("unmarshal frame: %v", err)
	}
	return f
}

// Game state read helpers — the hub's write goroutine mutates the
// shared *game.Game under its own write lock, so tests that assert on
// fields from the test goroutine must take the read lock too. Without
// these helpers `go test -race` flags a data race on every field
// read. Wrapping each assertion in a ReadSnapshot closure would work
// but is noisy; a pair of tiny accessors keeps the tests readable.
func handSize(g *game.Game, seat int) int {
	var n int
	g.ReadSnapshot(func() { n = g.Seats[seat].Hand.Size() })
	return n
}

func undosRemaining(g *game.Game, seat int) int {
	var n int
	g.ReadSnapshot(func() { n = g.Seats[seat].UndosRemaining })
	return n
}

func undoLimit(g *game.Game) int {
	var n int
	g.ReadSnapshot(func() { n = g.UndoLimit })
	return n
}

// TestUndoRejectsCrossPlayerCaller covers the S11 caller gate: a
// seated player may only undo their own most recent action. Player A
// draws; Player B tries to undo A's draw; B's request must error
// out with bad_request (and not affect game state).
func TestUndoRejectsCrossPlayerCaller(t *testing.T) {
	wsURL, g, cleanup := newE2EServer(t)
	defer cleanup()

	playerA := g.Seats[0].ID
	playerB := g.Seats[1].ID

	connA := dialAs(t, wsURL, playerA)
	defer connA.Close()
	readNextFrame(t, connA) // initial snapshot to A
	connB := dialAs(t, wsURL, playerB)
	defer connB.Close()
	readNextFrame(t, connB) // initial snapshot to B

	// A draws. Both A and B see the broadcast snapshot.
	sendActionAndWait(t, connA, protocol.ActionPayload{
		Type: "draw_card", Player: playerA.String(),
	})
	readNextFrame(t, connB) // B's broadcast copy
	if n := handSize(g, 0); n != 8 {
		t.Fatalf("after A draw: A.hand=%d, want 8", n)
	}

	// B tries to undo A's action. Should fail with bad_request and
	// leave game state untouched.
	sendActionFrame(t, connB, protocol.ActionPayload{Type: "undo"})
	frame := readNextFrame(t, connB)
	if frame.Kind != protocol.KindError {
		t.Fatalf("expected error frame for cross-player undo, got %q", frame.Kind)
	}
	var ep protocol.ErrorPayload
	_ = json.Unmarshal(frame.Payload, &ep)
	if ep.Code != protocol.CodeBadRequest {
		t.Errorf("error code: got %q, want %q", ep.Code, protocol.CodeBadRequest)
	}
	if n := handSize(g, 0); n != 8 {
		t.Errorf("rejected undo mutated state: A.hand=%d, want 8", n)
	}

	// A undoes their own action. Should succeed; broadcast snapshot
	// goes to both clients.
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "undo"})
	readNextFrame(t, connB) // B's broadcast copy
	if n := handSize(g, 0); n != 7 {
		t.Errorf("after A undo: A.hand=%d, want 7", n)
	}
}

// TestUndoBudgetExhausted covers the per-player per-turn budget. A
// draws twice, then undoes once (budget 1→0), then tries to undo
// again — should fail with bad_request even though the stack has
// another A-owned entry to pop.
func TestUndoBudgetExhausted(t *testing.T) {
	wsURL, g, cleanup := newE2EServer(t)
	defer cleanup()

	playerA := g.Seats[0].ID
	connA := dialAs(t, wsURL, playerA)
	defer connA.Close()
	readNextFrame(t, connA) // initial

	if n := undosRemaining(g, 0); n != game.DefaultUndoLimit {
		t.Fatalf("budget at start: got %d, want %d", n, game.DefaultUndoLimit)
	}

	// Two draws.
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "draw_card", Player: playerA.String()})
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "draw_card", Player: playerA.String()})
	if n := handSize(g, 0); n != 9 {
		t.Fatalf("after two draws: A.hand=%d, want 9", n)
	}

	// First undo: succeeds, budget 1→0.
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "undo"})
	if n := undosRemaining(g, 0); n != 0 {
		t.Fatalf("budget after first undo: got %d, want 0", n)
	}
	if n := handSize(g, 0); n != 8 {
		t.Fatalf("after first undo: A.hand=%d, want 8", n)
	}

	// Second undo: should fail with bad_request ("no undos remaining").
	sendActionFrame(t, connA, protocol.ActionPayload{Type: "undo"})
	frame := readNextFrame(t, connA)
	if frame.Kind != protocol.KindError {
		t.Fatalf("expected error frame for exhausted budget, got %q", frame.Kind)
	}
	var ep protocol.ErrorPayload
	_ = json.Unmarshal(frame.Payload, &ep)
	if ep.Code != protocol.CodeBadRequest {
		t.Errorf("error code: got %q, want %q", ep.Code, protocol.CodeBadRequest)
	}
	if n := handSize(g, 0); n != 8 {
		t.Errorf("rejected undo mutated state: A.hand=%d, want 8", n)
	}
}

// TestUndoBudgetRefreshesOnTurnRollover confirms the per-player undo
// budget refreshes when the cursor enters that player's untap step
// (via PassTurn here; the same hook fires on AdvanceStep and on
// PassPriority's wrap branch).
func TestUndoBudgetRefreshesOnTurnRollover(t *testing.T) {
	wsURL, g, cleanup := newE2EServer(t)
	defer cleanup()

	playerA := g.Seats[0].ID
	connA := dialAs(t, wsURL, playerA)
	defer connA.Close()
	readNextFrame(t, connA) // initial

	// Spend A's budget.
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "draw_card", Player: playerA.String()})
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "undo"})
	if n := undosRemaining(g, 0); n != 0 {
		t.Fatalf("budget post-undo: got %d, want 0", n)
	}

	// Pass turn (B becomes active) then back to A. A's budget refreshes
	// on entering A's untap step.
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "pass_turn"})
	if n := undosRemaining(g, 1); n != game.DefaultUndoLimit {
		t.Errorf("B budget after pass_turn: got %d, want %d", n, game.DefaultUndoLimit)
	}
}

// TestSetUndoLimitRefreshesAllSeats covers the in-game admin path:
// any seated player can dial the undo budget up via set_undo_limit
// and every player's UndosRemaining updates to the new limit
// immediately.
func TestSetUndoLimitRefreshesAllSeats(t *testing.T) {
	wsURL, g, cleanup := newE2EServer(t)
	defer cleanup()

	playerA := g.Seats[0].ID
	connA := dialAs(t, wsURL, playerA)
	defer connA.Close()
	readNextFrame(t, connA) // initial

	// Drain A's budget so we can verify the refresh.
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "draw_card", Player: playerA.String()})
	sendActionAndWait(t, connA, protocol.ActionPayload{Type: "undo"})
	if n := undosRemaining(g, 0); n != 0 {
		t.Fatalf("setup: A budget got %d, want 0", n)
	}

	// Bump the limit to 3 — every seat should snap to UndosRemaining=3.
	sendActionAndWait(t, connA, protocol.ActionPayload{
		Type:   "set_undo_limit",
		Params: json.RawMessage(`{"limit":3}`),
	})
	if n := undoLimit(g); n != 3 {
		t.Errorf("game UndoLimit: got %d, want 3", n)
	}
	g.ReadSnapshot(func() {
		for i, p := range g.Seats {
			if p.UndosRemaining != 3 {
				t.Errorf("seat %d UndosRemaining: got %d, want 3", i, p.UndosRemaining)
			}
		}
	})
}
