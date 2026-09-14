package lobby

// archive_test.go covers the reversible half of "get this old game
// off my lobby screen". The irreversible half (Delete) already has
// coverage; what matters here is that archiving hides a table
// WITHOUT destroying anything, and that it does not leave a bot
// runner playing to an empty room.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

func TestArchiveHidesTheTableAndUnarchiveBringsItBack(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("Old Friday Night")

	if got := len(l.List()); got != 1 {
		t.Fatalf("List before archive: %d games, want 1", got)
	}
	if got := len(l.ListArchived()); got != 0 {
		t.Fatalf("ListArchived before archive: %d games, want 0", got)
	}

	after, err := l.SetArchived(meta.ID, true)
	if err != nil {
		t.Fatalf("SetArchived: %v", err)
	}
	if !after.Archived() {
		t.Fatal("archived meta reports Archived() false")
	}
	if got := len(l.List()); got != 0 {
		t.Errorf("List after archive: %d games, want 0", got)
	}
	archived := l.ListArchived()
	if len(archived) != 1 || archived[0].ID != meta.ID {
		t.Fatalf("ListArchived after archive: %+v", archived)
	}
	// Nothing was destroyed: Get still answers, and the invite token
	// is intact, so unarchiving restores a usable table rather than
	// an unopenable one.
	got, err := l.Get(meta.ID)
	if err != nil {
		t.Fatalf("Get on an archived game: %v", err)
	}
	if got.InviteToken != meta.InviteToken {
		t.Errorf("invite token lost on archive: %q", got.InviteToken)
	}
	if l.RoomOf(meta.ID) == nil {
		t.Error("archive dropped the room from the manager")
	}

	if _, err := l.SetArchived(meta.ID, false); err != nil {
		t.Fatalf("unarchive: %v", err)
	}
	if got := len(l.List()); got != 1 {
		t.Errorf("List after unarchive: %d games, want 1", got)
	}
	if got := len(l.ListArchived()); got != 0 {
		t.Errorf("ListArchived after unarchive: %d, want 0", got)
	}
}

func TestArchiveIsIdempotentAndUnknownGameIs404(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("FNM")

	first, err := l.SetArchived(meta.ID, true)
	if err != nil {
		t.Fatalf("archive: %v", err)
	}
	second, err := l.SetArchived(meta.ID, true)
	if err != nil {
		t.Fatalf("archive twice: %v", err)
	}
	if !first.ArchivedAt.Equal(*second.ArchivedAt) {
		t.Error("re-archiving moved the archived-at timestamp")
	}
	if _, err := l.SetArchived(uuid.New(), true); err != ErrGameNotFound {
		t.Errorf("archive unknown game: %v, want ErrGameNotFound", err)
	}
}

// A running table's bot runners are goroutines committing moves to a
// room that archiving just hid. Stopping them is the whole reason
// archive is not a pure metadata flip.
func TestArchivingARunningGameStopsItsBotsAndUnarchivingRestartsThem(t *testing.T) {
	l := newTestLobby(t)
	host := newFakeBotHost()
	l.SetBotHost(host)

	meta, _ := l.Create("Bots at the table")
	_, human, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	_, botID, err := l.AddBot(meta.ID, "Bot 1", "random", "", "Mono Red", botDeck(20))
	if err != nil {
		t.Fatalf("AddBot: %v", err)
	}
	if _, err := l.SetDeck(meta.ID, human, "Alice's", botDeck(20)); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}

	if _, err := l.SetArchived(meta.ID, true); err != nil {
		t.Fatalf("archive: %v", err)
	}
	host.mu.Lock()
	stopped := append([]uuid.UUID(nil), host.stopped...)
	host.mu.Unlock()
	if len(stopped) != 1 || stopped[0] != meta.ID {
		t.Fatalf("archive did not stop the bots: %v", stopped)
	}

	// Unarchiving an active table puts the runner back, or the seat
	// is permanently empty and the table hangs at that player's
	// priority.
	host.mu.Lock()
	delete(host.started, meta.ID) // so a re-launch is distinguishable
	host.mu.Unlock()
	if _, err := l.SetArchived(meta.ID, false); err != nil {
		t.Fatalf("unarchive: %v", err)
	}
	host.mu.Lock()
	seats := host.started[meta.ID]
	host.mu.Unlock()
	if len(seats) != 1 || seats[0].PlayerID != botID {
		t.Fatalf("unarchive did not restart the bot runner: %+v", seats)
	}
}

