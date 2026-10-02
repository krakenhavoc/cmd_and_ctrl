package lobby

// admins_test.go covers ADR 0110 §3 (Delivery PR 3): admins from the
// Discord allowlist, isAdmin and isServerCredential, requireAdmin on
// every admin route, /me's admin bit, WebSocket parity, and the
// owner's requirement of 2026-10-02 that an allowlisted person sits at
// a table as their own Discord identity and is an admin from that seat.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

const (
	listedDiscordID   = "111111111111111111"
	unlistedDiscordID = "222222222222222222"
)

// adminStack is the HTTP and WebSocket stack with an allowlist wired
// into both, exactly as main wires it, and an audit log to read.
type adminStack struct {
	srv    *httptest.Server
	lobby  *Lobby
	auth   auth.Authenticator
	admins *AdminList
	log    *syncBuffer
	cfg    Config
	wsAuth *WSAuthorizer
}

func newAdminStack(t *testing.T, ids ...string) *adminStack {
	t.Helper()
	return newAdminStackWithCards(t, nil, ids...)
}

func newAdminStackWithCards(t *testing.T, idx *cards.Index, ids ...string) *adminStack {
	t.Helper()
	logBuf := &syncBuffer{}
	log := slog.New(slog.NewTextHandler(logBuf, nil))
	mgr := ws.NewRoomManager(slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	l := NewLobby(mgr)
	a := auth.NewMemoryAuthenticator()
	admins := NewAdminList(ids...)
	hub := ws.NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
	hub.SetManager(mgr)
	wsAuth := &WSAuthorizer{Auth: a, Admins: admins, Log: log}
	hub.SetAuthorizer(wsAuth)
	cfg := Config{
		Lobby:      l,
		Auth:       a,
		AdminToken: "shared-admin-token",
		Admins:     admins,
		Log:        log,
		Cards:      idx,
	}
	mux := http.NewServeMux()
	mux.Handle("/", Handler(cfg))
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &adminStack{srv: srv, lobby: l, auth: a, admins: admins, log: logBuf, cfg: cfg, wsAuth: wsAuth}
}

// issue mints p directly; the role and fields are the test's choice.
func (s *adminStack) issue(t *testing.T, p auth.Principal) string {
	t.Helper()
	tok, _, err := s.auth.Issue(context.Background(), p, time.Hour)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	return tok
}

// signedIn mints the identity session a Discord sign-in from the login
// page mints: a user and a Discord identity, no game.
func (s *adminStack) signedIn(t *testing.T, discordID, name string) (string, uuid.UUID) {
	t.Helper()
	user := uuid.New()
	return s.issue(t, auth.Principal{
		Role:              auth.RoleIdentified,
		UserID:            user,
		DiscordID:         discordID,
		DiscordUsername:   strings.ToLower(name),
		DiscordGlobalName: name,
	}), user
}

func (s *adminStack) adminToken(t *testing.T) string {
	t.Helper()
	resp := postJSON(t, s.srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	defer resp.Body.Close()
	var sess sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil || sess.Token == "" {
		t.Fatalf("admin login: %d %v", resp.StatusCode, err)
	}
	return sess.Token
}

func do(t *testing.T, srv *httptest.Server, method, path, token string, body any) *http.Response {
	t.Helper()
	var rd io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		rd = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, srv.URL+path, rd)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func status(t *testing.T, srv *httptest.Server, method, path, token string, body any) int {
	t.Helper()
	resp := do(t, srv, method, path, token, body)
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return resp.StatusCode
}

// --- the predicates -------------------------------------------------

func TestIsAdminTruthTable(t *testing.T) {
	c := Config{Admins: NewAdminList(listedDiscordID)}
	user := uuid.New()
	game := uuid.New()
	cases := []struct {
		name   string
		p      auth.Principal
		admin  bool
		server bool
	}{
		{"shared token", auth.Principal{Role: auth.RoleAdmin, AdminID: uuid.New()}, true, true},
		{"allowlisted identified", auth.Principal{Role: auth.RoleIdentified, UserID: user, DiscordID: listedDiscordID}, true, false},
		{"allowlisted seat", auth.Principal{Role: auth.RolePlayer, UserID: user, GameID: game, PlayerID: uuid.New(), DiscordID: listedDiscordID}, true, false},
		{"allowlisted spectator", auth.Principal{Role: auth.RoleSpectator, UserID: user, GameID: game, DiscordID: listedDiscordID}, true, false},
		{"unlisted signed-in", auth.Principal{Role: auth.RoleIdentified, UserID: user, DiscordID: unlistedDiscordID}, false, false},
		// redeemSeatReclaim copies the seat's DiscordID onto a session
		// with no user: a ticket for an admin's seat is not admin.
		{"reclaim ticket for a listed seat", auth.Principal{Role: auth.RolePlayer, GameID: game, PlayerID: uuid.New(), DiscordID: listedDiscordID}, false, false},
		{"user with no Discord ID", auth.Principal{Role: auth.RoleIdentified, UserID: user}, false, false},
		{"guest seat", auth.Principal{Role: auth.RolePlayer, GameID: game, PlayerID: uuid.New(), Name: "Guest"}, false, false},
		{"empty principal", auth.Principal{}, false, false},
	}
	for _, tc := range cases {
		if got := c.isAdmin(tc.p); got != tc.admin {
			t.Errorf("%s: isAdmin = %v, want %v", tc.name, got, tc.admin)
		}
		if got := isServerCredential(tc.p); got != tc.server {
			t.Errorf("%s: isServerCredential = %v, want %v", tc.name, got, tc.server)
		}
	}

	// No list at all: only the token.
	none := Config{}
	if none.isAdmin(auth.Principal{Role: auth.RoleIdentified, UserID: user, DiscordID: listedDiscordID}) {
		t.Error("a nil allowlist made a signed-in user admin")
	}
	if !none.isAdmin(auth.Principal{Role: auth.RoleAdmin}) {
		t.Error("a nil allowlist stopped the shared token being admin")
	}
}

func TestIsAdminFollowsTheListBetweenRequests(t *testing.T) {
	c := Config{Admins: NewAdminList(listedDiscordID)}
	p := auth.Principal{Role: auth.RoleIdentified, UserID: uuid.New(), DiscordID: listedDiscordID}
	if !c.isAdmin(p) {
		t.Fatal("listed user is not admin")
	}
	c.Admins.Replace(nil)
	if c.isAdmin(p) {
		t.Error("the same principal is still admin after its ID was removed")
	}
	c.Admins.Replace([]string{listedDiscordID})
	if !c.isAdmin(p) {
		t.Error("re-adding the ID did not restore admin")
	}
}

func TestParseAdminList(t *testing.T) {
	l, err := ParseAdminList(" 111111111111111111 ,222222222222222222,, ")
	if err != nil {
		t.Fatalf("ParseAdminList: %v", err)
	}
	if l.Len() != 2 || !l.Has("111111111111111111") || !l.Has("222222222222222222") {
		t.Errorf("parsed %d ids, want both", l.Len())
	}
	if l.Has("") {
		t.Error("the empty string is on the list")
	}
	for _, raw := range []string{"", "   ", ","} {
		l, err := ParseAdminList(raw)
		if err != nil || l.Len() != 0 {
			t.Errorf("ParseAdminList(%q) = %d ids, %v; want an empty list", raw, l.Len(), err)
		}
	}
	for _, raw := range []string{"12345,not-a-snowflake", "1234567890123456789012", "12 34"} {
		_, err := ParseAdminList(raw)
		if err == nil {
			t.Errorf("ParseAdminList(%q) accepted a malformed entry", raw)
			continue
		}
		// The error names the entry's position, never its value: the
		// boot log must not carry the list.
		if strings.Contains(err.Error(), "not-a-snowflake") || strings.Contains(err.Error(), "1234567890123456789012") {
			t.Errorf("ParseAdminList error leaks the value: %v", err)
		}
	}
	var nilList *AdminList
	if nilList.Has(listedDiscordID) || nilList.Len() != 0 {
		t.Error("a nil list is not empty")
	}
}

// --- the guard ------------------------------------------------------

// TestRoleAdminIsComparedOnlyInTheAdminPredicates fails on any use of
// auth.RoleAdmin in the server outside isServerCredential (the one
// comparison) and adminLogin (the one mint), and on any `.Role ==
// "admin"` string comparison. A new admin check has to choose isAdmin
// or isServerCredential (ADR 0110 §3 item 2). internal/auth, which
// defines the role, is exempt.
func TestRoleAdminIsComparedOnlyInTheAdminPredicates(t *testing.T) {
	root := filepath.Join("..", "..")
	allowed := map[string]bool{"isServerCredential": true, "adminLogin": true}
	fset := token.NewFileSet()
	var offences []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case "testdata", "vendor", "node_modules":
				return filepath.SkipDir
			}
			if filepath.ToSlash(path) == filepath.ToSlash(filepath.Join(root, "internal", "auth")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		check := func(fn string, n ast.Node) {
			ast.Inspect(n, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.SelectorExpr:
					if id, ok := x.X.(*ast.Ident); ok && id.Name == "auth" && x.Sel.Name == "RoleAdmin" && !allowed[fn] {
						offences = append(offences, fset.Position(x.Pos()).String()+" (in "+fn+")")
					}
				case *ast.BinaryExpr:
					if x.Op != token.EQL && x.Op != token.NEQ {
						return true
					}
					if isRoleSelector(x.X) && isAdminLiteral(x.Y) || isRoleSelector(x.Y) && isAdminLiteral(x.X) {
						offences = append(offences, fset.Position(x.Pos()).String()+` compares .Role with "admin"`)
					}
				}
				return true
			})
		}
		for _, decl := range f.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok {
				check(fd.Name.Name, fd)
				continue
			}
			check("", decl)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	for _, o := range offences {
		t.Errorf("bare auth.RoleAdmin at %s: use c.isAdmin(p) for an operator check, or isServerCredential(p) for one of the bot's own paths (ADR 0110 §3)", o)
	}
}

