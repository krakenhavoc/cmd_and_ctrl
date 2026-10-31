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
	"fmt"
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
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Lightning Bolt and Anguished Unmaking, for the capped windows
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/mcpseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

const (
	e2eAdminToken = "e2e-admin-token-0123456789"
	oracleBolt    = "4457ed35-7c10-48c8-9776-456485fdf070"
	// Anguished Unmaking: "Exile target nonland permanent." (#4737)
	oracleAnguishedUnmaking = "ad09b3c3-c8e7-481c-8c45-e7f234935117"
	plainsDeck              = "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n"
	maxDecisions            = 40
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

// choose is the scripted chooser: keep the hand, take the first turn
// when it wins the opening roll, pass when a pass is offered, otherwise
// move 0.
func choose(text string) int {
	pick := 0
	for _, m := range moveLineRE.FindAllStringSubmatch(text, -1) {
		label := strings.ToLower(m[2])
		if strings.HasPrefix(label, "keep") || strings.HasPrefix(label, "pass") || strings.HasPrefix(label, "i go first") {
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
	st := newE2EStack(t)
	meta, err := st.lobby.Create("e2e")
	if err != nil {
		t.Fatal(err)
	}
	admin := st.post(t, "/admin/login", "", map[string]string{"token": e2eAdminToken})["token"].(string)

	cs, logBuf, stateDir := newE2ESeat(t)
	ctx := context.Background()
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
		// ADR 0121: the lobby opens the opening roll. The agent's die is
		// absorbed; if it wins, the choice of who goes first is the one
		// opening-roll window the model sees, named for what it is.
		if strings.Contains(text, "kind: opening_roll") {
			t.Errorf("the agent's opening die reached the model; Layer A rolls it:\n%s", text)
		}
		if strings.Contains(text, "kind: choice:starting_player") &&
			(!strings.Contains(text, "I go first") || !strings.Contains(text, "goes first")) {
			t.Errorf("the opening roll's choice does not offer every seat:\n%s", text)
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

// newE2ESeat is the seat, driven through the SDK's own client: the
// client session, the seat's log and its state directory.
func newE2ESeat(t *testing.T) (*mcp.ClientSession, *lockedBuffer, string) {
	t.Helper()
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
	return cs, logBuf, stateDir
}

// #4737, against the real enumerator: a single-target removal spell over
// more nonland permanents than the enumerator's per-card cap, on a list
// far under the wire's 48-move cap, so legal_moves_truncated stays off and
// only the spell's legal_actions entry says it was cut. The window offers
// every target, folds the identical Treasures to one line, and the report
// counts the window as truncated.
func TestE2EACardCutUnderTheWireCapOffersEveryTarget(t *testing.T) {
	st := newE2EStack(t)
	meta, err := st.lobby.Create("e2e-cut")
	if err != nil {
		t.Fatal(err)
	}
	admin := st.post(t, "/admin/login", "", map[string]string{"token": e2eAdminToken})["token"].(string)
	cs, _, _ := newE2ESeat(t)
	invite := st.srv.URL + "/#/games/" + meta.ID.String() + "/join?t=" + meta.InviteToken
	text, isErr := call(t, cs, "join", map[string]any{"invite_url": invite, "display_name": "Claude"})
	if isErr {
		t.Fatalf("join: %s", text)
	}
	agentID := field(text, "player_id")
	if text, isErr = call(t, cs, "set_deck", map[string]any{"deck": map[string]any{"list": plainsDeck}}); isErr {
		t.Fatalf("set_deck: %s", text)
	}
	bot := st.post(t, "/games/"+meta.ID.String()+"/seats/bot", admin, map[string]any{"tier": "random", "format": "text", "source": plainsDeck})
	botID, _ := bot["player_id"].(string)
	st.post(t, "/games/"+meta.ID.String()+"/start", admin, map[string]any{})

	injected, checked := false, false
	for decisions := 0; decisions < maxDecisions && !checked; {
		text, isErr = call(t, cs, "wait_for_decision", map[string]any{"timeout_s": 20})
		if isErr {
			t.Fatalf("wait_for_decision: %s", text)
		}
		if s := field(text, "status"); s == "game_over" {
			break
		} else if s != "decision" {
			continue
		}
		decisions++
		switch {
		case injected && strings.Contains(text, "Anguished Unmaking {1}{W}{B} (hand)"):
			checked = true
			for i := 1; i <= cutRelics; i++ {
				if want := fmt.Sprintf("Cast Anguished Unmaking targeting Test Relic %02d (", i); !strings.Contains(text, want) {
					t.Errorf("Test Relic %02d is a legal target and the window does not offer it", i)
				}
			}
			if !regexp.MustCompile(`Cast Anguished Unmaking targeting Treasure \(.*\) \[and 5 more like it, each naming another identical permanent: moves \d`).MatchString(text) {
				t.Error("the six identical Treasures are not one line that counts the other five")
			}
			if strings.Contains(text, "not listed") {
				t.Error("a card fetched in full still says it was cut")
			}
			if t.Failed() {
				t.Logf("the window:\n%s", text)
			}
			continue
		case !injected && strings.Contains(text, "kind: priority") && strings.Contains(text, "precombat main") &&
			strings.Contains(text, "active player: «Claude» (YOU)"):
			injected = true
			injectCutCard(t, st, meta.ID, agentID, botID)
			continue // the window just went stale; wait for the new one
		}
		text, _ = call(t, cs, "act", map[string]any{"window": field(text, "window"), "move": choose(text)})
		if got := field(text, "status"); got != "accepted" && got != "rejected" && !strings.HasPrefix(got, "stale") {
			t.Fatalf("act was answered without an ack or an error: %s", text)
		}
	}
	if !checked {
		t.Fatal("the agent never saw the removal spell's window within the bound")
	}
	if text, _ = call(t, cs, "concede", map[string]any{"confirm": true}); !strings.Contains(text, "conceded") {
		t.Fatalf("concede: %s", text)
	}
	text, _ = call(t, cs, "wait_for_decision", map[string]any{"timeout_s": 20})
	m := regexp.MustCompile(`truncated windows: (\d+);`).FindStringSubmatch(text)
	if m == nil {
		t.Fatalf("no report after conceding: %s", text)
	}
	if n, _ := strconv.Atoi(m[1]); n < 1 {
		t.Errorf("the report counts %d truncated windows; the removal spell's window was cut", n)
	}
}

// cutRelics is how many distinct artifacts injectCutCard gives the bot.
const cutRelics = 20

// injectCutCard puts Anguished Unmaking in the agent's hand with the mana
// to cast it, and gives the bot twenty distinct artifacts and six
// Treasures: 26 legal targets for one spell, over twice the enumerator's
// per-card cap, on a list well under the wire cap.
func injectCutCard(t *testing.T, st *e2eStack, gameID uuid.UUID, agentID, botID string) {
	t.Helper()
	g, err := st.lobby.LookupGame(gameID)
	if err != nil {
		t.Fatal(err)
	}
	agent, bot := uuid.MustParse(agentID), uuid.MustParse(botID)
	room := st.mgr.Get(gameID)
	var spell uuid.UUID
	view, seq, err := room.ApplyExternal(func() error {
		au := game.Card{Name: "Anguished Unmaking", TypeLine: "Instant", ManaCost: "{1}{W}{B}", OracleID: oracleAnguishedUnmaking}
		ids, err := g.SpawnCards(agent, agent, game.ZoneHand, au, 1)
		if err != nil {
			return err
		}
		spell = ids[0]
		for _, land := range []game.Card{{Name: "Swamp", TypeLine: "Basic Land — Swamp"}, {Name: "Plains", TypeLine: "Basic Land — Plains"}} {
			if _, err := g.SpawnCards(agent, agent, game.ZoneBattlefield, land, 2); err != nil {
				return err
			}
		}
		for i := 1; i <= cutRelics; i++ {
			relic := game.Card{Name: fmt.Sprintf("Test Relic %02d", i), TypeLine: "Artifact", ManaCost: "{1}"}
			if _, err := g.SpawnCards(agent, bot, game.ZoneBattlefield, relic, 1); err != nil {
				return err
			}
		}
		treasure := game.Card{Name: "Treasure", TypeLine: "Token Artifact — Treasure"}
		_, err = g.SpawnCards(agent, bot, game.ZoneBattlefield, treasure, 6)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	// The shape #4737 needs, from the server's own enumeration: the spell
	// cut, the list under the wire cap.
	rep := legal.EnumerateReport(g, agent, legal.Options{})
	if len(rep.Moves) > 48 {
		t.Fatalf("%d moves: past the wire cap, so this is not the board #4737 is about", len(rep.Moves))
	}
	cut := false
	for _, c := range rep.Cuts {
		cut = cut || (c.Source == spell && c.Cap == legal.CapPerSource)
	}
	if !cut {
		t.Fatalf("the enumerator did not cut the spell: %+v", rep.Cuts)
	}
	st.hub.BroadcastState(gameID, seq, view)
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
