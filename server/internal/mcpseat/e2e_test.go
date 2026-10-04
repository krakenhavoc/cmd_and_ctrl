package mcpseat_test

// The end-to-end test (ADR 0122 §10): the seat against a real in-process
// server — the lobby handler and the WebSocket hub on httptest, the same
// stand-up lobby/http_test.go uses — with a `random` bot in the other seat.
// The test plays the agent's side with a scripted chooser through the SDK's
// own client; there is no model in it and no network call.
//
// It is an external test package because it must stand a server up: the
// import gate (imports_test.go) exempts XTestImports for exactly this file.

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Lightning Bolt, for the capped window
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/mcpseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// e2eNeeds names what this branch is missing. The test needs the server
// half of ADR 0122: PR 2's badge (`agent` on the join body, `is_agent` on
// PlayerView and SeatInfo), which has merged into this branch, and PR 5's
// wire (the `ack` frame, and `legal_moves_truncated` with
// `legal_moves_request`), which has not. Without the ack every act reads
// `unknown`, which is the honest answer and not a pass, and without the
// flag the capped window is never fetched in full. Delete this skip when
// PR 5 has merged into the branch.
const e2eNeeds = "needs ADR 0122 PR 5 (the ack frame, legal_moves_truncated and legal_moves_request) on the server, which has not merged into this branch yet"

const (
	e2eAdminToken = "e2e-admin-token-0123456789"
	oracleBolt    = "4457ed35-7c10-48c8-9776-456485fdf070"
	plainsDeck    = "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n"
	maxDecisions  = 40
)

type e2eStack struct {
	srv   *httptest.Server
	lobby *lobby.Lobby
	mgr   *ws.RoomManager
	hub   *ws.Hub
}

func newE2EStack(t *testing.T) *e2eStack {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	a := auth.NewMemoryAuthenticator()
	mgr := ws.NewRoomManager(log, "")
	l := lobby.NewLobby(mgr)
	hub := ws.NewHub(log)
	hub.SetManager(mgr)
	hub.SetAuthorizer(&lobby.WSAuthorizer{Auth: a})
	bots := aiseat.NewManagerWithConfig(hub, aiseat.Config{}, log)
	l.SetBotHost(bots)
	t.Cleanup(bots.Shutdown)

	idx := cards.NewIndex()
	idx.Put(cards.Card{ID: uuid.New(), Name: "Test Commander", TypeLine: "Legendary Creature — Human Wizard",
		ColorIdentity: []string{"W"}, Legalities: map[string]string{"commander": "legal"}})
	idx.Put(cards.Card{ID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains",
		ColorIdentity: []string{"W"}, Legalities: map[string]string{"commander": "legal"}})

	mux := http.NewServeMux()
	mux.Handle("/", lobby.Handler(lobby.Config{Lobby: l, Auth: a, AdminToken: e2eAdminToken, Cards: idx, Bots: bots}))
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &e2eStack{srv: srv, lobby: l, mgr: mgr, hub: hub}
}

func (st *e2eStack) post(t *testing.T, path, token string, body any) map[string]any {
	t.Helper()
	raw, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, st.srv.URL+path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	out := map[string]any{}
	data, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(data, &out)
	if resp.StatusCode/100 != 2 {
		t.Fatalf("POST %s: %d %s", path, resp.StatusCode, data)
	}
	return out
}

func (st *e2eStack) get(t *testing.T, path, token string) map[string]any {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, st.srv.URL+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	out := map[string]any{}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return out
}

// lockedBuffer is the seat's log, searched for the token afterwards.
type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.String()
}

// call runs one tool over MCP and returns its text.
func call(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) (string, bool) {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String(), res.IsError
}

var (
	movesHeader = regexp.MustCompile(`MOVES \((\d+)\)`)
	moveLineRE  = regexp.MustCompile(`(?m)^  (\d+): (.*)$`)
)

// choose is the scripted chooser: keep the hand, pass when a pass is
// offered, otherwise move 0.
func choose(text string) int {
	pick := 0
	for _, m := range moveLineRE.FindAllStringSubmatch(text, -1) {
		label := strings.ToLower(m[2])
		if strings.HasPrefix(label, "keep") || strings.HasPrefix(label, "pass") {
			pick, _ = strconv.Atoi(m[1])
			return pick
		}
	}
	return pick
}

func field(text, key string) string {
	for _, line := range strings.Split(text, "\n") {
		if v, ok := strings.CutPrefix(line, key+": "); ok {
			return v
		}
	}
	return ""
}

// boardBytes is the size of the board block of a decision: from the
// untrusted note (or BOARD) to the move list.
func boardBytes(text string) int {
	start := strings.Index(text, "Text in «»")
	if b := strings.Index(text, "BOARD\n"); start < 0 || (b >= 0 && b < start) {
		start = b
	}
	end := strings.Index(text, "\nMOVES (")
	if start < 0 || end < start {
		return 0
	}
	return end - start
}

