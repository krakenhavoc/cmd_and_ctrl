package mcpseat

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// sentinelToken is the session token the fake server mints. The tests
// fail if it appears in any tool result or log line (§8, §10).
const sentinelToken = "SENTINEL-session-token-7f3a9c2e5b1d"

// fakeServer is a stand-in for the lobby and the hub: enough of both to
// drive every tool without a real game. It speaks ADR 0122's wire (the
// ack frame, legal_moves_request) unless told to behave like an older
// server.
type fakeServer struct {
	t   *testing.T
	srv *httptest.Server

	gameID, playerID, oppID uuid.UUID
	inviteToken             string

	mu        sync.Mutex
	joins     []joinBody
	joinAuth  []string
	decks     []map[string]any
	wsHeaders []http.Header
	wsQueries []string
	actions   []protocol.ActionPayload
	chats     []string
	moveReqs  []protocol.LegalMovesRequestPayload
	conn      *websocket.Conn
	wmu       sync.Mutex
	seq       uint64
	gen       uint64
	view      protocol.GameView
	moves     []legal.Move
	truncated bool
	conns     int

	// oldServer answers legal_moves_request and never acks, like a
	// server from before PR 5.
	oldServer bool
	// onAction decides an action's answer: "ack", "error", or "" for
	// nothing. Nil acks everything.
	onAction func(a protocol.ActionPayload) string
	// fullMoves is what legal_moves_request answers with.
	fullMoves []legal.Move
	cuts      []protocol.LegalCutView
	meStatus  int
	joinCode  int
	joinMsg   string
}

func newFakeServer(t *testing.T) *fakeServer {
	t.Helper()
	f := &fakeServer{
		t:           t,
		gameID:      uuid.New(),
		playerID:    uuid.New(),
		oppID:       uuid.New(),
		inviteToken: "invite-abc123",
		meStatus:    http.StatusOK,
	}
	f.view = f.baseView("lobby")
	mux := http.NewServeMux()
	mux.HandleFunc("POST /games/{id}/join", f.handleJoin)
	mux.HandleFunc("POST /games/{id}/decks", f.handleDecks)
	mux.HandleFunc("GET /me", f.handleMe)
	mux.HandleFunc("GET /cards/{id}", f.handleCard)
	mux.HandleFunc("GET /ws", f.handleWS)
	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeServer) inviteURL() string {
	return f.srv.URL + "/#/games/" + f.gameID.String() + "/join?t=" + f.inviteToken
}

func (f *fakeServer) baseView(state string) protocol.GameView {
	return protocol.GameView{
		ID:    f.gameID.String(),
		State: state,
		Seats: []protocol.PlayerView{
			{ID: f.playerID.String(), Name: "Agent", Seat: 0, Life: 40},
			{ID: f.oppID.String(), Name: "Bob", Seat: 1, Life: 40},
		},
		Turn: protocol.TurnView{Number: 1, ActiveSeat: 1, PriorityHolder: 0, Step: "main1"},
	}
}

func (f *fakeServer) handleJoin(w http.ResponseWriter, r *http.Request) {
	var body joinBody
	_ = json.NewDecoder(r.Body).Decode(&body)
	f.mu.Lock()
	f.joins = append(f.joins, body)
	f.joinAuth = append(f.joinAuth, r.Header.Get("Authorization"))
	code, msg := f.joinCode, f.joinMsg
	f.mu.Unlock()
	if code != 0 {
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
		return
	}
	if body.InviteToken != f.inviteToken {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invite token did not match"}`))
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"token":      sentinelToken,
		"expires_at": time.Now().Add(12 * time.Hour),
		"player_id":  f.playerID,
		"principal":  map[string]any{"role": "player", "game_id": f.gameID, "player_id": f.playerID},
	})
}

func (f *fakeServer) authed(r *http.Request) bool {
	return r.Header.Get("Authorization") == "Bearer "+sentinelToken
}

func (f *fakeServer) handleMe(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	status := f.meStatus
	f.mu.Unlock()
	if !f.authed(r) || status != http.StatusOK {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"role": "player", "game_id": f.gameID, "player_id": f.playerID})
}