func isRoleSelector(e ast.Expr) bool {
	s, ok := e.(*ast.SelectorExpr)
	return ok && s.Sel.Name == "Role"
}

func isAdminLiteral(e ast.Expr) bool {
	b, ok := e.(*ast.BasicLit)
	return ok && b.Kind == token.STRING && b.Value == `"admin"`
}

// --- requireAdmin on every admin route ------------------------------

func TestRequireAdminOnEveryAdminRoute(t *testing.T) {
	s := newAdminStack(t, listedDiscordID)
	token := s.adminToken(t)
	listed, _ := s.signedIn(t, listedDiscordID, "Owner")
	unlisted, _ := s.signedIn(t, unlistedDiscordID, "Stranger")

	meta, err := s.lobby.Create("route table")
	if err != nil {
		t.Fatal(err)
	}
	_, guestPlayer, err := s.lobby.Join(meta.ID, meta.InviteToken, "Guest")
	if err != nil {
		t.Fatal(err)
	}
	guest := s.issue(t, auth.Principal{Role: auth.RolePlayer, GameID: meta.ID, PlayerID: guestPlayer, Name: "Guest"})
	// A reclaim ticket's session for an admin's seat: the seat's Discord
	// ID, no user.
	reclaim := s.issue(t, auth.Principal{Role: auth.RolePlayer, GameID: meta.ID, PlayerID: guestPlayer, DiscordID: listedDiscordID})

	routes := []struct {
		method, path string
		body         any
	}{
		// POST /games is not here: since ADR 0110 PR 7 any signed-in
		// person may create a table (player_tables_test.go).
		{"POST", "/games/" + meta.ID.String() + "/archive", nil},
		{"DELETE", "/games/" + meta.ID.String() + "/archive", nil},
		{"POST", "/games/" + meta.ID.String() + "/seats/" + guestPlayer.String() + "/reclaim", nil},
		{"GET", "/games/" + meta.ID.String() + "/creator?discord_id=" + listedDiscordID, nil},
		{"GET", "/bugreport/abc/replay", nil},
		{"GET", "/bugreport/abc/gamelog", nil},
		{"POST", "/admin/users/" + uuid.New().String() + "/revoke-sessions", nil},
		{"GET", "/games/" + meta.ID.String() + "/bot/stats", nil},
		// Last: it deletes the table the others use.
		{"DELETE", "/games/" + meta.ID.String(), nil},
	}
	refused := []struct {
		name, token string
		want        int
	}{
		{"no credential", "", http.StatusUnauthorized},
		{"guest seat", guest, http.StatusForbidden},
		{"unlisted signed-in", unlisted, http.StatusForbidden},
		{"reclaim ticket with a listed Discord ID", reclaim, http.StatusForbidden},
	}
	for _, rt := range routes {
		for _, who := range refused {
			if got := status(t, s.srv, rt.method, rt.path, who.token, rt.body); got != who.want {
				t.Errorf("%s %s as %s: got %d, want %d", rt.method, rt.path, who.name, got, who.want)
			}
		}
		for _, who := range []struct{ name, token string }{{"allowlisted", listed}, {"shared token", token}} {
			got := status(t, s.srv, rt.method, rt.path, who.token, rt.body)
			if got == http.StatusUnauthorized || got == http.StatusForbidden {
				t.Errorf("%s %s as %s: got %d, want the handler's answer", rt.method, rt.path, who.name, got)
			}
		}
	}

	// The audit line: the action and who, never a token.
	logs := s.log.String()
	if !strings.Contains(logs, "admin action") || !strings.Contains(logs, "admin_user_id=") || !strings.Contains(logs, "admin_id=") {
		t.Errorf("admin routes were not audited with admin_user_id and admin_id:\n%s", logs)
	}
	for _, tok := range []string{token, listed} {
		if strings.Contains(logs, tok) {
			t.Error("the audit log carries a session token")
		}
	}
}

