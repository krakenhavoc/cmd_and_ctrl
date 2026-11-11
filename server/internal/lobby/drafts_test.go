package lobby

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/capacity"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draft"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draftpacks"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// The draft as the lobby serves it (ADR 0222 §5, #5438): its routes,
// its socket, the load cap, the idle sweep, and that none of it touches
// a table. The draft's own rules are tested in internal/draft.

var draftTestNS = uuid.MustParse("00000000-0000-0000-0000-0000000d5438")

func draftTestID(parts ...any) uuid.UUID {
	return uuid.NewSHA1(draftTestNS, []byte(fmt.Sprint(parts...)))
}

// draftTestIndex holds two made-up sets: "syn", every card of which the
// test's catalog verdict calls Full, and "wip", which has one card that
// is not.
func draftTestIndex() *cards.Index {
	idx := cards.NewIndex()
	for _, set := range []string{"syn", "wip"} {
		n := 1
		put := func(prefix, rarity string, count int) {
			for i := range count {
				name := fmt.Sprintf("%s %s %02d", set, prefix, i)
				idx.Put(cards.Card{
					ID: draftTestID("print", set, n), OracleID: draftTestID("oracle", name), Name: name,
					SetCode: set, SetName: "Synthetic " + set, SetType: "expansion",
					CollectorNo: fmt.Sprint(n), Lang: "en", Layout: "normal",
					TypeLine: "Creature — Test", BorderColor: "black", ReleasedAt: "2026-03-01",
					Rarity: rarity, Booster: true, Colors: []string{"G"}, ManaCost: "{1}{G}",
				})
				n++
			}
		}
		put("Common", "common", 80)
		put("Uncommon", "uncommon", 100)
		put("Rare", "rare", 60)
		put("Mythic", "mythic", 20)
	}
	return idx
}

var draftNotFull = draftTestID("oracle", "wip Common 07")

// draftStack is the lobby's handler over a real hub, as main wires it.
type draftStack struct {
	srv   *httptest.Server
	lobby *Lobby
	auth  auth.Authenticator
	hub   *ws.Hub
}

func newDraftStack(t *testing.T) *draftStack {
	t.Helper()
	// The flows make more calls than the public buckets allow; the
	// limiters have their own test below.
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	return newDraftStackLimited(t)
}

func newDraftStackLimited(t *testing.T) *draftStack {
	t.Helper()
	return newDraftStackWith(t, nil)
}

// newDraftStackWith is newDraftStackLimited with the lobby's Config
// adjusted before the handler is built.
func newDraftStackWith(t *testing.T, adjust func(*Config)) *draftStack {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	l := NewLobby(mgr)
	hub := ws.NewHub(log)
	hub.SetManager(mgr)
	a := auth.NewMemoryAuthenticator()
	hub.SetAuthorizer(&WSAuthorizer{Auth: a})
	idx := draftTestIndex()
	cfg := Config{
		Lobby: l, Auth: a, AdminToken: "shared-admin-token", Cards: idx,
		DraftSets: draftpacks.NewListerWith(idx, func(id uuid.UUID) bool { return id != draftNotFull }),
		DraftHub:  hub,
	}
	if adjust != nil {
		adjust(&cfg)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", hub.ServeWS)
	mux.Handle("/", Handler(cfg))
	srv := httptest.NewServer(mux)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		hub.Shutdown(ctx)
		srv.Close()
	})
	return &draftStack{srv: srv, lobby: l, auth: a, hub: hub}
}

// signIn mints a Discord sign-in, as the OAuth callback does.
func (s *draftStack) signIn(t *testing.T, name string) (token string, user uuid.UUID) {
	t.Helper()
	user = uuid.New()
	tok, _, err := s.auth.Issue(context.Background(), auth.Principal{
		Role: auth.RoleIdentified, UserID: user, DiscordID: "9" + strings.Repeat("1", 16), DiscordGlobalName: name,
	}, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	return tok, user
}

type draftAnswer struct {
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	Draft     draft.View `json:"draft"`
	Error     string     `json:"error"`
	Code      string     `json:"code"`
}

// call makes one request with an optional Bearer token and decodes the
// JSON answer.
func (s *draftStack) call(t *testing.T, method, path, token, body string) (*http.Response, draftAnswer) {
	t.Helper()
	req, err := http.NewRequest(method, s.srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := s.srv.Client().Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out draftAnswer
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatalf("%s %s: %d, body is not JSON: %s", method, path, resp.StatusCode, raw)
		}
	}
	return resp, out
}

// create makes a draft and fails the test unless it is made.
func (s *draftStack) create(t *testing.T, token, body string) draftAnswer {
	t.Helper()
	resp, out := s.call(t, http.MethodPost, "/draft", token, body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /draft: %d %s", resp.StatusCode, out.Error)
	}
	return out
}

// draftConn is one draft socket and every frame it has been sent.
type draftConn struct {
	t    *testing.T
	conn *websocket.Conn
	raw  [][]byte
}

func (s *draftStack) dial(t *testing.T, draftID uuid.UUID, token string) *draftConn {
	t.Helper()
	u := "ws" + strings.TrimPrefix(s.srv.URL, "http") + "/draft/" + draftID.String() + "/ws?token=" + token
	conn, resp, err := websocket.DefaultDialer.Dial(u, nil)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		t.Fatalf("draft socket: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return &draftConn{t: t, conn: conn}
}

// dialStatus is the HTTP status a refused draft socket gets.
func (s *draftStack) dialStatus(t *testing.T, path string) int {
	t.Helper()
	u := "ws" + strings.TrimPrefix(s.srv.URL, "http") + path
	conn, resp, err := websocket.DefaultDialer.Dial(u, nil)
	if err == nil {
		_ = conn.Close()
		return http.StatusSwitchingProtocols
	}
	if resp == nil {
		t.Fatalf("dial %s: %v", path, err)
	}
	_ = resp.Body.Close()
	return resp.StatusCode
}

// next reads frames until a draft frame arrives and returns its view.
// An error frame fails the test.
func (c *draftConn) next() draft.View {
	c.t.Helper()
	for {
		_ = c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			c.t.Fatalf("read draft frame: %v", err)
		}
		c.raw = append(c.raw, raw)
		var f protocol.Frame
		if err := json.Unmarshal(raw, &f); err != nil {
			c.t.Fatalf("frame is not JSON: %s", raw)
		}
		switch f.Kind {
		case protocol.KindDraft:
			var v draft.View
			if err := json.Unmarshal(f.Payload, &v); err != nil {
				c.t.Fatalf("draft payload: %v", err)
			}
			return v
		case protocol.KindError:
			c.t.Fatalf("error frame: %s", f.Payload)
		}
	}
}

