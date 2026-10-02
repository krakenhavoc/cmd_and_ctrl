package lobby

// Tests for ADR 0110 Delivery PR 7's opening of POST /games (§5 item 4,
// owner answer 2): any signed-in person may create a table, capped at
// three open tables and one creation per 30 seconds, and becomes its
// creator. A guest is refused and the shared admin token keeps working,
// uncapped, for the Discord bot.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decklibrary"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/tablesetups"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// tableStack is a database-backed stack with Discord sign-in, the
// stores PR 7 reads and writes, a card index that resolves both the
// test bot deck and the first pre-built deck, and a fake bot host.
type tableStack struct {
	userStack
	host    *fakeBotHost
	setups  *tablesetups.SQLStore
	library decklibrary.Store
}

// testDeckSource is the bot decklist every tableStack bot plays.
const testDeckSource = "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n"

// newTableStack relaxes the IP rate limits, which a test that signs
// in and joins several people would otherwise trip; newStrictTableStack
// keeps every limit, for the test of the creation bucket itself.
func newTableStack(t *testing.T) *tableStack {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	return newTableStackWith(t)
}

func newStrictTableStack(t *testing.T) *tableStack {
	t.Helper()
	return newTableStackWith(t)
}

func newTableStackWith(t *testing.T) *tableStack {
	t.Helper()
	host := newFakeBotHost()
	idx := buildPrebuiltDeckIndex(t, decks.All()[0])
	for _, c := range buildMinimalDeckIndexCards() {
		idx.Put(c)
	}
	d := openTestDB(t, t.TempDir())
	us := users.NewSQLStore(d, nil)
	setups := tablesetups.NewSQLStore(d)
	library := decklibrary.NewSQLStore(d)
	a, err := auth.NewHMACAuthenticator([]byte("lobby-test-session-key-0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewHMACAuthenticator: %v", err)
	}
	srv, l, stub, state := newDiscordTestStackWith(t, func(c *Config) {
		c.Lobby = NewLobbyWithStore(ws.NewRoomManager(quietLogger(), ""), NewSQLStore(d))
		c.Lobby.SetBotHost(host)
		c.Auth = a
		c.Users = us
		c.TableSetups = setups
		c.DeckLibrary = library
		c.Bots = host
		c.BotDecks = fakeDeckSource{}
		c.Cards = idx
		c.Log = discardLogger()
	})
	return &tableStack{
		userStack: userStack{srv: srv, lobby: l, stub: stub, state: state, users: us, auth: a, db: d},
		host:      host,
		setups:    setups,
		library:   library,
	}
}

func buildMinimalDeckIndexCards() []cards.Card {
	return []cards.Card{
		{ID: uuid.New(), Name: "Test Commander", TypeLine: "Legendary Creature — Human Wizard", ColorIdentity: []string{"W"}, Legalities: map[string]string{"commander": "legal"}},
		{ID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains", ColorIdentity: []string{"W"}, Legalities: map[string]string{"commander": "legal"}},
	}
}

// signIn walks a Discord sign-in as the given account and returns the
// identity session and its user.
func (s *tableStack) signIn(t *testing.T, discordID, name string) (string, uuid.UUID) {
	t.Helper()
	s.stub.setUser(fmt.Sprintf(`{"id":%q,"username":%q,"global_name":%q,"avatar":"av"}`, discordID, strings.ToLower(name), name))
	tok := identityTokenFromCallback(t, s.srv, s.state)
	return tok, mustValidate(t, s.auth, tok).UserID
}

func (s *tableStack) admin(t *testing.T) string {
	t.Helper()
	return adminSession(t, s.auth)
}

// create posts POST /games and decodes the answer.
func (s *tableStack) create(t *testing.T, tok string, body createGameRequest) (int, createGameResponse, http.Header, string) {
	t.Helper()
	resp := postJSON(t, s.srv, "/games", tok, body)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out createGameResponse
	if resp.StatusCode == http.StatusCreated {
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatalf("decode create: %v (%s)", err, raw)
		}
	}
	return resp.StatusCode, out, resp.Header, string(raw)
}

func (s *tableStack) mustCreate(t *testing.T, tok, name string) createGameResponse {
	t.Helper()
	code, out, _, raw := s.create(t, tok, createGameRequest{Name: name})
	if code != http.StatusCreated {
		t.Fatalf("create %q: %d %s", name, code, raw)
	}
	return out
}

// join seats tok's person (or a guest named name, when tok is "") at
// meta's table.
func (s *tableStack) join(t *testing.T, tok string, meta GameMeta, name string) sessionResponse {
	t.Helper()
	resp := postJSON(t, s.srv, "/games/"+meta.ID.String()+"/join", tok,
		joinRequest{InviteToken: meta.InviteToken, Name: name})
	return decodeSession(t, resp, http.StatusOK)
}

