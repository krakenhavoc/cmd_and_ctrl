package lobby

// agent_seat_test.go: ADR 0122 §7, the agent badge on the lobby side.
// An MCP client declares itself in the join body; the seat is a guest,
// badged for the whole table, never the host, never linked to Discord,
// and the badge never comes off.

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestNormalizeAgentClient(t *testing.T) {
	cases := map[string]string{
		"claude-code":                           "claude-code",
		"Claude Code":                           "claude-code",
		"codex_cli.v2":                          "codex_cli.v2",
		"":                                      "unknown",
		"   ":                                   "unknown",
		"!!!":                                   "unknown",
		"Ünïcode":                               "n-code",
		"a-very-long-client-name-that-goes-on":  "a-very-long-client-name-that-goe",
		"-trimmed-":                             "trimmed",
		"0123456789012345678901234567890-xyz":   "0123456789012345678901234567890",
		"MixedCase/With/Slashes":                "mixedcase-with-slashes",
		strings.Repeat("x", 1000):               strings.Repeat("x", 32),
		"<script>alert(1)</script>":             "script-alert-1---script",
		"claude-code‮noitcejni":                 "claude-code-noitcejni",
		"ok.name_with-every.allowed_char-01234": "ok.name_with-every.allowed_char", // cut at 32, then the trailing "-" trimmed
	}
	for in, want := range cases {
		got := NormalizeAgentClient(in)
		if got != want {
			t.Errorf("NormalizeAgentClient(%q) = %q, want %q", in, got, want)
		}
		if len(got) > maxAgentClientLen {
			t.Errorf("NormalizeAgentClient(%q) = %q, longer than %d", in, got, maxAgentClientLen)
		}
	}
}

// agentJoinBody is a join body with the agent declaration, as the MCP
// seat sends it.
func agentJoinBody(invite, name, client string) map[string]any {
	return map[string]any{"invite_token": invite, "name": name, "agent": map[string]string{"client": client}}
}

// seatNamed finds a seat by name in a meta.
func seatNamed(t *testing.T, m GameMeta, name string) SeatInfo {
	t.Helper()
	for _, s := range m.Players {
		if s.Name == name {
			return s
		}
	}
	t.Fatalf("no seat %q in %+v", name, m.Players)
	return SeatInfo{}
}

// assertAgentEverywhere checks that playerID is badged as an agent of
// client, and no other seat is, in the lobby meta and in the room's
// game view.
func assertAgentEverywhere(t *testing.T, l *Lobby, gameID, playerID uuid.UUID, client string) {
	t.Helper()
	m := mustGet(t, l, gameID)
	for _, s := range m.Players {
		want := s.PlayerID == playerID
		if s.IsAgent != want || (want && s.AgentClient != client) || (!want && s.AgentClient != "") {
			t.Errorf("meta seat %s: is_agent %v agent_client %q, want %v / %q", s.Name, s.IsAgent, s.AgentClient, want, client)
		}
		if want && (s.IsHost || s.IsBot || s.UserID != "" || s.DiscordID != "") {
			t.Errorf("agent seat %+v is a host, a bot or a person", s)
		}
	}
	view, _, err := l.RoomOf(gameID).Snapshot()
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	for _, s := range view.Seats {
		want := s.ID == playerID.String()
		if s.IsAgent != want || (want && s.AgentClient != client) {
			t.Errorf("view seat %s: is_agent %v agent_client %q, want %v / %q", s.Name, s.IsAgent, s.AgentClient, want, client)
		}
	}
}