// Archiving must survive a deploy — otherwise the table an operator
// retired reappears on the next restart, bots and all.
func TestArchivedStateSurvivesARestartAndDoesNotRelaunchBots(t *testing.T) {
	dir := t.TempDir()
	log := quietLogger()

	mgr := ws.NewRoomManager(log, dir)
	l := NewLobby(mgr)
	host := newFakeBotHost()
	l.SetBotHost(host)

	meta, _ := l.Create("Retired table")
	_, human, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = l.AddBot(meta.ID, "Bot 1", "random", "", "Mono Red", botDeck(20))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.SetDeck(meta.ID, human, "Alice's", botDeck(20)); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := l.SetArchived(meta.ID, true); err != nil {
		t.Fatal(err)
	}

	// --- the deploy -----------------------------------------------
	mgr2 := ws.NewRoomManager(log, dir)
	l2 := NewLobby(mgr2)
	host2 := newFakeBotHost()
	l2.SetBotHost(host2)
	l2.RestoreFromDisk(log)

	back, err := l2.Get(meta.ID)
	if err != nil {
		t.Fatalf("restored lobby lost the archived game: %v", err)
	}
	if !back.Archived() {
		t.Error("game came back unarchived — the operator's retirement was undone by a deploy")
	}
	if got := len(l2.List()); got != 0 {
		t.Errorf("restored List: %d games, want 0", got)
	}
	host2.mu.Lock()
	started := len(host2.started)
	host2.mu.Unlock()
	if started != 0 {
		t.Error("restore relaunched bot runners for an archived game")
	}
	// The engine snapshot is still on disk — archive removes nothing.
	if _, err := os.Stat(metaPath(dir, meta.ID)); err != nil {
		t.Errorf("lobby metadata missing after archive: %v", err)
	}
}

// --- HTTP surface --------------------------------------------------

func TestArchiveRoutesAreAdminOnly(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	meta, _ := l.Create("FNM")
	_, _, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}

	// A seated player's session is not enough.
	playerTok := playerSession(t, a, meta.ID)
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/archive", playerTok, nil)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("player archive: got %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
	resp.Body.Close()

	// No credential at all.
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/archive", "", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("anonymous archive: got %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	resp.Body.Close()
}

func TestArchiveRoundTripOverHTTP(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	adminTok := adminSession(t, a)
	meta, _ := l.Create("FNM")

	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/archive", adminTok, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("archive: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body GameMeta
	_ = json.NewDecoder(resp.Body).Decode(&body)
	resp.Body.Close()
	if body.ArchivedAt == nil {
		t.Error("archive response carries no archived_at")
	}
	if body.InviteToken != "" || body.SpectatorInvite != "" {
		t.Error("archive response leaked an invite token")
	}

	// Default listing hides it; ?archived=1 shows it.
	if games := listVia(t, srv, adminTok, "/games"); len(games) != 0 {
		t.Errorf("GET /games after archive: %d games, want 0", len(games))
	}
	games := listVia(t, srv, adminTok, "/games?archived=1")
	if len(games) != 1 || games[0].ID != meta.ID || games[0].ArchivedAt == nil {
		t.Fatalf("GET /games?archived=1: %+v", games)
	}

	// DELETE /archive puts it back.
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/games/"+meta.ID.String()+"/archive", nil)
	req.Header.Set("Authorization", "Bearer "+adminTok)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("unarchive: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unarchive: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if games := listVia(t, srv, adminTok, "/games"); len(games) != 1 {
		t.Errorf("GET /games after unarchive: %d games, want 1", len(games))
	}
}

// --- shared helpers ------------------------------------------------

func listVia(t *testing.T, srv *httptest.Server, token, path string) []GameMeta {
	t.Helper()
	resp := doGet(t, srv, path, token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: got %d", path, resp.StatusCode)
	}
	var body listResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	return body.Games
}

// startTwoSeatGame seats Alice and Bob, uploads decks and starts —
// the state seat reclaim exists for.
func startTwoSeatGame(t *testing.T, l *Lobby, name string) (GameMeta, uuid.UUID, uuid.UUID) {
	t.Helper()
	meta, err := l.Create(name)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatalf("Join Alice: %v", err)
	}
	_, bob, err := l.Join(meta.ID, meta.InviteToken, "Bob")
	if err != nil {
		t.Fatalf("Join Bob: %v", err)
	}
	// Big enough that the opening hands do not empty the library —
	// the reclaim end-to-end draws a card to prove the seat can act.
	if _, err := l.SetDeck(meta.ID, alice, "Alice's", botDeck(20)); err != nil {
		t.Fatalf("SetDeck Alice: %v", err)
	}
	if _, err := l.SetDeck(meta.ID, bob, "Bob's", botDeck(20)); err != nil {
		t.Fatalf("SetDeck Bob: %v", err)
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return meta, alice, bob
}