func TestSignedInPlayerCreatesATableAndIsItsCreator(t *testing.T) {
	s := newTableStack(t)
	alice, aliceID := s.signIn(t, "discord-99", "Alice")

	created := s.mustCreate(t, alice, "Alice's table")
	if !created.IsCreator {
		t.Error("the creator is not told the table is theirs")
	}
	if created.InviteToken == "" || created.SpectatorInvite == "" {
		t.Fatalf("the creator did not receive both invites: %+v", created.GameMeta)
	}
	if created.Setup != nil {
		t.Errorf("a create with no setup reported one: %+v", created.Setup)
	}
	if got, _ := s.lobby.CreatedBy(created.ID); got != aliceID.String() {
		t.Errorf("created_by = %q, want %s", got, aliceID)
	}

	// GET /games/{id}: the creator keeps the invites without a seat.
	resp := doGet(t, s.srv, "/games/"+created.ID.String(), alice)
	var meta GameMeta
	_ = json.NewDecoder(resp.Body).Decode(&meta)
	resp.Body.Close()
	if meta.InviteToken != created.InviteToken || !meta.IsCreator {
		t.Errorf("GET /games/{id} for the creator: %+v", meta)
	}
	// GET /games marks it as theirs.
	resp = doGet(t, s.srv, "/games", alice)
	var list listResponse
	_ = json.NewDecoder(resp.Body).Decode(&list)
	resp.Body.Close()
	if len(list.Games) != 1 || !list.Games[0].IsCreator {
		t.Errorf("GET /games for the creator: %+v", list.Games)
	}

	// Creator powers: rotate the invites...
	if got := status(t, s.srv, "POST", "/games/"+created.ID.String()+"/invites/rotate", alice, map[string]string{"kind": "player"}); got != http.StatusOK {
		t.Errorf("creator rotating the invite: %d", got)
	}
	// ...DM tablemates before sitting down...
	full, _ := s.lobby.Get(created.ID)
	p := mustValidate(t, s.auth, alice)
	if !canInviteDM(p, full, aliceID.String(), false) {
		t.Error("the unseated creator may not DM an invite")
	}
	// ...and /c2-end: the bot's creator check answers yes for them.
	resp = doGet(t, s.srv, "/games/"+created.ID.String()+"/creator?discord_id=discord-99", s.admin(t))
	var who gameCreatorResponse
	_ = json.NewDecoder(resp.Body).Decode(&who)
	resp.Body.Close()
	if !who.IsCreator {
		t.Error("GET /games/{id}/creator does not recognise the player who created it")
	}

	// Somebody else signed in sees the table but not its invites.
	bob, _ := s.signIn(t, "discord-77", "Bob")
	resp = doGet(t, s.srv, "/games/"+created.ID.String(), bob)
	meta = GameMeta{}
	_ = json.NewDecoder(resp.Body).Decode(&meta)
	resp.Body.Close()
	if meta.InviteToken != "" || meta.IsCreator {
		t.Errorf("another person sees the creator's view: %+v", meta)
	}
}

// TestPlayerCreatedTableIsHostedByItsCreator: the table names its
// creator as host, so they host it once seated even when someone sat
// down first, and a host_discord_id naming somebody else is ignored.
func TestPlayerCreatedTableIsHostedByItsCreator(t *testing.T) {
	s := newTableStack(t)
	alice, _ := s.signIn(t, "discord-99", "Alice")
	code, created, _, raw := s.create(t, alice, createGameRequest{Name: "hosted", HostDiscordID: "discord-12345"})
	if code != http.StatusCreated {
		t.Fatalf("create: %d %s", code, raw)
	}
	s.join(t, "", created.GameMeta, "Early guest")
	seat := s.join(t, alice, created.GameMeta, "")
	meta, _ := s.lobby.Get(created.ID)
	if meta.HostPlayerID != seat.PlayerID {
		t.Errorf("host = %s, want the creator's seat %s", meta.HostPlayerID, seat.PlayerID)
	}
}