func TestAgentJoinsOnBothRoutes(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	srv, l, a := newTestHTTPStack(t)
	admin, meta := adminGame(t, srv, "FNM")

	// The invite-link route.
	sess := decodeSession(t, postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		agentJoinBody(meta.InviteToken, "Claude", "Claude Code")), http.StatusOK)
	if sess.Principal.Role != auth.RolePlayer || sess.Principal.UserID != uuid.Nil || sess.Principal.DiscordID != "" {
		t.Errorf("agent seat session = %+v, want a guest player", sess.Principal)
	}
	if p := mustValidate(t, a, sess.Token); p.PlayerID != sess.PlayerID {
		t.Errorf("token binds %v, want %v", p.PlayerID, sess.PlayerID)
	}
	if sess.Game == nil {
		t.Fatal("join response carries no game")
	}
	if s := seatNamed(t, *sess.Game, "Claude"); !s.IsAgent || s.AgentClient != "claude-code" {
		t.Errorf("join response seat = %+v, want the badge", s)
	}
	if sess.Game.InviteToken != "" || sess.Game.SpectatorInvite != "" {
		t.Error("agent join response leaked an invite")
	}

	// The code route, with an empty declaration: still an agent.
	byCode := decodeSession(t, postJSON(t, srv, "/join", "",
		map[string]any{"invite_token": meta.InviteToken, "name": "Codex", "agent": map[string]any{}}), http.StatusOK)

	// A plain join beside them is no agent.
	human := decodeSession(t, postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		map[string]any{"invite_token": meta.InviteToken, "name": "Alice"}), http.StatusOK)

	m := mustGet(t, l, meta.ID)
	if s := seatNamed(t, m, "Codex"); !s.IsAgent || s.AgentClient != "unknown" {
		t.Errorf("POST /join agent seat = %+v, want is_agent and \"unknown\"", s)
	}
	if s := seatNamed(t, m, "Alice"); s.IsAgent || s.AgentClient != "" {
		t.Errorf("human seat = %+v, badged", s)
	}
	// Two agents joined first; the first human hosts.
	if m.HostPlayerID != human.PlayerID {
		t.Errorf("host = %v, want the first human %v", m.HostPlayerID, human.PlayerID)
	}

	// Every place a seat list goes: GET /games, GET /games/{id}, the
	// invite preview, and the game view.
	type seatWire struct {
		Name        string `json:"name"`
		IsAgent     bool   `json:"is_agent"`
		AgentClient string `json:"agent_client"`
	}
	check := func(where string, seats []seatWire) {
		t.Helper()
		want := map[string]string{"Claude": "claude-code", "Codex": "unknown", "Alice": ""}
		if len(seats) != len(want) {
			t.Fatalf("%s: %d seats", where, len(seats))
		}
		for _, s := range seats {
			if s.IsAgent != (want[s.Name] != "") || s.AgentClient != want[s.Name] {
				t.Errorf("%s: %s is_agent %v agent_client %q", where, s.Name, s.IsAgent, s.AgentClient)
			}
		}
	}
	decode := func(resp *http.Response, into any) {
		t.Helper()
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			t.Fatalf("status %d: %s", resp.StatusCode, b)
		}
		if err := json.NewDecoder(resp.Body).Decode(into); err != nil {
			t.Fatalf("decode: %v", err)
		}
	}
	var list struct {
		Games []struct {
			Players []seatWire `json:"players"`
		} `json:"games"`
	}
	decode(doGet(t, srv, "/games", admin.Token), &list)
	if len(list.Games) != 1 {
		t.Fatalf("GET /games: %d games", len(list.Games))
	}
	check("GET /games", list.Games[0].Players)

	var one struct {
		Players []seatWire `json:"players"`
	}
	decode(doGet(t, srv, "/games/"+meta.ID.String(), human.Token), &one)
	check("GET /games/{id} as a seat", one.Players)

	var preview struct {
		Game struct {
			Players []seatWire `json:"players"`
		} `json:"game"`
	}
	decode(doGet(t, srv, "/games/"+meta.ID.String()+"/preview?t="+meta.InviteToken, ""), &preview)
	check("the invite preview", preview.Game.Players)

	view, _, err := l.RoomOf(meta.ID).Snapshot()
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	wantView := map[string]string{sess.PlayerID.String(): "claude-code", byCode.PlayerID.String(): "unknown", human.PlayerID.String(): ""}
	for _, s := range view.Seats {
		if s.IsAgent != (wantView[s.ID] != "") || s.AgentClient != wantView[s.ID] {
			t.Errorf("game view: %s is_agent %v agent_client %q", s.Name, s.IsAgent, s.AgentClient)
		}
	}
}