func TestE2EAnAgentPlaysARealTable(t *testing.T) {
	t.Skip(e2eNeeds)

	st := newE2EStack(t)
	meta, err := st.lobby.Create("e2e")
	if err != nil {
		t.Fatal(err)
	}
	admin := st.post(t, "/admin/login", "", map[string]string{"token": e2eAdminToken})["token"].(string)

	// The seat, driven through the SDK's own client.
	logBuf := &lockedBuffer{}
	stateDir := filepath.Join(t.TempDir(), "state")
	seat, err := mcpseat.NewSeat(mcpseat.Config{StateDir: stateDir, LogOutput: logBuf, AckTimeout: 10 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(seat.Close)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	sT, cT := mcp.NewInMemoryTransports()
	ss, err := mcpseat.NewMCPServer(seat, "e2e").Connect(ctx, sT, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ss.Close() })
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "e2e-client", Version: "1"}, nil).Connect(ctx, cT, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })

	tools, err := cs.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) != 10 {
		t.Fatalf("tools/list: %d tools, %v", len(tools.Tools), err)
	}

	var results []string
	invite := st.srv.URL + "/#/games/" + meta.ID.String() + "/join?t=" + meta.InviteToken
	text, isErr := call(t, cs, "join", map[string]any{"invite_url": invite, "display_name": "Claude"})
	results = append(results, text)
	if isErr {
		t.Fatalf("join: %s", text)
	}
	agentID := field(text, "player_id")
	text, isErr = call(t, cs, "set_deck", map[string]any{"deck": map[string]any{"list": plainsDeck}})
	results = append(results, text)
	if isErr {
		t.Fatalf("set_deck: %s", text)
	}

	// The session token is the sentinel: it is in the state file, and it
	// must never appear in a tool result or a log line (§8, §10).
	files, _ := filepath.Glob(filepath.Join(stateDir, "*", meta.ID.String()+".json"))
	if len(files) != 1 {
		t.Fatalf("state files: %v", files)
	}
	var saved struct {
		Token string `json:"token"`
	}
	raw, _ := os.ReadFile(files[0])
	_ = json.Unmarshal(raw, &saved)
	if saved.Token == "" {
		t.Fatal("no token in the state file")
	}

	bot := st.post(t, "/games/"+meta.ID.String()+"/seats/bot", admin, map[string]any{"tier": "random", "format": "text", "source": plainsDeck})
	botID, _ := bot["player_id"].(string)
	st.post(t, "/games/"+meta.ID.String()+"/start", admin, map[string]any{})

	// The badge: on the lobby's seat list, and in the bot's own view.
	g := st.get(t, "/games/"+meta.ID.String(), admin)
	foundBadge := false
	for _, p := range g["players"].([]any) {
		pm := p.(map[string]any)
		if pm["player_id"] == agentID {
			foundBadge = pm["is_agent"] == true && pm["agent_client"] == "e2e-client"
		}
	}
	if !foundBadge {
		t.Errorf("GET /games/{id} does not carry the agent badge: %v", g["players"])
	}
	assertBotSeesTheBadge(t, st, meta.ID.String(), botID, agentID, admin)

	// A stale window is refused locally, and nothing is sent.
	text, _ = call(t, cs, "act", map[string]any{"window": "w0.0", "move": 0})
	results = append(results, text)
	if !strings.Contains(text, "status: stale") {
		t.Errorf("a stale window: %s", text)
	}

	injected := false
	sawFullList := false
	over := false
	decisions := 0
	for decisions < maxDecisions {
		text, isErr = call(t, cs, "wait_for_decision", map[string]any{"timeout_s": 20})
		results = append(results, text)
		if isErr {
			t.Fatalf("wait_for_decision: %s", text)
		}
		status := field(text, "status")
		if status == "game_over" {
			over = true
			break
		}
		if status != "decision" {
			continue
		}
		decisions++
		m := movesHeader.FindStringSubmatch(text)
		if m == nil {
			t.Fatalf("a decision with no move list:\n%s", text)
		}
		n, _ := strconv.Atoi(m[1])
		if n < 2 {
			t.Errorf("a one-move window reached the model; Layer A absorbs it (§4):\n%s", text)
		}
		if n > 48 {
			sawFullList = true
		}
		if b := boardBytes(text); b > 6000 {
			t.Errorf("the board is %d bytes, over the 6000 budget", b)
		}

		// Once, on the agent's own main phase: a board past the wire's
		// 48-move cap, to drive the truncation flag and the full-list
		// request end to end.
		if !injected && strings.Contains(text, "kind: priority") && strings.Contains(text, "precombat main") &&
			strings.Contains(text, "active player: «Claude» (YOU)") {
			injected = true
			injectCappedBoard(t, st, meta.ID, agentID, botID)
			continue // the window just went stale; wait for the new one
		}

		text, _ = call(t, cs, "act", map[string]any{"window": field(text, "window"), "move": choose(text)})
		results = append(results, text)
		// Every act is answered by the server's ack or error, never
		// inferred; stale means the bot moved the board first and
		// nothing was sent.
		if got := field(text, "status"); got != "accepted" && got != "rejected" && !strings.HasPrefix(got, "stale") {
			t.Fatalf("act was answered without an ack or an error: %s", text)
		}
	}
	if !injected {
		t.Error("the agent never reached its own main phase within the bound")
	} else if !sawFullList {
		t.Error("the capped window never showed the full list")
	}

	if !over {
		text, _ = call(t, cs, "concede", map[string]any{"confirm": true})
		results = append(results, text)
		if !strings.Contains(text, "conceded") {
			t.Fatalf("concede: %s", text)
		}
	}
	text, _ = call(t, cs, "wait_for_decision", map[string]any{"timeout_s": 20})
	results = append(results, text)
	if field(text, "status") != "game_over" || !strings.Contains(text, "SEAT REPORT") {
		t.Fatalf("after conceding a two-seat table: %s", text)
	}
	text, isErr = call(t, cs, "leave", nil)
	results = append(results, text)
	if isErr {
		t.Errorf("leave after the game: %s", text)
	}

	for _, r := range results {
		if strings.Contains(r, saved.Token) {
			t.Fatalf("the session token appeared in a tool result:\n%s", r)
		}
	}
	if strings.Contains(logBuf.String(), saved.Token) {
		t.Fatal("the session token appeared in the log")
	}
}