func TestAdminCreatedTableIsAttributedToThePerson(t *testing.T) {
	s := newAdminStack(t, listedDiscordID)
	listed, user := s.signedIn(t, listedDiscordID, "Owner")
	resp := postJSON(t, s.srv, "/games", listed, createGameRequest{Name: "owner's table"})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: %d", resp.StatusCode)
	}
	var meta GameMeta
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		t.Fatal(err)
	}
	if !meta.IsCreator {
		t.Error("an allowlisted creator is not told the table is theirs")
	}
	if got, _ := s.lobby.CreatedBy(meta.ID); got != user.String() {
		t.Errorf("created_by = %q, want the allowlisted user %s", got, user)
	}
}

func TestAdminRouteRefusedOnceTheIDIsRemoved(t *testing.T) {
	s := newAdminStack(t, listedDiscordID)
	listed, _ := s.signedIn(t, listedDiscordID, "Owner")
	meta, err := s.lobby.Create("someone else's table")
	if err != nil {
		t.Fatal(err)
	}
	archive := "/games/" + meta.ID.String() + "/archive"
	if got := status(t, s.srv, "POST", archive, listed, nil); got != http.StatusOK {
		t.Fatalf("listed archive: %d", got)
	}
	s.admins.Replace(nil)
	if got := status(t, s.srv, "DELETE", archive, listed, nil); got != http.StatusForbidden {
		t.Errorf("unarchive after removal, same token: got %d, want 403", got)
	}
	if got := status(t, s.srv, "GET", "/me", listed, nil); got != http.StatusOK {
		t.Errorf("/me after removal: %d (the session itself must survive)", got)
	}
}