func TestAgentJoinRefusesASignedInSession(t *testing.T) {
	s := newMyGamesStack(t)
	meta, _ := s.lobby.Create("FNM")
	other, _ := s.lobby.Create("Other")
	idTok := identityTokenFromCallback(t, s.srv, s.state)

	// A signed-in person at another table: a seat session with a user.
	seatSess := decodeSession(t, postJSON(t, s.srv, "/games/"+other.ID.String()+"/join", idTok,
		map[string]string{"invite_token": other.InviteToken}), http.StatusOK)
	if seatSess.Principal.UserID == uuid.Nil {
		t.Fatal("seat session carries no user")
	}

	for _, tc := range []struct{ name, tok string }{
		{"identified", idTok},
		{"signed-in seat", seatSess.Token},
	} {
		for _, path := range []string{"/games/" + meta.ID.String() + "/join", "/join"} {
			resp := postJSON(t, s.srv, path, tc.tok, agentJoinBody(meta.InviteToken, "Claude", "claude-code"))
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest || !strings.Contains(string(body), "an agent seat joins as a guest") {
				t.Errorf("%s session, %s: %d %s, want 400 \"an agent seat joins as a guest\"", tc.name, path, resp.StatusCode, body)
			}
		}
	}
	if n := len(mustGet(t, s.lobby, meta.ID).Players); n != 0 {
		t.Errorf("a refused agent join left %d seat(s)", n)
	}

	// The lobby holds the rule too, for any caller that is not HTTP.
	if _, _, err := s.lobby.join(meta.ID, meta.InviteToken, "x", DiscordIdentity{ID: "1", Username: "u"}, uuid.Nil, &AgentDecl{Client: "c"}); !errors.Is(err, ErrAgentSignedIn) {
		t.Errorf("join with an identity and an agent: %v, want ErrAgentSignedIn", err)
	}
	if _, _, err := s.lobby.join(meta.ID, meta.InviteToken, "x", DiscordIdentity{}, uuid.New(), &AgentDecl{Client: "c"}); !errors.Is(err, ErrAgentSignedIn) {
		t.Errorf("join with a user and an agent: %v, want ErrAgentSignedIn", err)
	}

	// With no session, the same body joins as a guest agent.
	guest := decodeSession(t, postJSON(t, s.srv, "/join", "", agentJoinBody(meta.InviteToken, "Claude", "claude-code")), http.StatusOK)
	if guest.Principal.UserID != uuid.Nil {
		t.Errorf("guest agent seat carries user %v", guest.Principal.UserID)
	}
	if uid, pending := seatRow(t, s, meta.ID, guest.PlayerID); uid.Valid || pending.Valid {
		t.Errorf("agent seat row: user_id %v pending %v, want both NULL", uid, pending)
	}

	// A guest's seat session on POST /join is the same 409 as before:
	// it already belongs to a table.
	resp := postJSON(t, s.srv, "/join", guest.Token, agentJoinBody(meta.InviteToken, "Again", "claude-code"))
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("agent join on POST /join with a guest seat session: %d, want 409", resp.StatusCode)
	}
}

func TestAgentSeatNeverHosts(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	srv, l, _ := newTestHTTPStack(t)
	admin, meta := adminGame(t, srv, "FNM")

	_, agent, err := l.JoinAgent(meta.ID, meta.InviteToken, "Claude", AgentDecl{Client: "claude-code"})
	if err != nil {
		t.Fatalf("JoinAgent: %v", err)
	}
	// Alone at the table, the agent does not take it.
	assertHost(t, l, meta.ID, uuid.Nil)

	_, alice, _ := l.Join(meta.ID, meta.InviteToken, "Alice")
	assertHost(t, l, meta.ID, alice)

	if _, err := l.TransferHost(meta.ID, agent); !errors.Is(err, ErrHostIneligible) {
		t.Errorf("TransferHost to the agent: %v, want ErrHostIneligible", err)
	}
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/host", admin.Token, map[string]string{"player_id": agent.String()})
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("POST /host naming the agent: %d, want 422", resp.StatusCode)
	}
	assertHost(t, l, meta.ID, alice)
}

