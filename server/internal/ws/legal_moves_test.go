package ws

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// legal_moves_test.go — ADR 0122 §6.1: the legal_moves_request frame
// and its reply. The reply is the bound seat's and nobody else's,
// carries the list the snapshot's 48-move cap degraded, names the
// state it describes, and is rationed to four a second.

const oracleBolt = "4457ed35-7c10-48c8-9776-456485fdf070" // Lightning Bolt

// serveBound serves room on a fresh hub that binds every connection to
// b, and dials it. clock, when set, is the hub's clock.
func serveBound(t *testing.T, room *Room, b Binding, clock func() time.Time) *websocket.Conn {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewHub(log)
	hub.SetRoom(room)
	hub.clock = clock
	hub.SetAuthorizer(UpgradeAuthorizerFunc(func(*http.Request) (Binding, error) { return b, nil }))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	conn := dial(t, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws")
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// requestLegalMoves sends a legal_moves_request with a raw payload and
// returns the reply frame, which must carry the request's id.
func requestLegalMoves(t *testing.T, conn *websocket.Conn, payload string) protocol.Frame {
	t.Helper()
	id := uuid.NewString()
	f := protocol.Frame{V: protocol.Version, Kind: protocol.KindLegalMovesRequest, ID: id}
	if payload != "" {
		f.Payload = json.RawMessage(payload)
	}
	sendFrame(t, conn, f)
	reply := readFrame(t, conn)
	if reply.ID != id {
		t.Fatalf("reply id %q, want the request's %q (kind %q)", reply.ID, id, reply.Kind)
	}
	return reply
}

func legalMovesReply(t *testing.T, f protocol.Frame) protocol.LegalMovesPayload {
	t.Helper()
	if f.Kind != protocol.KindLegalMoves {
		var ep protocol.ErrorPayload
		_ = json.Unmarshal(f.Payload, &ep)
		t.Fatalf("want a legal_moves reply, got %q (%s: %s)", f.Kind, ep.Code, ep.Message)
	}
	var p protocol.LegalMovesPayload
	if err := json.Unmarshal(f.Payload, &p); err != nil {
		t.Fatalf("decode legal_moves: %v", err)
	}
	return p
}

func errorCode(t *testing.T, f protocol.Frame) string {
	t.Helper()
	if f.Kind != protocol.KindError {
		t.Fatalf("want an error frame, got %q: %s", f.Kind, f.Payload)
	}
	var ep protocol.ErrorPayload
	if err := json.Unmarshal(f.Payload, &ep); err != nil {
		t.Fatal(err)
	}
	return ep.Code
}

// cappedBoard is a room whose active seat holds four Lightning Bolts
// over twenty Bears: 48 Bolt moves and four Mountains' mana past the
// wire cap, so the snapshot's legal_moves is degraded.
func cappedBoard(t *testing.T) (*Room, *game.Game, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := seedTestGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	seat := g.Seats[g.Turn.ActiveSeat].ID
	opp := g.Seats[(g.Turn.ActiveSeat+1)%2].ID
	bolts, err := g.SpawnCards(seat, seat, game.ZoneHand, game.Card{
		Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleBolt,
	}, 4)
	if err != nil {
		t.Fatalf("spawn Bolts: %v", err)
	}
	if _, err := g.SpawnCards(seat, seat, game.ZoneBattlefield, game.Card{
		Name: "Mountain", TypeLine: "Basic Land — Mountain",
	}, 4); err != nil {
		t.Fatalf("spawn Mountains: %v", err)
	}
	if _, err := g.SpawnCards(opp, opp, game.ZoneBattlefield, game.Card{
		Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2,
	}, 20); err != nil {
		t.Fatalf("spawn Bears: %v", err)
	}
	return NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), ""), g, seat, bolts
}