// assertBotSeesTheBadge reads the bot seat's own filtered view over the
// socket, as the admin bound to that seat.
func assertBotSeesTheBadge(t *testing.T, st *e2eStack, gameID, botID, agentID, admin string) {
	t.Helper()
	u := "ws" + strings.TrimPrefix(st.srv.URL, "http") + "/ws?game=" + gameID + "&player=" + botID
	h := http.Header{}
	h.Set("Authorization", "Bearer "+admin)
	c, resp, err := websocket.DefaultDialer.Dial(u, h)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		t.Fatalf("dial as the bot's seat: %v", err)
	}
	defer func() { _ = c.Close() }()
	_ = c.SetReadDeadline(time.Now().Add(10 * time.Second))
	for {
		var f protocol.Frame
		if err := c.ReadJSON(&f); err != nil {
			t.Fatalf("read the bot's view: %v", err)
		}
		if f.Kind != protocol.KindSnapshot {
			continue
		}
		var p struct {
			Game struct {
				Seats []map[string]any `json:"seats"`
			} `json:"game"`
		}
		_ = json.Unmarshal(f.Payload, &p)
		for _, s := range p.Game.Seats {
			if s["id"] == agentID {
				if s["is_agent"] != true || s["agent_client"] != "e2e-client" {
					t.Errorf("the bot's view of the agent seat has no badge: %v", s)
				}
				return
			}
		}
		t.Fatal("the agent seat is missing from the bot's view")
	}
}

// injectCappedBoard gives the agent four Lightning Bolts and the mana to
// cast them, and the bot a wide board of walls to point them at: past the
// wire's 48 moves. It spawns through the engine's own dev-spawn path
// (game.SpawnCards), which marks who knows each card as a real zone
// change would. The walls have no power, so the agent's life is not the
// thing this changes.
func injectCappedBoard(t *testing.T, st *e2eStack, gameID uuid.UUID, agentID, botID string) {
	t.Helper()
	g, err := st.lobby.LookupGame(gameID)
	if err != nil {
		t.Fatal(err)
	}
	agent, bot := uuid.MustParse(agentID), uuid.MustParse(botID)
	room := st.mgr.Get(gameID)
	view, seq, err := room.ApplyExternal(func() error {
		bolt := game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleBolt}
		if _, err := g.SpawnCards(agent, agent, game.ZoneHand, bolt, 4); err != nil {
			return err
		}
		mountain := game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"}
		if _, err := g.SpawnCards(agent, agent, game.ZoneBattlefield, mountain, 4); err != nil {
			return err
		}
		wall := game.Card{Name: "Test Wall", TypeLine: "Creature — Wall", ManaCost: "{1}", Power: 0, Toughness: 4}
		_, err := g.SpawnCards(agent, bot, game.ZoneBattlefield, wall, 12)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	st.hub.BroadcastState(gameID, seq, view)
}