func TestCreatingTablesIsCappedPerPerson(t *testing.T) {
	s := newStrictTableStack(t)
	alice, _ := s.signIn(t, "discord-99", "Alice")
	for i := 1; i <= MaxOpenTablesPerCreator; i++ {
		s.mustCreate(t, alice, fmt.Sprintf("open %d", i))
	}

	// A fourth open table: 409, naming the three.
	code, _, _, raw := s.create(t, alice, createGameRequest{Name: "one too many"})
	if code != http.StatusConflict {
		t.Fatalf("fourth open table: %d %s", code, raw)
	}
	for i := 1; i <= MaxOpenTablesPerCreator; i++ {
		if !strings.Contains(raw, fmt.Sprintf("open %d", i)) {
			t.Errorf("the 409 does not name %q: %s", fmt.Sprintf("open %d", i), raw)
		}
	}

	// Another person is not held to Alice's cap.
	bob, _ := s.signIn(t, "discord-77", "Bob")
	s.mustCreate(t, bob, "Bob's")

	// An archived table no longer counts, but the burst is spent: three
	// creations inside 30 seconds is the bucket, so the next is a 429.
	open := s.lobby.OpenTablesCreatedBy(mustValidate(t, s.auth, alice).UserID)
	if got := status(t, s.srv, "POST", "/games/"+open[0].ID.String()+"/archive", s.admin(t), nil); got != http.StatusOK {
		t.Fatalf("archive: %d", got)
	}
	code, _, hdr, raw := s.create(t, alice, createGameRequest{Name: "too soon"})
	if code != http.StatusTooManyRequests {
		t.Fatalf("a fourth creation inside the burst: %d %s", code, raw)
	}
	if hdr.Get("Retry-After") == "" {
		t.Error("the 429 carries no Retry-After")
	}
}

// TestCreateCappedCountsOnlyOpenTables: a started table, an archived
// one and another person's do not count; the rate limiter is asked only
// once the cap has passed.
func TestCreateCappedCountsOnlyOpenTables(t *testing.T) {
	s := newTableStack(t)
	_, alice := s.signIn(t, "discord-99", "Alice")
	allowed := 0
	allow := func() bool { allowed++; return true }
	var made []GameMeta
	for i := 0; i < MaxOpenTablesPerCreator; i++ {
		m, err := s.lobby.CreateCapped(fmt.Sprintf("t%d", i), alice, "", MaxOpenTablesPerCreator, allow)
		if err != nil {
			t.Fatal(err)
		}
		made = append(made, m)
	}
	if _, err := s.lobby.CreateCapped("full", alice, "", MaxOpenTablesPerCreator, allow); !errorsIsTooMany(err) {
		t.Fatalf("fourth: %v", err)
	}
	if allowed != MaxOpenTablesPerCreator {
		t.Errorf("the limiter was asked %d times, want %d (never for a refused create)", allowed, MaxOpenTablesPerCreator)
	}
	// Start one (two bots seat it), archive another.
	for i := 0; i < 2; i++ {
		if _, _, err := s.lobby.AddBot(made[0].ID, fmt.Sprintf("B%d", i), "random", "", "d", botDeck(5)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.lobby.Start(made[0].ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.lobby.SetArchived(made[1].ID, true); err != nil {
		t.Fatal(err)
	}
	if got := len(s.lobby.OpenTablesCreatedBy(alice)); got != 1 {
		t.Errorf("open tables after a start and an archive: %d, want 1", got)
	}
	if _, err := s.lobby.CreateCapped("again", alice, "", MaxOpenTablesPerCreator, func() bool { return false }); err != ErrCreateRateLimited {
		t.Errorf("an empty bucket: %v, want ErrCreateRateLimited", err)
	}
}

func errorsIsTooMany(err error) bool {
	var e *TooManyOpenTablesError
	return errors.As(err, &e) && len(e.Open) == MaxOpenTablesPerCreator
}

func TestCreatingATableRefusesGuestsAndKeepsTheBotToken(t *testing.T) {
	s := newTableStack(t)
	table := s.mustCreate(t, s.admin(t), "someone's table")
	guest := s.join(t, "", table.GameMeta, "Guest")
	watcher, _, err := s.auth.Issue(context.Background(), auth.Principal{Role: auth.RoleSpectator, GameID: table.ID, Name: "Watcher"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	for _, who := range []struct {
		name, token string
		want        int
	}{
		{"no credential", "", http.StatusUnauthorized},
		{"guest seat", guest.Token, http.StatusForbidden},
		{"guest spectator", watcher, http.StatusForbidden},
	} {
		if got := status(t, s.srv, "POST", "/games", who.token, createGameRequest{Name: "nope"}); got != who.want {
			t.Errorf("%s: got %d, want %d", who.name, got, who.want)
		}
	}

	// The shared token is the bot's: uncapped, no creator, and it may
	// still name a host.
	admin := s.admin(t)
	for i := 0; i < MaxOpenTablesPerCreator+2; i++ {
		code, out, _, raw := s.create(t, admin, createGameRequest{Name: fmt.Sprintf("bot table %d", i), HostDiscordID: "discord-5"})
		if code != http.StatusCreated {
			t.Fatalf("bot create %d: %d %s", i, code, raw)
		}
		if out.IsCreator || out.InviteToken == "" {
			t.Errorf("bot create %d: %+v", i, out.GameMeta)
		}
		if got, _ := s.lobby.CreatedBy(out.ID); got != "" {
			t.Errorf("a token-created table recorded a creator: %q", got)
		}
	}
}