// The snapshot says the list was cut; the request returns the list
// before the cut, stamped with the snapshot's seq and generation, and
// a card asked for alone comes back with every target.
func TestLegalMovesRequestReturnsTheUncappedListForTheBoundSeat(t *testing.T) {
	room, g, seat, bolts := cappedBoard(t)
	conn := serveBound(t, room, Binding{GameID: g.ID, PlayerID: seat}, nil)

	initial := readSnapshotFrame(t, conn)
	if !initial.Game.LegalMovesTruncated {
		t.Fatalf("the snapshot carries %d moves and does not say they were cut", len(initial.Game.LegalMoves))
	}
	if n := len(initial.Game.LegalMoves); n > 48 {
		t.Fatalf("the snapshot carries %d moves, over the wire cap", n)
	}

	reply := legalMovesReply(t, requestLegalMoves(t, conn, ""))
	if reply.Seq != initial.Seq || reply.Generation != initial.Generation || reply.Seq != room.Seq() {
		t.Errorf("reply names seq %d generation %d, the state is seq %d generation %d (room %d)",
			reply.Seq, reply.Generation, initial.Seq, initial.Generation, room.Seq())
	}
	want, err := json.Marshal(legal.EnumerateFor(g, seat))
	if err != nil {
		t.Fatal(err)
	}
	got, err := json.Marshal(reply.Moves)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("the reply is not the uncapped enumeration:\ngot  %s\nwant %s", got, want)
	}
	if len(reply.Moves) <= len(initial.Game.LegalMoves) {
		t.Errorf("the full list (%d) is no longer than the capped one (%d)", len(reply.Moves), len(initial.Game.LegalMoves))
	}
	// Each Bolt has 22 targets and offers twelve: the enumerator's own
	// cut, reported per card.
	cut := map[string]protocol.LegalCutView{}
	for _, c := range reply.Truncated {
		cut[c.Source] = c
	}
	for _, b := range bolts {
		c, ok := cut[b.String()]
		if !ok || c.Cap != string(legal.CapPerSource) || c.Omitted != 10 || c.AtLeast {
			t.Errorf("Bolt %s: cut %+v, want per_source 10 exact (truncated %+v)", b, c, reply.Truncated)
		}
	}
	// And the digest files the same cut under each Bolt.
	for _, b := range bolts {
		src := initial.Game.LegalActions.Sources[b.String()]
		if src == nil || len(src.Truncated) != 1 || src.Truncated[0].Omitted != 10 {
			t.Errorf("digest for Bolt %s: %+v, want its per_source cut", b, src)
		}
	}

	one := legalMovesReply(t, requestLegalMoves(t, conn, `{"source":"`+bolts[0].String()+`"}`))
	if len(one.Moves) != 22 || len(one.Truncated) != 0 {
		t.Fatalf("one Bolt alone: want its 22 targets and no cut, got %d moves, %+v", len(one.Moves), one.Truncated)
	}
	for _, m := range one.Moves {
		if m.Source != bolts[0] {
			t.Errorf("a request for one card returned %q", m.Label)
		}
	}
}

// Nobody but a seat may ask, and a seat may ask only while it owes a
// decision.
func TestLegalMovesRequestRefusals(t *testing.T) {
	room, g, seat, _ := cappedBoard(t)
	other := g.Seats[(g.Turn.ActiveSeat+1)%2].ID

	for _, tc := range []struct {
		name string
		b    Binding
		code string
	}{
		{"spectator", Binding{GameID: g.ID, ReadOnly: true}, protocol.CodeBadRequest},
		{"unseated admin", Binding{GameID: g.ID, Admin: true}, protocol.CodeBadRequest},
		{"seat with no decision", Binding{GameID: g.ID, PlayerID: other}, protocol.CodeNoDecision},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conn := serveBound(t, room, tc.b, nil)
			readSnapshotFrame(t, conn)
			if code := errorCode(t, requestLegalMoves(t, conn, "")); code != tc.code {
				t.Errorf("code %q, want %q", code, tc.code)
			}
		})
	}

	// An admin bound to the seat asks as the seat.
	conn := serveBound(t, room, Binding{GameID: g.ID, PlayerID: seat, Admin: true}, nil)
	readSnapshotFrame(t, conn)
	legalMovesReply(t, requestLegalMoves(t, conn, ""))

	for _, bad := range []string{`{"source":"not-a-uuid"}`, `{"choice":"nope"}`} {
		if code := errorCode(t, requestLegalMoves(t, conn, bad)); code != protocol.CodeBadRequest {
			t.Errorf("%s: code %q, want bad_request", bad, code)
		}
	}
	if code := errorCode(t, requestLegalMoves(t, conn, `"not an object"`)); code != protocol.CodeBadJSON {
		t.Errorf("a payload that is not an object: code %q, want bad_json", code)
	}
}