// --- the bot-only paths stay on the token ---------------------------

func TestServerCredentialPathsStayTokenOnly(t *testing.T) {
	c := Config{Admins: NewAdminList(listedDiscordID)}
	listed := auth.Principal{Role: auth.RoleIdentified, UserID: uuid.New(), DiscordID: listedDiscordID}
	token := auth.Principal{Role: auth.RoleAdmin, AdminID: uuid.New()}

	t.Run("callerKey", func(t *testing.T) {
		key := func(p auth.Principal) string {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			return callerKey(r.WithContext(auth.WithPrincipal(r.Context(), p)))
		}
		if got := key(listed); got != "user:"+listed.UserID.String() {
			t.Errorf("allowlisted callerKey = %q, want their own user bucket", got)
		}
		if got := key(token); got != "admin" {
			t.Errorf("token callerKey = %q, want the shared admin bucket", got)
		}
	})

	t.Run("practiceOwner", func(t *testing.T) {
		if got := practiceOwner(c, listed); got != "user:"+listed.UserID.String() {
			t.Errorf("allowlisted practiceOwner = %q, want their user", got)
		}
		if got := practiceOwner(c, token); got != "admin" {
			t.Errorf("token practiceOwner = %q, want admin", got)
		}
	})

	t.Run("deck request on a member's behalf", func(t *testing.T) {
		named := &deckRequester{DiscordID: unlistedDiscordID, DisplayName: "Member"}
		_, err := deckRequesterFor(context.Background(), c, listed, named)
		var he *httpStatusError
		if !errors.As(err, &he) || he.code != http.StatusForbidden {
			t.Errorf("an allowlisted person naming a requester: %v, want 403", err)
		}
		who, err := deckRequesterFor(context.Background(), c, token, named)
		if err != nil || who.DiscordID != unlistedDiscordID {
			t.Errorf("the token naming a requester: %+v, %v", who, err)
		}
	})

	t.Run("deck coverage bucket", func(t *testing.T) {
		s := newAdminStack(t, listedDiscordID)
		listedTok, _ := s.signedIn(t, listedDiscordID, "Owner")
		adminTok := s.adminToken(t)
		body := deckCoverageRequest{Text: "1 Sol Ring"}
		// The public bucket is a burst of 3: the fourth call from the
		// allowlisted person is throttled like anyone's.
		var last int
		for i := 0; i < 4; i++ {
			last = status(t, s.srv, "POST", "/deck-coverage", listedTok, body)
		}
		if last != http.StatusTooManyRequests {
			t.Errorf("allowlisted person's 4th coverage check: %d, want 429 (the public bucket)", last)
		}
		for i := 0; i < 4; i++ {
			if got := status(t, s.srv, "POST", "/deck-coverage", adminTok, body); got == http.StatusTooManyRequests {
				t.Errorf("token's coverage check %d throttled; it has the bot's own bucket", i+1)
			}
		}
	})

	t.Run("raw snowflake DM", func(t *testing.T) {
		s := newAdminStack(t, listedDiscordID)
		listedTok, _ := s.signedIn(t, listedDiscordID, "Owner")
		meta, err := s.lobby.Create("dm table")
		if err != nil {
			t.Fatal(err)
		}
		path := "/games/" + meta.ID.String() + "/invites/dm"
		body := dmInviteRequest{DiscordID: unlistedDiscordID}
		resp := do(t, s.srv, "POST", path, listedTok, body)
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden || !strings.Contains(string(raw), "discord_id is admin-only") {
			t.Errorf("allowlisted person DMing a raw snowflake: %d %s, want 403 discord_id is admin-only", resp.StatusCode, raw)
		}
		// The token passes that check and stops at the unconfigured bot.
		if got := status(t, s.srv, "POST", path, s.adminToken(t), body); got != http.StatusServiceUnavailable {
			t.Errorf("token DMing a raw snowflake: %d, want 503 (no bot configured)", got)
		}
	})
}

