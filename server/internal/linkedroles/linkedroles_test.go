package linkedroles

// Tests for ADR 0163's service: the stats query over a real database,
// and linking, pushing, unlinking and deletion against a fake Discord
// (an httptest.Server behind the real discord calls). No real Discord
// call is made.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
)

const (
	testAppID = "777777777777777777"
	testKey   = "0123456789abcdef0123456789abcdef-identity"
	aliceID   = "111"
)

// fakeDiscord is Discord's token endpoint and REST API. Refresh token
// "rt-N" becomes access token "at-N" and refresh token "rt-N+1"; a token
// in rejected is refused with invalid_grant. PUTs are recorded with the
// bearer token they carried, and answered from putReplies (200 when they
// run out).
type fakeDiscord struct {
	mu          sync.Mutex
	rejected    map[string]bool
	refreshDown bool
	refreshes   []string
	putReplies  []int
	puts        []put
	metadata    [][]byte
}

type put struct {
	auth string
	body map[string]any
}

func newFakeDiscord(t *testing.T) *fakeDiscord {
	t.Helper()
	f := &fakeDiscord{rejected: map[string]bool{}}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		f.mu.Lock()
		defer f.mu.Unlock()
		rt := r.PostForm.Get("refresh_token")
		f.refreshes = append(f.refreshes, rt)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case f.refreshDown:
			w.WriteHeader(http.StatusBadGateway)
			return
		case f.rejected[rt] || !strings.HasPrefix(rt, "rt-"):
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
			return
		}
		n := strings.TrimPrefix(rt, "rt-")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "at-" + n, "token_type": "Bearer", "expires_in": 604800,
			"refresh_token": "rt-" + n + "+", "scope": "identify role_connections.write",
		})
	})
	mux.HandleFunc("PUT /applications/{app}/role-connections/metadata", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		f.mu.Lock()
		f.metadata = append(f.metadata, b)
		f.mu.Unlock()
		_, _ = w.Write([]byte(`[]`))
	})
	mux.HandleFunc("PUT /users/@me/applications/{app}/role-connection", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.mu.Lock()
		defer f.mu.Unlock()
		f.puts = append(f.puts, put{auth: r.Header.Get("Authorization"), body: body})
		status := http.StatusOK
		if len(f.putReplies) > 0 {
			status, f.putReplies = f.putReplies[0], f.putReplies[1:]
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if status == http.StatusTooManyRequests {
			_, _ = w.Write([]byte(`{"retry_after":60,"global":false}`))
			return
		}
		_, _ = w.Write([]byte(`{}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	t.Cleanup(discord.SwapEndpointsForTesting(srv.URL+"/oauth2/token", srv.URL+"/users/@me"))
	t.Cleanup(discord.SwapBotAPIBaseForTesting(srv.URL))
	return f
}

func (f *fakeDiscord) snapshot() (refreshes []string, puts []put) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.refreshes...), append([]put(nil), f.puts...)
}

// stack is a migrated database with one signed-in person, the store
// over it, and the service with the real Discord calls pointed at the
// fake. waits records each rate-limit wait instead of sleeping.
type stack struct {
	d     *db.DB
	store *users.SQLStore
	svc   *Service
	fake  *fakeDiscord
	user  uuid.UUID
	logs  *bytes.Buffer
	mu    sync.Mutex
	waits []time.Duration
}

func newStack(t *testing.T) *stack {
	t.Helper()
	fake := newFakeDiscord(t)
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	sealer, err := users.NewSealer(testKey)
	if err != nil {
		t.Fatal(err)
	}
	store := users.NewSQLStore(d, sealer)
	u, err := store.UpsertFromDiscord(context.Background(),
		discord.User{ID: aliceID, Username: "alice", GlobalName: "Alice"}, "signin-rt", "identify")
	if err != nil {
		t.Fatal(err)
	}
	s := &stack{d: d, store: store, fake: fake, user: u.ID, logs: &bytes.Buffer{}}
	api := NewDiscord(discord.Config{ClientID: "cid", ClientSecret: "sec", RedirectURI: "http://x/cb"},
		discord.LinkedRoles{AppID: testAppID, Token: "bot-secret", Requested: true}, nil)
	s.svc = New(api, store, NewSQLStats(d), slog.New(slog.NewTextHandler(s.logs, nil)))
	s.svc.sleep = func(ctx context.Context, d time.Duration) error {
		if d == pushGap {
			return ctx.Err()
		}
		s.mu.Lock()
		s.waits = append(s.waits, d)
		s.mu.Unlock()
		return ctx.Err()
	}
	return s
}

// grant is what the callback's code exchange returns.
func grant(rt string) discord.TokenResponse {
	return discord.TokenResponse{AccessToken: "at-link", RefreshToken: rt, Scope: "identify role_connections.write"}
}

func exec(t *testing.T, d *db.DB, q string, args ...any) {
	t.Helper()
	if _, err := d.Exec(q, args...); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
}

// game seeds one stored game with the person at seat 0, a stranger's
// guest at seat 1 and a bot at seat 2.
func (s *stack) game(t *testing.T, state string, outcome any, winner any) {
	t.Helper()
	id := uuid.NewString()
	exec(t, s.d, `INSERT INTO games (id, name, state, created_at, ended_at, winner_seat, outcome) VALUES (?, 'T', ?, 1, 2, ?, ?)`,
		id, state, winner, outcome)
	exec(t, s.d, `INSERT INTO seats (game_id, seat, player_id, user_id, guest_name) VALUES (?, 0, ?, ?, 'Alice')`,
		id, uuid.NewString(), s.user.String())
	exec(t, s.d, `INSERT INTO seats (game_id, seat, player_id, guest_name) VALUES (?, 1, ?, 'Guest')`, id, uuid.NewString())
	exec(t, s.d, `INSERT INTO seats (game_id, seat, player_id, guest_name, bot_tier) VALUES (?, 2, ?, 'Bot', 'heuristic')`,
		id, uuid.NewString())
}

func metadataOf(p put) map[string]any {
	m, _ := p.body["metadata"].(map[string]any)
	return m
}

func TestStatsCountsOnlyFinishedGamesAndWins(t *testing.T) {
	s := newStack(t)
	s.game(t, "ended", "win", 0) // won
	s.game(t, "ended", "win", 1) // lost
	s.game(t, "ended", "draw", nil)
	s.game(t, "ended", nil, nil)  // abandoned: not counted
	s.game(t, "active", nil, nil) // in play: not counted
	s.game(t, "lobby", nil, nil)
	st, err := NewSQLStats(s.d).Stats(context.Background(), s.user)
	if err != nil {
		t.Fatal(err)
	}
	if st.GamesPlayed != 3 || st.GamesWon != 1 || st.Name != "Alice" || st.Joined.IsZero() {
		t.Errorf("stats = %+v, want 3 played, 1 won, named Alice", st)
	}
	if _, err := NewSQLStats(s.d).Stats(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown account: %v, want ErrNotFound", err)
	}
	exec(t, s.d, `UPDATE users SET deleted_at = 5 WHERE id = ?`, s.user.String())
	if _, err := NewSQLStats(s.d).Stats(context.Background(), s.user); !errors.Is(err, ErrNotFound) {
		t.Errorf("deleted account: %v, want ErrNotFound", err)
	}
}

func TestLinkStoresTheGrantSealedAndPushesTheNumbers(t *testing.T) {
	s := newStack(t)
	s.game(t, "ended", "win", 0)
	ctx := context.Background()
	if err := s.svc.Link(ctx, s.user, aliceID, grant("rt-1")); err != nil {
		t.Fatalf("Link: %v", err)
	}
	if rt, ok, err := s.store.RoleConnectionToken(ctx, aliceID); err != nil || !ok || rt != "rt-1" {
		t.Errorf("stored grant = %q %v %v, want rt-1", rt, ok, err)
	}
	if rt, _, _ := s.store.RefreshToken(ctx, aliceID); rt != "signin-rt" {
		t.Errorf("the sign-in grant changed: %q", rt)
	}
	refreshes, puts := s.fake.snapshot()
	if len(refreshes) != 0 {
		t.Errorf("linking refreshed %v; the code exchange's access token is enough", refreshes)
	}
	if len(puts) != 1 || puts[0].auth != "Bearer at-link" {
		t.Fatalf("puts = %+v, want one with the grant's access token", puts)
	}
	m := metadataOf(puts[0])
	if m["games_played"] != "1" || m["games_won"] != "1" || m["joined"] == "" ||
		puts[0].body["platform_name"] != "CMD & CTRL" || puts[0].body["platform_username"] != "Alice" {
		t.Errorf("pushed %+v", puts[0].body)
	}
}

func TestLinkRefusesAGrantItCannotUse(t *testing.T) {
	s := newStack(t)
	ctx := context.Background()
	noScope := grant("rt-1")
	noScope.Scope = "identify"
	if err := s.svc.Link(ctx, s.user, aliceID, noScope); !errors.Is(err, ErrScopeMissing) {
		t.Errorf("no role_connections.write: %v, want ErrScopeMissing", err)
	}
	if err := s.svc.Link(ctx, s.user, aliceID, grant("")); !errors.Is(err, ErrNoRefreshToken) {
		t.Errorf("no refresh token: %v, want ErrNoRefreshToken", err)
	}
	// A first push that fails keeps nothing.
	s.fake.mu.Lock()
	s.fake.putReplies = []int{http.StatusBadGateway}
	s.fake.mu.Unlock()
	if err := s.svc.Link(ctx, s.user, aliceID, grant("rt-1")); err == nil {
		t.Error("a failed first push linked")
	}
	if _, ok, _ := s.store.RoleConnectionToken(ctx, aliceID); ok {
		t.Error("a failed link kept the grant")
	}
	if _, puts := s.fake.snapshot(); len(puts) != 1 {
		t.Errorf("%d puts, want only the failed one", len(puts))
	}
}

func TestLinkWaitsOnceOnARateLimitCappedForThePersonWaiting(t *testing.T) {
	s := newStack(t)
	s.fake.mu.Lock()
	s.fake.putReplies = []int{http.StatusTooManyRequests}
	s.fake.mu.Unlock()
	if err := s.svc.Link(context.Background(), s.user, aliceID, grant("rt-1")); err != nil {
		t.Fatalf("Link after one 429: %v", err)
	}
	if len(s.waits) != 1 || s.waits[0] != linkRetryWait {
		t.Errorf("waits = %v, want one of %v (Discord asked for 60s)", s.waits, linkRetryWait)
	}
}

func TestPushAfterAGameRefreshesRotatesAndRecomputes(t *testing.T) {
	s := newStack(t)
	ctx := context.Background()
	if err := s.svc.Link(ctx, s.user, aliceID, grant("rt-1")); err != nil {
		t.Fatal(err)
	}
	s.game(t, "ended", "win", 0)
	s.game(t, "ended", "draw", nil)

	runCtx, stop := context.WithCancel(ctx)
	defer stop()
	go s.svc.Run(runCtx)
	// The same person twice before the worker runs is one push.
	s.svc.Enqueue(s.user, s.user, uuid.Nil)
	s.svc.Wait()

	refreshes, puts := s.fake.snapshot()
	if len(refreshes) != 1 || refreshes[0] != "rt-1" {
		t.Errorf("refreshes = %v, want one with the stored grant", refreshes)
	}
	if rt, _, _ := s.store.RoleConnectionToken(ctx, aliceID); rt != "rt-1+" {
		t.Errorf("the rotated grant was not stored: %q", rt)
	}
	if len(puts) != 2 {
		t.Fatalf("%d puts, want the link's and one after the games", len(puts))
	}
	last := puts[1]
	if last.auth != "Bearer at-1" {
		t.Errorf("pushed with %q, want the refreshed access token", last.auth)
	}
	if m := metadataOf(last); m["games_played"] != "2" || m["games_won"] != "1" {
		t.Errorf("pushed %v, want 2 played and 1 won, recomputed", m)
	}
}

func TestPushSkipsAnAccountThatIsNotLinked(t *testing.T) {
	s := newStack(t)
	s.game(t, "ended", "win", 0)
	s.svc.Push(context.Background(), s.user)
	s.svc.Push(context.Background(), uuid.New())
	if refreshes, puts := s.fake.snapshot(); len(refreshes) != 0 || len(puts) != 0 {
		t.Errorf("an unlinked account made %d refreshes and %d puts", len(refreshes), len(puts))
	}
}

func TestPushUnlinksWhenDiscordRejectsTheGrant(t *testing.T) {
	s := newStack(t)
	ctx := context.Background()
	if err := s.svc.Link(ctx, s.user, aliceID, grant("rt-1")); err != nil {
		t.Fatal(err)
	}
	s.fake.mu.Lock()
	s.fake.rejected["rt-1"] = true
	s.fake.mu.Unlock()
	s.svc.Push(ctx, s.user)
	if _, ok, _ := s.store.RoleConnectionToken(ctx, aliceID); ok {
		t.Error("an invalid_grant left the person linked")
	}
	if rt, _, _ := s.store.RefreshToken(ctx, aliceID); rt != "signin-rt" {
		t.Errorf("unlinking touched the sign-in grant: %q", rt)
	}
	if _, puts := s.fake.snapshot(); len(puts) != 1 {
		t.Errorf("%d puts, want only the link's", len(puts))
	}
}

func TestPushLeavesAnOutageForTheNextGame(t *testing.T) {
	s := newStack(t)
	ctx := context.Background()
	if err := s.svc.Link(ctx, s.user, aliceID, grant("rt-1")); err != nil {
		t.Fatal(err)
	}
	s.fake.mu.Lock()
	s.fake.refreshDown = true
	s.fake.mu.Unlock()
	s.svc.Push(ctx, s.user)
	if rt, ok, _ := s.store.RoleConnectionToken(ctx, aliceID); !ok || rt != "rt-1" {
		t.Errorf("an outage changed the grant: %q %v", rt, ok)
	}
	if !strings.Contains(s.logs.String(), "level=WARN") {
		t.Errorf("an outage was not logged at WARN:\n%s", s.logs)
	}

	// A 5xx on the PUT: warned, one try, no wait.
	s.fake.mu.Lock()
	s.fake.refreshDown = false
	s.fake.putReplies = []int{http.StatusServiceUnavailable}
	s.fake.mu.Unlock()
	s.svc.Push(ctx, s.user)
	if _, puts := s.fake.snapshot(); len(puts) != 2 {
		t.Errorf("%d puts, want the link's and one failed try", len(puts))
	}
	if len(s.waits) != 0 {
		t.Errorf("a 5xx waited %v", s.waits)
	}
}

func TestPushWaitsOnceOnARateLimitThenGivesUp(t *testing.T) {
	s := newStack(t)
	ctx := context.Background()
	if err := s.svc.Link(ctx, s.user, aliceID, grant("rt-1")); err != nil {
		t.Fatal(err)
	}
	s.fake.mu.Lock()
	s.fake.putReplies = []int{http.StatusTooManyRequests, http.StatusTooManyRequests}
	s.fake.mu.Unlock()
	s.svc.Push(ctx, s.user)
	if _, puts := s.fake.snapshot(); len(puts) != 3 {
		t.Errorf("%d puts, want the link's and two tries", len(puts))
	}
	if len(s.waits) != 1 || s.waits[0] != maxRetryWait {
		t.Errorf("waits = %v, want one capped at %v (Discord asked for 60s)", s.waits, maxRetryWait)
	}
}

func TestForgetEmptiesTheConnectionOfADeletedAccount(t *testing.T) {
	s := newStack(t)
	ctx := context.Background()
	if err := s.svc.Link(ctx, s.user, aliceID, grant("rt-1")); err != nil {
		t.Fatal(err)
	}
	var sealed []byte
	if err := s.d.QueryRow(`SELECT role_connection_token FROM identities WHERE subject = ?`, aliceID).Scan(&sealed); err != nil {
		t.Fatal(err)
	}
	exec(t, s.d, `DELETE FROM identities WHERE subject = ?`, aliceID)
	s.svc.Forget(aliceID, sealed)
	s.svc.Forget("", sealed)   // nothing to do
	s.svc.Forget(aliceID, nil) // not linked
	s.svc.Wait()
	_, puts := s.fake.snapshot()
	if len(puts) != 2 {
		t.Fatalf("%d puts, want the link's and the clear", len(puts))
	}
	clear := puts[1]
	if clear.auth != "Bearer at-1" || clear.body["platform_name"] != nil || clear.body["platform_username"] != nil || len(metadataOf(clear)) != 0 {
		t.Errorf("clear = %+v, want an empty connection with the refreshed token", clear)
	}
}

func TestRegisterSendsTheMetadata(t *testing.T) {
	s := newStack(t)
	if err := s.svc.Register(context.Background()); err != nil {
		t.Fatal(err)
	}
	s.fake.mu.Lock()
	defer s.fake.mu.Unlock()
	if len(s.fake.metadata) != 1 || !bytes.Contains(s.fake.metadata[0], []byte(`"games_played"`)) {
		t.Errorf("metadata = %q", s.fake.metadata)
	}
}

func TestEnqueueNeverBlocksAndDropsPastTheCap(t *testing.T) {
	s := newStack(t)
	ids := make([]uuid.UUID, maxPending+5)
	for i := range ids {
		ids[i] = uuid.New()
	}
	done := make(chan struct{})
	go func() { s.svc.Enqueue(ids...); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Enqueue blocked with no worker running")
	}
	s.svc.mu.Lock()
	n := len(s.svc.pending)
	s.svc.mu.Unlock()
	if n != maxPending {
		t.Errorf("%d pending, want the cap %d", n, maxPending)
	}
	// A stopped worker releases them, so Wait returns at shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s.svc.Run(ctx)
	s.svc.Wait()
}