// Four a second, per connection. A refused request is not counted, so
// the fifth is let in as soon as the first leaves the window.
func TestLegalMovesRequestIsRateLimited(t *testing.T) {
	room, g, seat, _ := cappedBoard(t)
	var mu sync.Mutex
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		return now
	}
	conn := serveBound(t, room, Binding{GameID: g.ID, PlayerID: seat}, clock)
	readSnapshotFrame(t, conn)

	for i := range protocol.LegalMovesRequestsPerSecond {
		if f := requestLegalMoves(t, conn, ""); f.Kind != protocol.KindLegalMoves {
			t.Fatalf("request %d of %d refused: %s", i+1, protocol.LegalMovesRequestsPerSecond, f.Payload)
		}
	}
	if code := errorCode(t, requestLegalMoves(t, conn, "")); code != protocol.CodeRateLimited {
		t.Fatalf("the fifth request in one second: code %q, want rate_limited", code)
	}
	if code := errorCode(t, requestLegalMoves(t, conn, "")); code != protocol.CodeRateLimited {
		t.Fatalf("still inside the second: code %q, want rate_limited", code)
	}
	mu.Lock()
	now = now.Add(time.Second)
	mu.Unlock()
	legalMovesReply(t, requestLegalMoves(t, conn, ""))
}

// The security pin (ADR 0122 §6.1): a connection bound to seat A that
// asks for seat B's moves gets nothing of B's. The payload has no field
// that could name a seat, so naming one — in any spelling, or by one of
// B's cards — changes nothing: the answer is A's, or A's refusal.
func TestLegalMovesRequestNeverAnswersForAnotherSeat(t *testing.T) {
	g := seedTestGame(t) // the mulligan window: both seats owe a decision
	a, b := g.Seats[0], g.Seats[1]
	var bHand []string
	g.ReadSnapshot(func() {
		for _, c := range b.Hand.Cards {
			bHand = append(bHand, c.InstanceID.String())
		}
	})
	room := NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	// Each request a second after the last, so the rate limit stays
	// out of a test about something else.
	var mu sync.Mutex
	tick := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		tick = tick.Add(time.Second)
		return tick
	}
	conn := serveBound(t, room, Binding{GameID: g.ID, PlayerID: a.ID}, clock)
	readSnapshotFrame(t, conn)

	leaksB := func(raw []byte) bool {
		s := string(raw)
		if strings.Contains(s, b.ID.String()) {
			return true
		}
		for _, id := range bHand {
			if strings.Contains(s, id) {
				return true
			}
		}
		return false
	}
	asB := []string{
		`{"player":"` + b.ID.String() + `"}`,
		`{"seat":"` + b.ID.String() + `","player_id":"` + b.ID.String() + `"}`,
		`{"source":"` + bHand[0] + `","player":"` + b.ID.String() + `"}`,
	}

	// A owes its mulligan decision: every answer is A's own list.
	for _, payload := range asB {
		f := requestLegalMoves(t, conn, payload)
		reply := legalMovesReply(t, f)
		if leaksB(f.Payload) {
			t.Fatalf("%s: the reply names seat B or a card in B's hand: %s", payload, f.Payload)
		}
		for _, m := range reply.Moves {
			if m.Player != a.ID {
				t.Fatalf("%s: a move for %s reached seat A's connection: %q", payload, m.Player, m.Label)
			}
		}
	}

	// A keeps; B still owes its decision and has moves. A asking for
	// them gets A's refusal, not B's list.
	if _, _, err := room.Apply(a.ID, func() error { return g.KeepHand(a.ID) }); err != nil {
		t.Fatalf("KeepHand: %v", err)
	}
	if len(legal.EnumerateFor(g, b.ID)) == 0 {
		t.Fatal("seat B owes nothing, so the test proves nothing")
	}
	for _, payload := range asB {
		f := requestLegalMoves(t, conn, payload)
		if code := errorCode(t, f); code != protocol.CodeNoDecision {
			t.Errorf("%s: code %q, want no_decision", payload, code)
		}
		if leaksB(f.Payload) {
			t.Errorf("%s: the refusal names seat B or a card in B's hand: %s", payload, f.Payload)
		}
	}
}