func (f *fakeServer) handleDecks(w http.ResponseWriter, r *http.Request) {
	if !f.authed(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	f.mu.Lock()
	f.decks = append(f.decks, body)
	f.mu.Unlock()
	if body["deck"] == "nope" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"unknown deck \"nope\"; this server has: izzet-aggro, mono-white"}`))
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"deck_name": "Mono White", "card_count": 100, "commanders": []string{"Test Commander"}})
}

func (f *fakeServer) handleCard(w http.ResponseWriter, r *http.Request) {
	if !f.authed(r) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"name": "Grizzly Bears", "type_line": "Creature — Bear", "mana_cost": "{1}{G}",
		"oracle_text": "", "power": "2", "toughness": "2",
	})
}

func (f *fakeServer) handleWS(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.wsHeaders = append(f.wsHeaders, r.Header.Clone())
	f.wsQueries = append(f.wsQueries, r.URL.RawQuery)
	f.mu.Unlock()
	if !f.authed(r) || r.URL.Query().Get("token") != "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	up := websocket.Upgrader{}
	c, err := up.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	f.mu.Lock()
	f.conn = c
	f.conns++
	f.mu.Unlock()
	f.sendSnapshot()
	for {
		_, raw, err := c.ReadMessage()
		if err != nil {
			return
		}
		var fr protocol.Frame
		_ = json.Unmarshal(raw, &fr)
		switch string(fr.Kind) {
		case string(protocol.KindAction):
			var a protocol.ActionPayload
			_ = json.Unmarshal(fr.Payload, &a)
			f.mu.Lock()
			f.actions = append(f.actions, a)
			decide, old := f.onAction, f.oldServer
			f.mu.Unlock()
			answer := "ack"
			if decide != nil {
				answer = decide(a)
			}
			switch answer {
			case "ack":
				if !old {
					f.mu.Lock()
					seq, gen := f.seq, f.gen
					f.mu.Unlock()
					f.write(protocol.Frame{V: 0, Kind: protocol.KindAck, ID: fr.ID, Payload: mustJSON(protocol.AckPayload{Seq: seq, Generation: gen})})
				}
			case "error":
				f.write(protocol.Frame{V: 0, Kind: protocol.KindError, ID: fr.ID,
					Payload: mustJSON(protocol.ErrorPayload{Code: "bad_request", Message: "Bob says no"})})
			}
		case string(protocol.KindLegalMovesRequest):
			var req protocol.LegalMovesRequestPayload
			_ = json.Unmarshal(fr.Payload, &req)
			f.mu.Lock()
			f.moveReqs = append(f.moveReqs, req)
			old := f.oldServer
			rep := protocol.LegalMovesPayload{Seq: f.seq, Generation: f.gen, Moves: f.fullMoves, Truncated: f.cuts}
			f.mu.Unlock()
			if old {
				f.write(protocol.Frame{V: 0, Kind: protocol.KindError, ID: fr.ID,
					Payload: mustJSON(protocol.ErrorPayload{Code: protocol.CodeBadRequest, Message: "unknown or unsupported kind"})})
				continue
			}
			f.write(protocol.Frame{V: 0, Kind: protocol.KindLegalMoves, ID: fr.ID, Payload: mustJSON(rep)})
		case string(protocol.KindChat):
			var p protocol.ChatPayload
			_ = json.Unmarshal(fr.Payload, &p)
			f.mu.Lock()
			f.chats = append(f.chats, p.Text)
			f.mu.Unlock()
		}
	}
}

func mustJSON(v any) json.RawMessage {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return raw
}

func (f *fakeServer) write(fr protocol.Frame) {
	f.mu.Lock()
	c := f.conn
	f.mu.Unlock()
	if c == nil {
		return
	}
	f.wmu.Lock()
	defer f.wmu.Unlock()
	_ = c.WriteJSON(fr)
}

// setState replaces the view and the seat's move list and broadcasts a
// snapshot with the next seq.
func (f *fakeServer) setState(v protocol.GameView, moves []legal.Move, truncated bool) {
	f.mu.Lock()
	f.view, f.moves, f.truncated = v, moves, truncated
	f.mu.Unlock()
	f.sendSnapshot()
}

func (f *fakeServer) sendSnapshot() {
	f.mu.Lock()
	f.seq++
	v, moves, trunc, seq, gen := f.view, f.moves, f.truncated, f.seq, f.gen
	f.mu.Unlock()
	game := map[string]any{}
	raw, _ := json.Marshal(v)
	_ = json.Unmarshal(raw, &game)
	if len(moves) > 0 {
		game["legal_moves"] = moves
	}
	if trunc {
		game["legal_moves_truncated"] = true
	}
	payload := mustJSON(map[string]any{"seq": seq, "generation": gen, "game": game})
	f.write(protocol.Frame{V: 0, Kind: protocol.KindSnapshot, Payload: payload})
}

func (f *fakeServer) chat(author, text string) {
	f.write(protocol.Frame{V: 0, Kind: protocol.KindChat, Payload: mustJSON(protocol.ChatPayload{AuthorID: f.oppID.String(), AuthorName: author, Text: text})})
}

func (f *fakeServer) actionCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.actions)
}

func (f *fakeServer) lastAction() protocol.ActionPayload {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.actions[len(f.actions)-1]
}

// --- moves ---------------------------------------------------------------

func (f *fakeServer) pass() legal.Move {
	return legal.Move{Type: legal.TypePassPriority, Player: f.playerID, Kind: legal.KindPass, Label: "Pass priority", AlwaysLegal: true}
}

func (f *fakeServer) mana(card uuid.UUID) legal.Move {
	return legal.Move{Type: legal.TypeActivateManaAbility, Player: f.playerID, Kind: legal.KindMana, Label: "Tap Forest for {G}", Source: card,
		Params: mustJSON(map[string]any{"instance_id": card})}
}

func (f *fakeServer) cast(card uuid.UUID, label string) legal.Move {
	return legal.Move{Type: legal.TypeCastSpell, Player: f.playerID, Kind: legal.KindCast, Label: label, Source: card,
		Params: mustJSON(map[string]any{"instance_id": card})}
}

func (f *fakeServer) land(card uuid.UUID, label string) legal.Move {
	return legal.Move{Type: "play_land", Player: f.playerID, Kind: legal.KindLand, Label: label, Source: card,
		Params: mustJSON(map[string]any{"instance_id": card})}
}

// --- the seat ------------------------------------------------------------

// testSeat builds a seat whose log goes to a buffer the test can search.
type testSeat struct {
	*Seat
	logBuf *syncBuffer
}

type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.String()
}

func newTestSeat(t *testing.T, mutate func(*Config)) *testSeat {
	t.Helper()
	buf := &syncBuffer{}
	cfg := Config{StateDir: t.TempDir() + "/state", LogOutput: buf, AckTimeout: 2 * time.Second}
	if mutate != nil {
		mutate(&cfg)
	}
	s, err := NewSeat(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	return &testSeat{Seat: s, logBuf: buf}
}

// noSentinel fails if the session token reached the text.
func noSentinel(t *testing.T, where, text string) {
	t.Helper()
	if strings.Contains(text, sentinelToken) {
		t.Fatalf("the session token appeared in %s:\n%s", where, text)
	}
}

// waitFor polls a monotone predicate (AGENTS.md §5, #634). The budget is
// a backstop, never an assertion.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// closeConn closes the socket with a close code, as hub.Shutdown (1001)
// or hub.EvictGame (1000) would.
func (f *fakeServer) closeConn(code int) {
	f.mu.Lock()
	c := f.conn
	f.conn = nil
	f.mu.Unlock()
	if c == nil {
		return
	}
	f.wmu.Lock()
	_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, "bye"), time.Now().Add(time.Second))
	f.wmu.Unlock()
	_ = c.Close()
}

func (f *fakeServer) connCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.conns
}