// --- GET /me --------------------------------------------------------

func TestMeReportsAdmin(t *testing.T) {
	s := newAdminStack(t, listedDiscordID)
	listed, user := s.signedIn(t, listedDiscordID, "Owner")
	unlisted, _ := s.signedIn(t, unlistedDiscordID, "Stranger")
	reclaim := s.issue(t, auth.Principal{Role: auth.RolePlayer, GameID: uuid.New(), PlayerID: uuid.New(), DiscordID: listedDiscordID})
	cases := []struct {
		name, token string
		want        bool
	}{
		{"shared token", s.adminToken(t), true},
		{"allowlisted", listed, true},
		{"unlisted", unlisted, false},
		{"reclaim ticket with a listed Discord ID", reclaim, false},
	}
	for _, tc := range cases {
		resp := doGet(t, s.srv, "/me", tc.token)
		var body map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s: /me %d", tc.name, resp.StatusCode)
		}
		if got, _ := body["admin"].(bool); got != tc.want {
			t.Errorf("%s: /me admin = %v, want %v", tc.name, body["admin"], tc.want)
		}
		if _, ok := body["role"]; !ok {
			t.Errorf("%s: /me no longer carries the principal's fields: %v", tc.name, body)
		}
	}
	// The principal stays flat, as before.
	resp := doGet(t, s.srv, "/me", listed)
	var flat struct {
		Role   string `json:"role"`
		UserID string `json:"user_id"`
		Admin  bool   `json:"admin"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&flat)
	resp.Body.Close()
	if flat.Role != "identified" || flat.UserID != user.String() || !flat.Admin {
		t.Errorf("/me = %+v", flat)
	}
	// Never the list.
	s.admins.Replace([]string{listedDiscordID, "333333333333333333"})
	resp = doGet(t, s.srv, "/me", listed)
	raw, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if strings.Contains(string(raw), "333333333333333333") {
		t.Error("/me serves the allowlist")
	}
}

// --- WebSocket parity -----------------------------------------------

func upgradeRequest(token, query string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/ws?"+query, nil)
	r.Header.Set("Authorization", "Bearer "+token)
	return r
}

// dialStatus opens the stack's real /ws with token and query and
// returns the upgrade's HTTP status: 101 when it was accepted.
func dialStatus(t *testing.T, srv *httptest.Server, token, query string) int {
	t.Helper()
	u := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?" + query
	h := http.Header{}
	h.Set("Authorization", "Bearer "+token)
	conn, resp, err := websocket.DefaultDialer.Dial(u, h)
	if conn != nil {
		_ = conn.Close()
	}
	if resp == nil {
		t.Fatalf("dial %s: %v", query, err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

func TestWSAdminParityForAnAllowlistedUser(t *testing.T) {
	s := newAdminStack(t, listedDiscordID)
	gameA, gameB := uuid.New(), uuid.New()
	seatA, seatB := uuid.New(), uuid.New()
	user := uuid.New()
	identified := s.issue(t, auth.Principal{Role: auth.RoleIdentified, UserID: user, DiscordID: listedDiscordID})
	seated := s.issue(t, auth.Principal{Role: auth.RolePlayer, UserID: user, GameID: gameA, PlayerID: seatA, DiscordID: listedDiscordID})
	watching := s.issue(t, auth.Principal{Role: auth.RoleSpectator, UserID: user, GameID: gameA, DiscordID: listedDiscordID})

	cases := []struct {
		name, token, query string
		want               ws.Binding
	}{
		{"own seat, no ?game=", seated, "", ws.Binding{GameID: gameA, PlayerID: seatA, UserID: user, Admin: true}},
		{"own seat at own game", seated, "game=" + gameA.String() + "&player=" + seatA.String(), ws.Binding{GameID: gameA, PlayerID: seatA, UserID: user, Admin: true}},
		{"another seat at own game", seated, "game=" + gameA.String() + "&player=" + seatB.String(), ws.Binding{GameID: gameA, PlayerID: seatB, UserID: user, Admin: true}},
		{"another game, watching", seated, "game=" + gameB.String(), ws.Binding{GameID: gameB, UserID: user, Admin: true}},
		{"another game, as a seat", seated, "game=" + gameB.String() + "&player=" + seatB.String(), ws.Binding{GameID: gameB, PlayerID: seatB, UserID: user, Admin: true}},
		{"identified, any game", identified, "game=" + gameB.String(), ws.Binding{GameID: gameB, UserID: user, Admin: true}},
		{"own spectator session", watching, "game=" + gameA.String(), ws.Binding{GameID: gameA, ReadOnly: true, UserID: user, Admin: true}},
	}
	for _, tc := range cases {
		b, err := s.wsAuth.AuthorizeUpgrade(upgradeRequest(tc.token, tc.query))
		if err != nil {
			t.Errorf("%s: refused: %v", tc.name, err)
			continue
		}
		b.IssuedAt = time.Time{}
		if b != tc.want {
			t.Errorf("%s: binding %+v, want %+v", tc.name, b, tc.want)
		}
	}
	if got := dialStatus(t, s.srv, identified, ""); got != http.StatusBadRequest {
		t.Errorf("identified admin with no ?game=: %d, want 400 missing game id", got)
	}

	logs := s.log.String()
	if !strings.Contains(logs, "admin websocket binding") ||
		!strings.Contains(logs, "admin_user_id="+user.String()) ||
		!strings.Contains(logs, "game_id="+gameB.String()) ||
		!strings.Contains(logs, "seat="+seatB.String()) {
		t.Errorf("admin bindings were not logged with the user, game and seat:\n%s", logs)
	}
	for _, tok := range []string{identified, seated, watching} {
		if strings.Contains(logs, tok) {
			t.Error("the binding log carries a session token")
		}
	}
}

func TestWSUnlistedUserIsRefusedAndRemovalTakesEffect(t *testing.T) {
	s := newAdminStack(t, listedDiscordID)
	gameA, gameB := uuid.New(), uuid.New()
	seatA, seatB := uuid.New(), uuid.New()
	user := uuid.New()
	unlistedIdentified := s.issue(t, auth.Principal{Role: auth.RoleIdentified, UserID: uuid.New(), DiscordID: unlistedDiscordID})
	unlistedSeat := s.issue(t, auth.Principal{Role: auth.RolePlayer, UserID: uuid.New(), GameID: gameA, PlayerID: seatA, DiscordID: unlistedDiscordID})
	reclaim := s.issue(t, auth.Principal{Role: auth.RolePlayer, GameID: gameA, PlayerID: seatA, DiscordID: listedDiscordID})
	listedSeat := s.issue(t, auth.Principal{Role: auth.RolePlayer, UserID: user, GameID: gameA, PlayerID: seatA, DiscordID: listedDiscordID})

	for _, tc := range []struct{ name, token, query string }{
		{"unlisted identified", unlistedIdentified, "game=" + gameB.String()},
		{"unlisted seat, another game", unlistedSeat, "game=" + gameB.String()},
		{"reclaim ticket, another game", reclaim, "game=" + gameB.String()},
	} {
		if got := dialStatus(t, s.srv, tc.token, tc.query); got != http.StatusForbidden {
			t.Errorf("%s: %d, want 403", tc.name, got)
		}
	}
	// An unlisted seat asking for another seat at its own table still
	// gets its own seat, exactly as before: ?player= is ignored.
	b, err := s.wsAuth.AuthorizeUpgrade(upgradeRequest(unlistedSeat, "game="+gameA.String()+"&player="+seatB.String()))
	if err != nil || b.PlayerID != seatA || b.Admin {
		t.Errorf("unlisted seat naming another seat: %+v, %v; want its own seat, not admin", b, err)
	}

	if _, err := s.wsAuth.AuthorizeUpgrade(upgradeRequest(listedSeat, "game="+gameB.String())); err != nil {
		t.Fatalf("listed seat, another game: %v", err)
	}
	s.admins.Replace(nil)
	if got := dialStatus(t, s.srv, listedSeat, "game="+gameB.String()); got != http.StatusForbidden {
		t.Errorf("after removal, same token, another game: %d, want 403", got)
	}
	b, err = s.wsAuth.AuthorizeUpgrade(upgradeRequest(listedSeat, "game="+gameA.String()))
	if err != nil || b.Admin {
		t.Errorf("after removal, own seat: %+v, %v; want an ordinary binding", b, err)
	}
	if strings.Contains(s.log.String(), "user_id="+uuid.Nil.String()) {
		t.Error("a non-admin binding was logged as admin")
	}
}

// --- the owner's requirement: an admin at a table as themselves -----

// TestAllowlistedUserSitsAsThemselvesAndIsAdmin is the owner's
// requirement of 2026-10-02: "I want to be admin and able to join under
// my same discord." The allowlisted person joins by invite link and by
// code, gets their own seat under their own Discord identity, and that
// seat session is an admin: /me says so, an admin route answers, and
// the socket reaches another table. A seated player who is not listed
// is refused the same route.
func TestAllowlistedUserSitsAsThemselvesAndIsAdmin(t *testing.T) {
	s := newAdminStack(t, listedDiscordID)
	listed, user := s.signedIn(t, listedDiscordID, "Owner")
	unlisted, _ := s.signedIn(t, unlistedDiscordID, "Stranger")

	byLink, err := s.lobby.Create("joined by link")
	if err != nil {
		t.Fatal(err)
	}
	byCode, err := s.lobby.Create("joined by code")
	if err != nil {
		t.Fatal(err)
	}
	elsewhere, err := s.lobby.Create("somebody else's table")
	if err != nil {
		t.Fatal(err)
	}

	join := func(t *testing.T, path, token string, body joinRequest) sessionResponse {
		t.Helper()
		resp := postJSON(t, s.srv, path, token, body)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			raw, _ := io.ReadAll(resp.Body)
			t.Fatalf("join %s: %d %s", path, resp.StatusCode, raw)
		}
		var sess sessionResponse
		if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil {
			t.Fatal(err)
		}
		return sess
	}

	for _, tc := range []struct {
		name string
		meta GameMeta
		sess func() sessionResponse
	}{
		{"invite link", byLink, func() sessionResponse {
			return join(t, "/games/"+byLink.ID.String()+"/join", listed, joinRequest{InviteToken: byLink.InviteToken, Name: "typed name"})
		}},
		{"code", byCode, func() sessionResponse {
			return join(t, "/join", listed, joinRequest{InviteToken: byCode.InviteToken, Name: "typed name"})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sess := tc.sess()
			p := sess.Principal
			if p.Role != auth.RolePlayer || p.UserID != user || p.DiscordID != listedDiscordID || p.Name != "Owner" {
				t.Fatalf("seat session %+v: want a player seat as the Discord identity Owner", p)
			}
			meta, err := s.lobby.Get(tc.meta.ID)
			if err != nil {
				t.Fatal(err)
			}
			var seat *SeatInfo
			for i := range meta.Players {
				if meta.Players[i].PlayerID == sess.PlayerID {
					seat = &meta.Players[i]
				}
			}
			if seat == nil || seat.DiscordID != listedDiscordID || (seat.Name != "Owner" && seat.DisplayName != "Owner") || seat.UserID != user.String() {
				t.Fatalf("seat %+v: want the admin's own Discord identity and user", seat)
			}

			resp := doGet(t, s.srv, "/me", sess.Token)
			var me meResponse
			_ = json.NewDecoder(resp.Body).Decode(&me)
			resp.Body.Close()
			if !me.Admin || me.PlayerID != sess.PlayerID {
				t.Errorf("/me from the seat: admin=%v player=%s; want admin at their own seat", me.Admin, me.PlayerID)
			}

			// An admin route, from the seat session.
			if got := status(t, s.srv, "POST", "/games/"+elsewhere.ID.String()+"/archive", sess.Token, nil); got != http.StatusNoContent && got != http.StatusOK {
				t.Errorf("archive another table from the seat: %d, want success", got)
			}
			_ = status(t, s.srv, "DELETE", "/games/"+elsewhere.ID.String()+"/archive", sess.Token, nil)

			// Plays as their own seat by default…
			b, err := s.wsAuth.AuthorizeUpgrade(upgradeRequest(sess.Token, "game="+tc.meta.ID.String()+"&player="+sess.PlayerID.String()))
			if err != nil || b.PlayerID != sess.PlayerID || !b.Admin {
				t.Errorf("own table socket: %+v, %v; want their own seat with admin", b, err)
			}
			// …and reaches any other table, like the token.
			b, err = s.wsAuth.AuthorizeUpgrade(upgradeRequest(sess.Token, "game="+elsewhere.ID.String()))
			if err != nil || b.GameID != elsewhere.ID || !b.Admin {
				t.Errorf("another table's socket: %+v, %v; want an admin binding", b, err)
			}
		})
	}

	// A seated player who is not listed: same join, same route, refused.
	sess := join(t, "/games/"+byLink.ID.String()+"/join", unlisted, joinRequest{InviteToken: byLink.InviteToken})
	if sess.Principal.DiscordID != unlistedDiscordID {
		t.Fatalf("unlisted seat %+v", sess.Principal)
	}
	if got := status(t, s.srv, "POST", "/games/"+elsewhere.ID.String()+"/archive", sess.Token, nil); got != http.StatusForbidden {
		t.Errorf("unlisted seated player archiving a table: %d, want 403", got)
	}
	resp := doGet(t, s.srv, "/me", sess.Token)
	var me meResponse
	_ = json.NewDecoder(resp.Body).Decode(&me)
	resp.Body.Close()
	if me.Admin {
		t.Error("/me calls an unlisted seated player admin")
	}
}

// TestAllowlistedSeatManagesAnyTable: the host-or-admin routes take
// isAdmin's answer, so an admin seated at one table manages another,
// and the engine records them as "the admin" there, not as their seat
// at a table it has never heard of.
func TestAllowlistedSeatManagesAnyTable(t *testing.T) {
	s := newAdminStack(t, listedDiscordID)
	listed, _ := s.signedIn(t, listedDiscordID, "Owner")
	mine, err := s.lobby.Create("mine")
	if err != nil {
		t.Fatal(err)
	}
	theirs, err := s.lobby.Create("theirs")
	if err != nil {
		t.Fatal(err)
	}
	resp := postJSON(t, s.srv, "/games/"+mine.ID.String()+"/join", listed, joinRequest{InviteToken: mine.InviteToken})
	var sess sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&sess)
	resp.Body.Close()

	limit := 3
	resp = do(t, s.srv, http.MethodPatch, "/games/"+theirs.ID.String()+"/settings", sess.Token, map[string]any{"undo_limit": limit})
	raw, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("admin seat patching another table's settings: %d %s", resp.StatusCode, raw)
	}
	if got := s.lobby.RoomOf(theirs.ID).Game.TableSettingsSnapshot().UndoLimit; got != limit {
		t.Errorf("undo limit = %d, want %d", got, limit)
	}
	if got := actorIn(sess.Principal, theirs.ID); got != uuid.Nil {
		t.Errorf("actorIn at another table = %s, want uuid.Nil (the admin)", got)
	}
	if got := actorIn(sess.Principal, mine.ID); got != sess.PlayerID {
		t.Errorf("actorIn at their own table = %s, want their seat", got)
	}

	// Their GET of the other table shows its invites, like the token's.
	resp = doGet(t, s.srv, "/games/"+theirs.ID.String(), sess.Token)
	var meta GameMeta
	_ = json.NewDecoder(resp.Body).Decode(&meta)
	resp.Body.Close()
	if meta.InviteToken == "" {
		t.Error("an admin seated elsewhere does not see this table's invite")
	}
}

// TestUploadDeckAnySeatIsAdminOnly: "an admin may set any seat's deck"
// now asks isAdmin, and everyone else has to BE the seat. Before ADR
// 0110 only a RolePlayer session was checked, so a spectator or an
// unseated sign-in could set any seat's deck.
func TestUploadDeckAnySeatIsAdminOnly(t *testing.T) {
	s := newAdminStackWithCards(t, buildMinimalDeckIndex(t), listedDiscordID)
	meta, err := s.lobby.Create("decks")
	if err != nil {
		t.Fatal(err)
	}
	_, alice, err := s.lobby.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	valid := "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n"
	body := uploadDeckRequest{Format: "text", Source: valid, PlayerID: alice}
	path := "/games/" + meta.ID.String() + "/decks"

	spectator := s.issue(t, auth.Principal{Role: auth.RoleSpectator, GameID: meta.ID, Name: "watcher"})
	unlisted, _ := s.signedIn(t, unlistedDiscordID, "Stranger")
	for _, who := range []struct{ name, token string }{{"spectator", spectator}, {"unlisted sign-in", unlisted}} {
		if got := status(t, s.srv, "POST", path, who.token, body); got != http.StatusForbidden {
			t.Errorf("%s setting Alice's deck: %d, want 403", who.name, got)
		}
	}
	listed, _ := s.signedIn(t, listedDiscordID, "Owner")
	if got := status(t, s.srv, "POST", path, listed, body); got != http.StatusOK {
		t.Errorf("allowlisted admin setting Alice's deck: %d, want 200", got)
	}
	if got := status(t, s.srv, "POST", path, s.adminToken(t), body); got != http.StatusOK {
		t.Errorf("token setting Alice's deck: %d, want 200", got)
	}
}