// drain reads and discards every frame from here on, until the socket
// closes. A test that only needs a seat to stay connected calls it: a
// socket nobody reads from can hit the TCP buffers, and the hub drops a
// client whose send buffer fills (#5696). It must be the only reader:
// do not call next or until after it.
func (c *draftConn) drain() {
	go func() {
		for {
			if _, _, err := c.conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}

// until reads views until one satisfies ok. Views only move forward,
// so this is a wait on a monotone condition.
func (c *draftConn) until(what string, ok func(draft.View) bool) draft.View {
	c.t.Helper()
	for range 2000 {
		if v := c.next(); ok(v) {
			return v
		}
	}
	c.t.Fatalf("never saw: %s", what)
	return draft.View{}
}

func (c *draftConn) send(kind protocol.Kind, payload any) string {
	c.t.Helper()
	body, _ := json.Marshal(payload)
	id := uuid.NewString()
	raw, _ := json.Marshal(protocol.Frame{V: protocol.Version, Kind: kind, ID: id, Payload: body})
	if err := c.conn.WriteMessage(websocket.TextMessage, raw); err != nil {
		c.t.Fatalf("write %s: %v", kind, err)
	}
	return id
}

// pick takes the first cards of the pack in v.
func (c *draftConn) pick(v draft.View) {
	c.t.Helper()
	ids := make([]string, 0, v.Take)
	for _, card := range v.Pack[:v.Take] {
		ids = append(ids, card.ID)
	}
	c.send(protocol.KindDraftPick, protocol.DraftPickPayload{Cards: ids})
}

// expectError reads until the error frame answering id.
func (c *draftConn) expectError(id string) protocol.ErrorPayload {
	c.t.Helper()
	for range 200 {
		_ = c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			c.t.Fatalf("read: %v", err)
		}
		c.raw = append(c.raw, raw)
		var f protocol.Frame
		_ = json.Unmarshal(raw, &f)
		if f.Kind == protocol.KindError && f.ID == id {
			var e protocol.ErrorPayload
			_ = json.Unmarshal(f.Payload, &e)
			return e
		}
	}
	c.t.Fatal("no error frame")
	return protocol.ErrorPayload{}
}

// --- a visitor drafts against bots ------------------------------------------

// A visitor with no session makes a draft, starts it and drafts all 42
// cards against seven bots, over the socket, with no input but their
// own. This is "one person and seven bots is a normal draft" end to
// end.
func TestAVisitorDraftsAgainstSevenBots(t *testing.T) {
	s := newDraftStack(t)
	resp, out := s.call(t, http.MethodPost, "/draft", "", `{"set":"syn"}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /draft with no session: %d %s", resp.StatusCode, out.Error)
	}
	if _, ok := sessionCookie(resp); ok {
		t.Fatal("creating a draft set the session cookie; the seat token belongs in the body only")
	}
	v := out.Draft
	if !v.Solo || v.Invite != "" || v.State != draft.StateLobby || v.Seats != 8 || v.Mode != draft.PickOne ||
		v.Set != "syn" || v.SetName != "Synthetic syn" || v.Booster != "play" || v.PickSeconds != 60 || v.Name == "" {
		t.Fatalf("a visitor's draft: %+v", v)
	}
	if len(v.Players) != 1 || !v.Players[0].Host || v.Players[0].Name != "Player" || v.You != 0 {
		t.Fatalf("players: %+v", v.Players)
	}
	p, err := s.auth.Validate(context.Background(), out.Token)
	if err != nil || p.Role != auth.RoleDrafter || p.DraftID != v.ID || p.PlayerID == uuid.Nil || p.UserID != uuid.Nil || p.GameID != uuid.Nil {
		t.Fatalf("the seat token: %+v, %v", p, err)
	}

	conn := s.dial(t, v.ID, out.Token)
	if first := conn.next(); first.State != draft.StateLobby || !first.Players[0].Connected {
		t.Fatalf("first frame: %+v", first)
	}
	resp, started := s.call(t, http.MethodPost, "/draft/"+v.ID.String()+"/start", out.Token, "")
	if resp.StatusCode != http.StatusOK || started.Draft.State != draft.StatePicking {
		t.Fatalf("start: %d %s", resp.StatusCode, started.Error)
	}
	picks := 0
	cur := conn.until("a pack", func(v draft.View) bool { return len(v.Pack) > 0 })
	if len(cur.Players) != 8 {
		t.Fatalf("%d seats after the start, want 8", len(cur.Players))
	}
	for cur.State == draft.StatePicking {
		conn.pick(cur)
		picks++
		want := picks
		cur = conn.until("the next pack or the end", func(v draft.View) bool {
			return v.State == draft.StateDone || (len(v.Picks) == want && len(v.Pack) > 0)
		})
	}
	if picks != 42 || len(cur.Pool) != 42 || len(cur.Picks) != 42 {
		t.Fatalf("%d picks, pool %d", picks, len(cur.Pool))
	}
	for _, pl := range cur.Players {
		if pl.Picked != 42 {
			t.Fatalf("seat %d (%s) drafted %d cards, want 42", pl.Seat, pl.Name, pl.Picked)
		}
	}
	// The pools are held for slice 6 to read.
	pools, done := s.lobby.DraftRoom(v.ID).Pools()
	if !done || len(pools) != 8 {
		t.Fatalf("Pools: done %v, %d seats", done, len(pools))
	}
}

// A draft is not a table: making one adds no game, no room and no row,
// and a Commander table made beside it is what it always was.
func TestADraftIsNotATable(t *testing.T) {
	s := newDraftStack(t)
	meta, err := s.lobby.Create("Commander night")
	if err != nil {
		t.Fatal(err)
	}
	before := s.lobby.RoomOf(meta.ID).Game.TableSettingsSnapshot()
	out := s.create(t, "", `{"set":"syn","seats":4}`)
	if got := s.lobby.List(); len(got) != 1 || got[0].ID != meta.ID {
		t.Fatalf("the table list after a draft was made: %d tables", len(got))
	}
	if _, err := s.lobby.Get(out.Draft.ID); err == nil {
		t.Fatal("a draft answers as a table")
	}
	if s.lobby.RoomOf(out.Draft.ID) != nil {
		t.Fatal("a draft has a game room")
	}
	if s.lobby.DraftCount() != 1 {
		t.Fatalf("%d drafts", s.lobby.DraftCount())
	}
	after := s.lobby.RoomOf(meta.ID).Game.TableSettingsSnapshot()
	if before.Format != after.Format || after.EffectiveFormat() != game.FormatCommander {
		t.Fatalf("the Commander table's format moved: %q to %q", before.Format, after.Format)
	}
}

// --- which sets ----------------------------------------------------------------

func TestOnlyADraftableSetCanBeDrafted(t *testing.T) {
	s := newDraftStack(t)
	for _, body := range []string{`{"set":"wip"}`, `{"set":"nope"}`} {
		resp, out := s.call(t, http.MethodPost, "/draft", "", body)
		if resp.StatusCode != http.StatusUnprocessableEntity || out.Code != CodeSetNotDraftable {
			t.Fatalf("POST /draft %s: %d code %q, want 422 %s", body, resp.StatusCode, out.Code, CodeSetNotDraftable)
		}
	}
	if s.lobby.DraftCount() != 0 {
		t.Fatal("a refused draft was made")
	}
	for body, want := range map[string]int{
		`{}`:                               http.StatusBadRequest,
		`{"set":"syn","mode":"pick_nine"}`: http.StatusBadRequest,
		`{"set":"syn","seats":1}`:          http.StatusBadRequest,
		`{"set":"syn","seats":9}`:          http.StatusBadRequest,
		`{"set":"syn","pick_seconds":5}`:   http.StatusBadRequest,
		`{"set":"syn","pick_seconds":900}`: http.StatusBadRequest,
		`{"set":" SYN ","mode":"pick_two","seats":4,"pick_seconds":30}`: http.StatusCreated,
	} {
		if resp, out := s.call(t, http.MethodPost, "/draft", "", body); resp.StatusCode != want {
			t.Errorf("POST /draft %s: %d (%s), want %d", body, resp.StatusCode, out.Error, want)
		}
	}
}

func TestADraftNameGoesThroughTheNameFilter(t *testing.T) {
	s := newDraftStack(t)
	resp, out := s.call(t, http.MethodPost, "/draft", "", `{"set":"syn","name":"kys everyone"}`)
	if resp.StatusCode != http.StatusBadRequest || out.Code != CodeNameNotAllowed {
		t.Fatalf("a filtered name: %d code %q", resp.StatusCode, out.Code)
	}
	ok := s.create(t, "", `{"set":"syn","name":"Friday draft"}`)
	if ok.Draft.Name != "Friday draft" {
		t.Fatalf("name %q", ok.Draft.Name)
	}
}

// --- a signed-in host, an invite and a guest ------------------------------------

func TestInviteJoinStartAndThePickTwoDraft(t *testing.T) {
	s := newDraftStack(t)
	hostTok, hostUser := s.signIn(t, "Ian")
	host := s.create(t, hostTok, `{"set":"syn","mode":"pick_two","seats":4,"name":"Four of us"}`)
	v := host.Draft
	if v.Solo || v.Invite == "" || v.Players[0].Name != "Ian" || v.Mode != draft.PickTwo {
		t.Fatalf("a signed-in host's draft: %+v", v)
	}
	if p, _ := s.auth.Validate(context.Background(), host.Token); p.UserID != hostUser || p.Role != auth.RoleDrafter {
		t.Fatalf("the host's seat token: %+v", p)
	}
	base := "/draft/" + v.ID.String()

	// The invite page.
	resp, _ := s.call(t, http.MethodGet, base+"?invite=wrong", "", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("preview with a wrong invite: %d, want 404", resp.StatusCode)
	}
	resp, _ = s.call(t, http.MethodGet, base, "", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("preview with no invite: %d, want 404", resp.StatusCode)
	}
	var preview draft.Preview
	func() {
		r, err := s.srv.Client().Get(s.srv.URL + base + "?invite=" + v.Invite)
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()
		if r.StatusCode != http.StatusOK {
			t.Fatalf("preview: %d", r.StatusCode)
		}
		_ = json.NewDecoder(r.Body).Decode(&preview)
	}()
	if preview.Name != "Four of us" || preview.Taken != 1 || preview.Seats != 4 || preview.Host != "Ian" || preview.Set != "syn" {
		t.Fatalf("preview: %+v", preview)
	}

	// A guest joins by name; a wrong invite is the same 404 as no draft.
	resp, _ = s.call(t, http.MethodPost, base+"/join", "", `{"invite":"wrong","name":"Sam"}`)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("join with a wrong invite: %d, want 404", resp.StatusCode)
	}
	resp, _ = s.call(t, http.MethodPost, "/draft/"+uuid.NewString()+"/join", "", `{"invite":"`+v.Invite+`"}`)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("join of no draft: %d, want 404", resp.StatusCode)
	}
	resp, bad := s.call(t, http.MethodPost, base+"/join", "", `{"invite":"`+v.Invite+`","name":"kys"}`)
	if resp.StatusCode != http.StatusBadRequest || bad.Code != CodeNameNotAllowed {
		t.Fatalf("join under a filtered name: %d %q", resp.StatusCode, bad.Code)
	}
	resp, guest := s.call(t, http.MethodPost, base+"/join", "", `{"invite":"`+v.Invite+`","name":"Sam"}`)
	if resp.StatusCode != http.StatusOK || guest.Draft.You != 1 || guest.Draft.Players[1].Name != "Sam" {
		t.Fatalf("join: %d %+v", resp.StatusCode, guest.Draft.Players)
	}
	if guest.Draft.Invite != "" {
		t.Fatal("a guest's view carries the invite; it is the host's")
	}
	if _, ok := sessionCookie(resp); ok {
		t.Fatal("joining a draft set the session cookie")
	}

	// The host coming back from another device gets the same seat.
	resp, again := s.call(t, http.MethodPost, base+"/join", hostTok, `{"invite":"`+v.Invite+`"}`)
	if resp.StatusCode != http.StatusOK || again.Draft.You != 0 || len(again.Draft.Players) != 2 {
		t.Fatalf("the host joining again: %d you=%d players=%d", resp.StatusCode, again.Draft.You, len(again.Draft.Players))
	}

	// Only the host starts.
	resp, _ = s.call(t, http.MethodPost, base+"/start", guest.Token, "")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("start by a guest: %d, want 403", resp.StatusCode)
	}
	hc, gc := s.dial(t, v.ID, host.Token), s.dial(t, v.ID, guest.Token)
	hc.next()
	gc.next()
	resp, _ = s.call(t, http.MethodPost, base+"/start", host.Token, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("start: %d", resp.StatusCode)
	}
	resp, late := s.call(t, http.MethodPost, base+"/start", host.Token, "")
	if resp.StatusCode != http.StatusConflict || late.Code != CodeDraftStarted {
		t.Fatalf("a second start: %d %q", resp.StatusCode, late.Code)
	}
	resp, late = s.call(t, http.MethodPost, base+"/join", "", `{"invite":"`+v.Invite+`","name":"Late"}`)
	if resp.StatusCode != http.StatusConflict || late.Code != CodeDraftStarted {
		t.Fatalf("a join after the start: %d %q", resp.StatusCode, late.Code)
	}

	// Pick two over the wire: two people and two bots, two cards a pick.
	hv := hc.until("the host's pack", func(v draft.View) bool { return len(v.Pack) > 0 })
	gv := gc.until("the guest's pack", func(v draft.View) bool { return len(v.Pack) > 0 })
	if hv.Take != 2 || gv.Take != 2 || hv.PicksInRound != 7 || hv.Invite != "" {
		t.Fatalf("take %d and %d, %d picks a round, invite %q", hv.Take, gv.Take, hv.PicksInRound, hv.Invite)
	}
	// One card is not a pick two pick.
	id := hc.send(protocol.KindDraftPick, protocol.DraftPickPayload{Cards: []string{hv.Pack[0].ID}})
	if e := hc.expectError(id); e.Code != protocol.CodeDraftPickRefused {
		t.Fatalf("one card in pick two: %+v", e)
	}
	// Nor is a card from the other seat's pack.
	id = hc.send(protocol.KindDraftPick, protocol.DraftPickPayload{Cards: []string{gv.Pack[0].ID, gv.Pack[1].ID}})
	if e := hc.expectError(id); e.Code != protocol.CodeDraftPickRefused {
		t.Fatalf("picking from another seat's pack: %+v", e)
	}
	id = hc.send(protocol.KindAction, map[string]string{"type": "pass_priority"})
	if e := hc.expectError(id); e.Code != protocol.CodeBadRequest {
		t.Fatalf("a game frame on the draft socket: %+v", e)
	}
	n := 0
	for hv.State == draft.StatePicking {
		hc.pick(hv)
		gc.pick(gv)
		n += 2
		want := n
		next := func(v draft.View) bool {
			return v.State == draft.StateDone || (len(v.Picks) == want && len(v.Pack) > 0)
		}
		hv, gv = hc.until("the host's next pack", next), gc.until("the guest's next pack", next)
	}
	if len(hv.Pool) != 42 || len(gv.Pool) != 42 {
		t.Fatalf("pools of %d and %d, want 42", len(hv.Pool), len(gv.Pool))
	}
}

func TestADraftFillsUp(t *testing.T) {
	s := newDraftStack(t)
	hostTok, _ := s.signIn(t, "Ian")
	host := s.create(t, hostTok, `{"set":"syn","seats":2}`)
	base := "/draft/" + host.Draft.ID.String()
	join := `{"invite":"` + host.Draft.Invite + `","name":"Sam"}`
	if resp, _ := s.call(t, http.MethodPost, base+"/join", "", join); resp.StatusCode != http.StatusOK {
		t.Fatalf("join: %d", resp.StatusCode)
	}
	resp, out := s.call(t, http.MethodPost, base+"/join", "", join)
	if resp.StatusCode != http.StatusConflict || out.Code != CodeDraftFull {
		t.Fatalf("a third person in a draft of two: %d %q", resp.StatusCode, out.Code)
	}
}

// A visitor's draft has no invite: nobody else can get in.
func TestAVisitorsDraftCannotBeJoined(t *testing.T) {
	s := newDraftStack(t)
	out := s.create(t, "", `{"set":"syn"}`)
	base := "/draft/" + out.Draft.ID.String()
	for _, body := range []string{`{"invite":""}`, `{"invite":"anything"}`} {
		if resp, _ := s.call(t, http.MethodPost, base+"/join", "", body); resp.StatusCode != http.StatusNotFound {
			t.Fatalf("join %s of a solo draft: %d, want 404", body, resp.StatusCode)
		}
	}
	// Its own seat may read the preview without one.
	if resp, _ := s.call(t, http.MethodGet, base, out.Token, ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("preview with the seat's own token: %d", resp.StatusCode)
	}
}

// A signed-in person holds one unfinished draft: a new one closes the
// old.
func TestANewDraftReplacesTheCreatorsOld(t *testing.T) {
	s := newDraftStack(t)
	tok, _ := s.signIn(t, "Ian")
	first := s.create(t, tok, `{"set":"syn"}`)
	conn := s.dial(t, first.Draft.ID, first.Token)
	conn.next()
	second := s.create(t, tok, `{"set":"syn"}`)
	if s.lobby.DraftCount() != 1 || s.lobby.DraftRoom(first.Draft.ID) != nil || s.lobby.DraftRoom(second.Draft.ID) == nil {
		t.Fatalf("%d drafts after the second create", s.lobby.DraftCount())
	}
	_ = conn.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, _, err := conn.conn.ReadMessage()
	ce, ok := err.(*websocket.CloseError)
	if !ok || ce.Code != websocket.CloseNormalClosure || ce.Text != draft.ClosedReason {
		t.Fatalf("the old draft's socket: %v, want a normal closure saying %q", err, draft.ClosedReason)
	}
}

// --- the seat token ---------------------------------------------------------------

// A draft seat's token opens that draft and nothing else, and no other
// session opens a draft.
func TestADraftSeatTokenIsGoodForItsDraftOnly(t *testing.T) {
	s := newDraftStack(t)
	a, b := s.create(t, "", `{"set":"syn"}`), s.create(t, "", `{"set":"syn"}`)
	// The lobby's "any session" routes refuse it.
	for _, path := range []string{"/games", "/me", "/decks", "/bot/options"} {
		resp := doGet(t, s.srv, path, a.Token)
		resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("GET %s with a draft seat's token: %d, want 403", path, resp.StatusCode)
		}
	}
	// So does the game socket.
	if got := s.dialStatus(t, "/ws?token="+a.Token); got != http.StatusForbidden {
		t.Errorf("the game socket with a draft seat's token: %d, want 403", got)
	}
	// Another draft's token is refused on this one.
	base := "/draft/" + a.Draft.ID.String()
	if resp, _ := s.call(t, http.MethodPost, base+"/start", b.Token, ""); resp.StatusCode != http.StatusForbidden {
		t.Errorf("start with another draft's token: %d, want 403", resp.StatusCode)
	}
	if got := s.dialStatus(t, base+"/ws?token="+b.Token); got != http.StatusForbidden {
		t.Errorf("the socket with another draft's token: %d, want 403", got)
	}
	// No token, a bad one, and a sign-in that is not a seat.
	if got := s.dialStatus(t, base+"/ws"); got != http.StatusUnauthorized {
		t.Errorf("the socket with no token: %d, want 401", got)
	}
	if got := s.dialStatus(t, base+"/ws?token=nonsense"); got != http.StatusUnauthorized {
		t.Errorf("the socket with a bad token: %d, want 401", got)
	}
	signedIn, _ := s.signIn(t, "Ian")
	if got := s.dialStatus(t, base+"/ws?token="+signedIn); got != http.StatusForbidden {
		t.Errorf("the socket with a sign-in that holds no seat: %d, want 403", got)
	}
	if resp, _ := s.call(t, http.MethodPost, base+"/leave", "", ""); resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("leave with no token: %d, want 401", resp.StatusCode)
	}
	if got := s.dialStatus(t, "/draft/"+uuid.NewString()+"/ws?token="+a.Token); got != http.StatusNotFound && got != http.StatusForbidden {
		t.Errorf("the socket of no draft: %d", got)
	}
}

// --- leaving and coming back -------------------------------------------------------

// A person who drops their socket mid-draft comes back with the same
// token to their seat, their picks and the pack they hold.
func TestReconnectingToADraft(t *testing.T) {
	s := newDraftStack(t)
	out := s.create(t, "", `{"set":"syn","seats":2,"pick_seconds":300}`)
	id := out.Draft.ID
	if resp, _ := s.call(t, http.MethodPost, "/draft/"+id.String()+"/start", out.Token, ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("start: %d", resp.StatusCode)
	}
	conn := s.dial(t, id, out.Token)
	v := conn.until("a pack", func(v draft.View) bool { return len(v.Pack) > 0 })
	for i := 1; i <= 3; i++ {
		conn.pick(v)
		want := i
		v = conn.until("the next pack", func(v draft.View) bool { return len(v.Picks) == want && len(v.Pack) > 0 })
	}
	mine := []string{v.Picks[0].ID, v.Picks[1].ID, v.Picks[2].ID}
	holding := v.Pack[0].ID
	_ = conn.conn.Close()

	back := s.dial(t, id, out.Token)
	got := back.next()
	if got.You != 0 || len(got.Picks) != 3 || got.Picks[0].ID != mine[0] || got.Picks[1].ID != mine[1] || got.Picks[2].ID != mine[2] {
		t.Fatalf("picks on return: %+v", got.Picks)
	}
	if len(got.Pack) == 0 || got.Pack[0].ID != holding || got.Round != 1 || got.Pick != 4 {
		t.Fatalf("on return: round %d pick %d, pack of %d", got.Round, got.Pick, len(got.Pack))
	}
	// And the draft goes on.
	back.pick(got)
	back.until("the pick after returning", func(v draft.View) bool { return len(v.Picks) == 4 })
}

func TestLeavingADraft(t *testing.T) {
	s := newDraftStack(t)
	hostTok, _ := s.signIn(t, "Ian")
	host := s.create(t, hostTok, `{"set":"syn","seats":4}`)
	base := "/draft/" + host.Draft.ID.String()
	_, guest := s.call(t, http.MethodPost, base+"/join", "", `{"invite":"`+host.Draft.Invite+`","name":"Sam"}`)
	gc := s.dial(t, host.Draft.ID, guest.Token)
	gc.next()

	// The host leaves the lobby: the guest hosts, and sees the invite.
	if resp, _ := s.call(t, http.MethodPost, base+"/leave", host.Token, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("leave: %d", resp.StatusCode)
	}
	v := gc.until("the guest hosting", func(v draft.View) bool { return len(v.Players) == 1 && v.Players[0].Host })
	if v.You != 0 || v.Invite != host.Draft.Invite {
		t.Fatalf("the new host's view: you %d invite %q", v.You, v.Invite)
	}
	// The leaver's token is spent.
	if resp, _ := s.call(t, http.MethodPost, base+"/start", host.Token, ""); resp.StatusCode != http.StatusForbidden {
		t.Fatalf("start with a seat that left: %d, want 403", resp.StatusCode)
	}
	if got := s.dialStatus(t, base+"/ws?token="+host.Token); got != http.StatusForbidden {
		t.Fatalf("the socket with a seat that left: %d, want 403", got)
	}
	// The last person leaving closes the draft.
	if resp, _ := s.call(t, http.MethodPost, base+"/leave", guest.Token, ""); resp.StatusCode != http.StatusNoContent {
		t.Fatalf("leave: %d", resp.StatusCode)
	}
	if s.lobby.DraftCount() != 0 {
		t.Fatal("a draft with nobody in it stayed open")
	}
	if resp, _ := s.call(t, http.MethodPost, base+"/leave", guest.Token, ""); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("leave of a closed draft: %d, want 404", resp.StatusCode)
	}
}

// --- hidden information over the wire -------------------------------------------

// TestNoDraftFrameCarriesAnotherSeatsCards walks every byte every seat
// is sent through a whole draft of three people and a bot, and fails on
// a card a seat may not see (CR 905.1c: "a player can look only at
// cards in the booster pack they are currently drafting from, cards
// they have already drafted"). The HTTP answers are walked too.
//
// What a seat may see is taken from the frame itself in one place only:
// the pack it is shown first, which is its own by construction (the
// seed fixes the packs, and the server's own reading of them is checked
// against it). Every later card must come from a pack that was passed
// to that seat, which the test tracks from the server's side.
func TestNoDraftFrameCarriesAnotherSeatsCards(t *testing.T) {
	s := newDraftStack(t)
	hostTok, _ := s.signIn(t, "Ian")
	host := s.create(t, hostTok, `{"set":"syn","seats":4,"pick_seconds":300}`)
	id := host.Draft.ID
	base := "/draft/" + id.String()
	tokens := []string{host.Token}
	httpBodies := [][][]byte{{mustJSON(t, host)}}
	for _, name := range []string{"Sam", "Alex"} {
		_, j := s.call(t, http.MethodPost, base+"/join", "", `{"invite":"`+host.Draft.Invite+`","name":"`+name+`"}`)
		tokens = append(tokens, j.Token)
		httpBodies = append(httpBodies, [][]byte{mustJSON(t, j)})
	}
	conns := make([]*draftConn, 3)
	for i, tok := range tokens {
		conns[i] = s.dial(t, id, tok)
		conns[i].next()
	}
	_, started := s.call(t, http.MethodPost, base+"/start", host.Token, "")
	httpBodies[0] = append(httpBodies[0], mustJSON(t, started))

	room := s.lobby.DraftRoom(id)
	// seatIDs by seat number, and every card of the draft, from the
	// server's side.
	var seatIDs []uuid.UUID
	room.With(func(d *draft.Draft) {
		for _, seat := range d.Seats() {
			seatIDs = append(seatIDs, seat.ID)
		}
	})
	type card struct{ id, scry, name string }
	all := map[string]card{}
	// allowed[i] is every card seat i has held or drafted so far.
	allowed := make([]map[string]bool, 3)
	for i := range allowed {
		allowed[i] = map[string]bool{}
	}
	observe := func() {
		room.With(func(d *draft.Draft) {
			for n, sid := range seatIDs {
				held := d.HeldPack(sid)
				// Held and drafted: a bot picks the moment it is handed a
				// pack, so its picks are never seen in its hand.
				for _, c := range append(held, d.Pool(sid)...) {
					all[c.ID] = card{c.ID, c.ScryfallID.String(), c.Name}
				}
				if n < 3 {
					for _, c := range held {
						allowed[n][c.ID] = true
					}
					for _, c := range d.Pool(sid) {
						allowed[n][c.ID] = true
					}
				}
			}
		})
	}
	observe()

	views := make([]draft.View, 3)
	for i, c := range conns {
		views[i] = c.until("a pack", func(v draft.View) bool { return len(v.Pack) > 0 })
	}
	n := 0
	for views[0].State == draft.StatePicking {
		for i, c := range conns {
			c.pick(views[i])
		}
		n++
		want := n
		for i, c := range conns {
			views[i] = c.until("the next pack", func(v draft.View) bool {
				return v.State == draft.StateDone || (len(v.Picks) == want && len(v.Pack) > 0)
			})
		}
		observe()
	}
	if len(all) != 4*3*14 {
		t.Fatalf("the walk saw %d cards, want %d", len(all), 4*3*14)
	}

	frames := 0
	for i, c := range conns {
		sent := append(httpBodies[i], c.raw...)
		for _, raw := range sent {
			frames++
			text := string(raw)
			for _, k := range all {
				if allowed[i][k.id] {
					continue
				}
				// A card ID is "c41": match it as a whole JSON string.
				if strings.Contains(text, `"`+k.id+`"`) {
					t.Fatalf("seat %d was sent card %s (%s), which it never held or drafted", i, k.id, k.name)
				}
				if strings.Contains(text, k.scry) && !seenAs(all, allowed[i], func(o card) string { return o.scry }, k.scry) {
					t.Fatalf("seat %d was sent the printing of %s (%s)", i, k.name, k.id)
				}
				if strings.Contains(text, `"`+k.name+`"`) && !seenAs(all, allowed[i], func(o card) string { return o.name }, k.name) {
					t.Fatalf("seat %d was sent the name %s (%s)", i, k.name, k.id)
				}
			}
			if strings.Contains(text, host.Draft.Invite) && (i != 0 || strings.Contains(text, `"state":"picking"`)) {
				t.Fatalf("seat %d was sent the invite outside the host's lobby view", i)
			}
			for _, sid := range seatIDs {
				if strings.Contains(text, sid.String()) {
					t.Fatalf("seat %d was sent a seat id", i)
				}
			}
		}
	}
	if frames < 150 {
		t.Fatalf("only %d messages were walked", frames)
	}
}

// seenAs reports whether some card the seat may see has this value: two
// copies of one card share a name and a printing.
func seenAs[T any](all map[string]T, allowed map[string]bool, key func(T) string, want string) bool {
	for id, c := range all {
		if allowed[id] && key(c) == want {
			return true
		}
	}
	return false
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// --- limits ---------------------------------------------------------------------

// A draft counts against the load cap as one table while people are
// picking, a person at a time, its bots as nothing, and as nothing once
// it is done (ADR 0222 §9, ADR 0148).
func TestADraftCountsAsOneTableAgainstTheCap(t *testing.T) {
	s := newDraftStack(t)
	lim := capacity.Limits{MaxLoad: 100, TableWeight: 2, HumanSeatWeight: 1, BotSeatWeight: 2}
	s.lobby.SetCapacity(lim, nil)
	if got := s.lobby.CapacityUsage().Load; got != 0 {
		t.Fatalf("load %d with nothing open", got)
	}
	hostTok, _ := s.signIn(t, "Ian")
	host := s.create(t, hostTok, `{"set":"syn","seats":8,"pick_seconds":300}`)
	base := "/draft/" + host.Draft.ID.String()
	if got, want := s.lobby.CapacityUsage().Load, lim.TableCost(1, 0); got != want {
		t.Fatalf("load %d with one draft and one person, want %d (a table and a human seat)", got, want)
	}
	s.call(t, http.MethodPost, base+"/join", "", `{"invite":"`+host.Draft.Invite+`","name":"Sam"}`)
	if got, want := s.lobby.CapacityUsage().Load, lim.TableCost(2, 0); got != want {
		t.Fatalf("load %d with two people, want %d", got, want)
	}
	// Started: six bots fill the seats and add nothing.
	s.call(t, http.MethodPost, base+"/start", host.Token, "")
	if got, want := s.lobby.CapacityUsage().Load, lim.TableCost(2, 0); got != want {
		t.Fatalf("load %d while picking with six bots, want %d: a draft bot runs no game", got, want)
	}
	// A table beside it costs what a table costs.
	if _, err := s.lobby.Create("Commander night"); err != nil {
		t.Fatal(err)
	}
	if got, want := s.lobby.CapacityUsage().Load, lim.TableCost(2, 0)+lim.TableCost(0, 0); got != want {
		t.Fatalf("load %d with a draft and a table, want %d", got, want)
	}
	// The picks are done and decks are being built: still one table and
	// its people. What the rounds cost is
	// TestADraftsMatchesCountAsTablesAgainstTheCap (#5440).
	room := s.lobby.DraftRoom(host.Draft.ID)
	room.With(func(d *draft.Draft) {
		for d.State() == draft.StatePicking {
			d.Expire(time.Now())
		}
	})
	if got, want := s.lobby.CapacityUsage().Load, lim.TableCost(2, 0)+lim.TableCost(0, 0); got != want {
		t.Fatalf("load %d with a draft building its decks and a table, want %d", got, want)
	}
}

func TestAFullServerRefusesADraft(t *testing.T) {
	s := newDraftStack(t)
	// Room for one draft with one person, and no more.
	lim := capacity.Limits{MaxLoad: 3, TableWeight: 2, HumanSeatWeight: 1, BotSeatWeight: 2}
	s.lobby.SetCapacity(lim, nil)
	hostTok, _ := s.signIn(t, "Ian")
	host := s.create(t, hostTok, `{"set":"syn"}`)
	resp, out := s.call(t, http.MethodPost, "/draft", "", `{"set":"syn"}`)
	if resp.StatusCode != http.StatusServiceUnavailable || out.Code != capacity.CodeServerFull {
		t.Fatalf("a draft past the cap: %d %q, want 503 %s", resp.StatusCode, out.Code, capacity.CodeServerFull)
	}
	if resp.Header.Get("Retry-After") == "" {
		t.Error("no Retry-After on the refusal")
	}
	if s.lobby.DraftCount() != 1 {
		t.Fatalf("%d drafts; the refused one must cost nothing", s.lobby.DraftCount())
	}
	// A person joining costs a seat, which does not fit either.
	resp, out = s.call(t, http.MethodPost, "/draft/"+host.Draft.ID.String()+"/join", "", `{"invite":"`+host.Draft.Invite+`","name":"Sam"}`)
	if resp.StatusCode != http.StatusServiceUnavailable || out.Code != capacity.CodeServerFull {
		t.Fatalf("a join past the cap: %d %q", resp.StatusCode, out.Code)
	}
	// And a table is refused by the draft's weight, as a draft is by a
	// table's.
	if _, err := s.lobby.Create("No room"); err != ErrServerFull {
		t.Fatalf("a table past the cap: %v, want ErrServerFull", err)
	}
	// The creator's own new draft replaces the old, so it fits.
	s.create(t, hostTok, `{"set":"syn"}`)
}

func TestAVisitorsDraftsAreRateLimitedPerAddress(t *testing.T) {
	s := newDraftStackLimited(t)
	post := func() int {
		resp, _ := s.call(t, http.MethodPost, "/draft", "", `{"set":"syn"}`)
		return resp.StatusCode
	}
	for i := range 3 {
		if got := post(); got != http.StatusCreated {
			t.Fatalf("call %d: %d, want 201", i+1, got)
		}
	}
	if got := post(); got != http.StatusTooManyRequests {
		t.Fatalf("the fourth draft in the window: %d, want 429", got)
	}
}

// Idle drafts close the way idle tables do (ADR 0211).
func TestIdleDraftsClose(t *testing.T) {
	s := newDraftStack(t)
	lim := capacity.Defaults()
	s.lobby.SetCapacity(lim, nil)
	now := time.Now()
	s.lobby.SetIdleClock(func() time.Time { return now })

	// A signed-in host's lobby: a visitor's has a shorter window of its
	// own (TestAnAbandonedVisitorsDraftClosesSoon).
	hostTok, _ := s.signIn(t, "Ian")
	lobbyDraft := s.create(t, hostTok, `{"set":"syn"}`)
	picking := s.create(t, "", `{"set":"syn"}`)
	watched := s.create(t, "", `{"set":"syn","pick_seconds":300}`)
	for _, d := range []draftAnswer{picking, watched} {
		s.call(t, http.MethodPost, "/draft/"+d.Draft.ID.String()+"/start", d.Token, "")
	}
	conn := s.dial(t, watched.Draft.ID, watched.Token)
	conn.next()

	s.lobby.TickIdle()
	if s.lobby.DraftCount() != 3 {
		t.Fatalf("%d drafts after a sweep with nothing idle", s.lobby.DraftCount())
	}
	// Past the bots window: the picking draft nobody is connected to
	// closes; the one being watched and the one in its lobby stay.
	now = now.Add(lim.IdleBotsAfter + time.Minute)
	s.lobby.TickIdle()
	if s.lobby.DraftRoom(picking.Draft.ID) != nil {
		t.Error("a picking draft with nobody connected stayed open past the idle window")
	}
	if s.lobby.DraftRoom(watched.Draft.ID) == nil || s.lobby.DraftRoom(lobbyDraft.Draft.ID) == nil {
		t.Fatal("a watched draft or a fresh lobby was closed early")
	}
	// Past the lobby window: the lobby draft closes.
	now = now.Add(lim.IdleLobbyAfter)
	s.lobby.TickIdle()
	if s.lobby.DraftRoom(lobbyDraft.Draft.ID) != nil {
		t.Error("an idle draft lobby stayed open")
	}
	// Past the game window: even a watched draft nobody has touched.
	now = now.Add(lim.IdleGameAfter)
	s.lobby.TickIdle()
	if s.lobby.DraftCount() != 0 {
		t.Errorf("%d drafts left after every window", s.lobby.DraftCount())
	}
	_ = conn.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		_, _, err := conn.conn.ReadMessage()
		if err == nil {
			continue
		}
		if ce, ok := err.(*websocket.CloseError); !ok || ce.Text != draft.ClosedReason {
			t.Errorf("the closed draft's socket: %v", err)
		}
		break
	}
}

// A visitor's draft that was opened and walked away from does not hold
// its place for a day: with its page closed it goes after
// SoloDraftAbandonedAfter. A signed-in host's lobby, which friends may
// be on their way to, keeps the table's day.
func TestAnAbandonedVisitorsDraftClosesSoon(t *testing.T) {
	s := newDraftStack(t)
	s.lobby.SetCapacity(capacity.Defaults(), nil)
	now := time.Now()
	s.lobby.SetIdleClock(func() time.Time { return now })
	hostTok, _ := s.signIn(t, "Ian")
	hosted := s.create(t, hostTok, `{"set":"syn"}`)
	abandoned := s.create(t, "", `{"set":"syn"}`)
	open := s.create(t, "", `{"set":"syn"}`)
	conn := s.dial(t, open.Draft.ID, open.Token)
	conn.next()

	now = now.Add(SoloDraftAbandonedAfter + time.Minute)
	s.lobby.TickIdle()
	if s.lobby.DraftRoom(abandoned.Draft.ID) != nil {
		t.Error("a visitor's draft with its page closed stayed open")
	}
	if s.lobby.DraftRoom(open.Draft.ID) == nil {
		t.Error("a visitor's draft with its page open was closed")
	}
	if s.lobby.DraftRoom(hosted.Draft.ID) == nil {
		t.Error("a signed-in host's lobby was closed on the visitor's window")
	}
}

// A draft socket counts against the connection cap like any other, and
// is not listed as a table socket.
func TestADraftSocketCountsAsAConnection(t *testing.T) {
	s := newDraftStack(t)
	out := s.create(t, "", `{"set":"syn"}`)
	if s.hub.Count() != 0 {
		t.Fatalf("%d connections before the socket", s.hub.Count())
	}
	conn := s.dial(t, out.Draft.ID, out.Token)
	conn.next()
	if s.hub.Count() != 1 {
		t.Fatalf("%d connections with one draft socket, want 1", s.hub.Count())
	}
	if got := s.hub.LiveSockets(); len(got) != 0 {
		t.Fatalf("a draft socket is listed as a table socket: %+v", got)
	}
	// Ping works as on the game socket.
	id := conn.send(protocol.KindPing, protocol.PingPayload{Msg: "hi"})
	for range 20 {
		_ = conn.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		_, raw, err := conn.conn.ReadMessage()
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		var f protocol.Frame
		_ = json.Unmarshal(raw, &f)
		if f.Kind == protocol.KindPong && f.ID == id {
			return
		}
	}
	t.Fatal("no pong")
}
