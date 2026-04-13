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
		ID:    n.id(v.ID),
		State: v.State,
		Turn:  v.Turn,
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
		ID:        id,
		Name:      p.Name,
		Seat:      p.Seat,
		Life:      p.Life,
		Poison:    p.Poison,
		Energy:    p.Energy,
		Library:   n.zone(p.Library),
		Hand:      n.zone(p.Hand),
		Graveyard: n.zone(p.Graveyard),
		Command:   n.zone(p.Command),
	}
	// Always initialise (possibly empty) — matches the wire shape
	// emitted by protocol.viewOfPlayer, which always allocates the
	// map. Leaving it nil here would make the golden file show
	// `null` while the real wire shows `{}`.
	out.CommanderDamage = make(map[string]int, len(p.CommanderDamage))
	for k, v := range p.CommanderDamage {
		out.CommanderDamage[n.id(k)] = v
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

	// Consume initial snapshot.
	initial := readSnapshotFrame(t, conn)
	if initial.Game.Seats[0].Hand.Count != 0 {
		t.Fatalf("initial hand count: got %d, want 0", initial.Game.Seats[0].Hand.Count)
	}

	seat0 := seat0ID.String()

	// 1) Draw a card.
	afterDraw := sendActionAndWait(t, conn, protocol.ActionPayload{
		Type:   "draw_card",
		Player: seat0,
	})
	if afterDraw.Game.Seats[0].Hand.Count != 1 {
		t.Errorf("after draw: hand=%d, want 1", afterDraw.Game.Seats[0].Hand.Count)
	}
	if afterDraw.Game.Seats[0].Library.Count != 19 {
		t.Errorf("after draw: library=%d, want 19", afterDraw.Game.Seats[0].Library.Count)
	}

	// 2) Play the drawn card onto the battlefield.
	drawnCardID := afterDraw.Game.Seats[0].Hand.Cards[0].InstanceID
	params, _ := json.Marshal(map[string]string{"instance_id": drawnCardID})
	afterPlay := sendActionAndWait(t, conn, protocol.ActionPayload{
		Type:   "play_card",
		Player: seat0,
		Params: params,
	})
	if afterPlay.Game.Seats[0].Hand.Count != 0 {
		t.Errorf("after play: hand=%d, want 0", afterPlay.Game.Seats[0].Hand.Count)
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