func TestDiscordLinkRefusesAnAgentSeat(t *testing.T) {
	s := newMyGamesStack(t)
	meta, _ := s.lobby.Create("FNM")
	sess := decodeSession(t, postJSON(t, s.srv, "/games/"+meta.ID.String()+"/join", "",
		agentJoinBody(meta.InviteToken, "Claude", "claude-code")), http.StatusOK)

	resp := getWithCookie(t, s.srv, s.srv.URL, "/auth/discord/link?game="+meta.ID.String(), sess.Token)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("GET /auth/discord/link from an agent seat: %d %s, want 422", resp.StatusCode, body)
	}

	// The callback's half refuses too, should a link ever be started.
	if _, _, err := s.lobby.LinkSeat(meta.ID, sess.PlayerID, DiscordIdentity{ID: "d-1", Username: "u"}, uuid.Nil); !errors.Is(err, ErrSeatIsAgent) {
		t.Errorf("LinkSeat on an agent seat: %v, want ErrSeatIsAgent", err)
	}
	assertAgentEverywhere(t, s.lobby, meta.ID, sess.PlayerID, "claude-code")
}

// TestAgentBadgeCannotBeCleared walks an agent seat through every path
// that changes a seat or could rebuild one — the host hand-over and
// Discord link (both refused), a settings change, an admin's reclaim
// ticket and its redemption, the start, play, and a restart — and
// checks the badge after each. game.TestAgentBadgeHasNoClearingWriter
// is the static half: no code in the server writes it off.
func TestAgentBadgeCannotBeCleared(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, err := l.Create("FNM")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatalf("Join: %v", err)
	}
	_, agent, err := l.JoinAgent(meta.ID, meta.InviteToken, "Claude", AgentDecl{Client: "codex"})
	if err != nil {
		t.Fatalf("JoinAgent: %v", err)
	}
	step := func(label string) {
		t.Helper()
		assertAgentEverywhere(t, l, meta.ID, agent, "codex")
		if t.Failed() {
			t.Fatalf("badge lost after: %s", label)
		}
	}
	step("join")

	if _, err := l.TransferHost(meta.ID, agent); !errors.Is(err, ErrHostIneligible) {
		t.Errorf("TransferHost: %v", err)
	}
	step("a refused host hand-over")

	if _, _, err := l.LinkSeat(meta.ID, agent, DiscordIdentity{ID: "d-9", Username: "x"}, uuid.Nil); !errors.Is(err, ErrSeatIsAgent) {
		t.Errorf("LinkSeat: %v", err)
	}
	step("a refused Discord link")

	// A second declaration on the engine seat keeps the first.
	if err := l.RoomOf(meta.ID).Game.SetAgent(agent, "claude-code"); err != nil {
		t.Errorf("SetAgent again: %v", err)
	}
	step("a second declaration")

	limit := 3
	if _, err := l.UpdateSettings(meta.ID, alice, game.SettingsPatch{UndoLimit: &limit}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	step("a settings change")

	for _, id := range []uuid.UUID{alice, agent} {
		if _, err := l.SetDeck(meta.ID, id, "deck", botDeck(20)); err != nil {
			t.Fatalf("SetDeck: %v", err)
		}
	}
	step("deck upload")
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	step("the start")

	ticket, err := l.MintReclaim(meta.ID, agent)
	if err != nil {
		t.Fatalf("MintReclaim: %v", err)
	}
	if _, seat, err := l.RedeemReclaim(meta.ID, ticket.Token); err != nil || seat.PlayerID != agent || !seat.IsAgent {
		t.Fatalf("RedeemReclaim = %+v, %v; want the agent seat, badged", seat, err)
	}
	step("an admin reclaim")

	// The deploy.
	l2, _ := newDurableLobby(t, dir)
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}
	l = l2
	step("a restart")
	if h := mustGet(t, l, meta.ID).HostPlayerID; h != alice {
		t.Errorf("host after restart = %v, want Alice", h)
	}
}
