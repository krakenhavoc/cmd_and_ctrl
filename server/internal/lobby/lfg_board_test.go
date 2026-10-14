package lobby

// lfg_board_test.go: ADR 0161 against an in-memory #lfg channel. The
// worker is never started: each test drives drain and boot itself, on a
// clock it controls, so every step is deterministic.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const lfgBase = "https://c2.test"

// fakeLFG is the #lfg channel: the messages the bot posted, and every
// call it was asked to make.
type fakeLFG struct {
	mu    sync.Mutex
	msgs  map[string]discord.LFGPost
	next  int
	calls []string
	// pings are the posts that pinged, in order.
	pings int
	// fail, when set, is returned by the next calls of that verb.
	fail map[string][]error
}

func newFakeLFG() *fakeLFG {
	return &fakeLFG{msgs: map[string]discord.LFGPost{}, fail: map[string][]error{}}
}

func (f *fakeLFG) failNext(verb string, errs ...error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fail[verb] = append(f.fail[verb], errs...)
}

func (f *fakeLFG) failing(verb string) error {
	if errs := f.fail[verb]; len(errs) > 0 {
		f.fail[verb] = errs[1:]
		return errs[0]
	}
	return nil
}

func (f *fakeLFG) Post(_ context.Context, _ *http.Client, p discord.LFGPost, ping bool) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.failing("post"); err != nil {
		return "", err
	}
	f.next++
	id := fmt.Sprintf("m%d", f.next)
	f.msgs[id] = p
	if ping {
		f.pings++
	}
	f.calls = append(f.calls, fmt.Sprintf("post %s ping=%v", id, ping))
	return id, nil
}

func (f *fakeLFG) Edit(_ context.Context, _ *http.Client, id string, p discord.LFGPost) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.failing("edit"); err != nil {
		return err
	}
	old, ok := f.msgs[id]
	if !ok {
		return discord.ErrLFGMessageGone
	}
	if p.JoinURL == "" {
		// Discord leaves the button as it was.
		p.JoinURL = old.JoinURL
	}
	f.msgs[id] = p
	f.calls = append(f.calls, "edit "+id)
	return nil
}

func (f *fakeLFG) Delete(_ context.Context, _ *http.Client, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.failing("delete"); err != nil {
		return err
	}
	delete(f.msgs, id)
	f.calls = append(f.calls, "delete "+id)
	return nil
}

func (f *fakeLFG) live() map[string]discord.LFGPost {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := map[string]discord.LFGPost{}
	for k, v := range f.msgs {
		out[k] = v
	}
	return out
}

func (f *fakeLFG) log() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.calls...)
}

func (f *fakeLFG) pinged() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.pings
}

// lfgHarness is a lobby with an LFG manager over a fake channel and a
// clock the test moves.
type lfgHarness struct {
	t     *testing.T
	l     *Lobby
	m     *LFGManager
	fake  *fakeLFG
	clock time.Time
}

func newLFGHarness(t *testing.T, l *Lobby) *lfgHarness {
	t.Helper()
	if l == nil {
		l = newTestLobby(t)
	}
	h := &lfgHarness{t: t, l: l, fake: newFakeLFG(), clock: time.Date(2026, 10, 13, 20, 0, 0, 0, time.UTC)}
	h.m = newLFGManager(h.fake, nil, lfgBase, quietLogger(), func() time.Time { return h.clock })
	h.l.SetLFGManager(h.m)
	return h
}

func (h *lfgHarness) drain() { h.m.drain(context.Background()) }

// later moves the clock past the edit window and reconciles every table
// again, as the coalescing timer and the sweep would.
func (h *lfgHarness) later() {
	h.clock = h.clock.Add(lfgEditEvery + time.Second)
	for _, id := range h.l.lfgTableIDs() {
		h.m.notify(id)
	}
	h.drain()
}

func (h *lfgHarness) message(id uuid.UUID) string {
	h.l.mu.Lock()
	defer h.l.mu.Unlock()
	if e, ok := h.l.games[id]; ok {
		return e.lfgMessageID
	}
	return ""
}

func (h *lfgHarness) post(id uuid.UUID) (discord.LFGPost, bool) {
	h.t.Helper()
	msg := h.message(id)
	if msg == "" {
		return discord.LFGPost{}, false
	}
	p, ok := h.fake.live()[msg]
	return p, ok
}

// table makes a lobby table hosted by Alice (Discord 501) and posts it.
func (h *lfgHarness) table(name string) (GameMeta, uuid.UUID) {
	h.t.Helper()
	meta, err := h.l.Create(name)
	if err != nil {
		h.t.Fatal(err)
	}
	_, alice, err := h.l.JoinWithIdentity(meta.ID, meta.InviteToken, "Alice", DiscordIdentity{ID: "501", Username: "alice"})
	if err != nil {
		h.t.Fatal(err)
	}
	if _, err := h.l.SetLFG(meta.ID, true); err != nil {
		h.t.Fatal(err)
	}
	h.drain()
	return meta, alice
}

func TestLFGPostsTheTableAndPingsOnce(t *testing.T) {
	h := newLFGHarness(t, nil)
	meta, _ := h.table("Friday pod")
	p, ok := h.post(meta.ID)
	if !ok {
		t.Fatalf("no post; calls %v", h.fake.log())
	}
	want := inviteURL(lfgBase, meta.ID, meta.InviteToken)
	if p.Table != "Friday pod" || p.HostID != "501" || p.Seated != 1 || p.Seats != game.MaxPlayers || p.JoinURL != want {
		t.Errorf("post = %+v", p)
	}
	if h.fake.pinged() != 1 {
		t.Errorf("pings = %d, want 1", h.fake.pinged())
	}
	if m := mustGet(t, h.l, meta.ID); !m.LFG || m.LFGStatus != LFGStatusPosted {
		t.Errorf("meta lfg %v status %q", m.LFG, m.LFGStatus)
	}
}

func TestLFGEditsFollowTheSeatsAndAreCoalesced(t *testing.T) {
	h := newLFGHarness(t, nil)
	meta, _ := h.table("Friday pod")
	for _, name := range []string{"Gus", "Hal"} {
		if _, _, err := h.l.Join(meta.ID, meta.InviteToken, name); err != nil {
			t.Fatal(err)
		}
	}
	h.drain()
	if p, _ := h.post(meta.ID); p.Seated != 1 {
		t.Errorf("an edit inside the window: seated %d, want still 1", p.Seated)
	}
	h.later()
	p, _ := h.post(meta.ID)
	if p.Seated != 3 || p.Full() {
		t.Errorf("after the window: %+v", p)
	}
	edits := 0
	for _, c := range h.fake.log() {
		if strings.HasPrefix(c, "edit") {
			edits++
		}
	}
	if edits != 1 {
		t.Errorf("edits = %d, want the two joins coalesced into 1 (%v)", edits, h.fake.log())
	}

	// The last seat: Full, and the post stays.
	_, ida, err := h.l.Join(meta.ID, meta.InviteToken, "Ida")
	if err != nil {
		t.Fatal(err)
	}
	h.later()
	if p, ok := h.post(meta.ID); !ok || !p.Full() {
		t.Errorf("a full table's post = %+v, %v", p, ok)
	}
	// A seat frees: open again.
	if _, err := h.l.RemoveSeat(meta.ID, ida); err != nil {
		t.Fatal(err)
	}
	h.later()
	if p, _ := h.post(meta.ID); p.Full() || p.Seated != 3 {
		t.Errorf("after a seat freed: %+v", p)
	}
	if h.fake.pinged() != 1 {
		t.Errorf("an edit pinged: %d pings", h.fake.pinged())
	}
}

func TestLFGFollowsTheInviteAndSettings(t *testing.T) {
	h := newLFGHarness(t, nil)
	meta, alice := h.table("Friday pod")
	tok, _, err := h.l.RotateInvite(meta.ID, InvitePlayer)
	if err != nil {
		t.Fatal(err)
	}
	h.later()
	if p, _ := h.post(meta.ID); p.JoinURL != inviteURL(lfgBase, meta.ID, tok) {
		t.Errorf("after a rotation the button is %q", p.JoinURL)
	}
	life := 30
	if _, err := h.l.UpdateSettings(meta.ID, alice, game.SettingsPatch{StartingLife: &life}); err != nil {
		t.Fatal(err)
	}
	h.later()
	p, _ := h.post(meta.ID)
	if len(p.Settings) != 1 || p.Settings[0] != (discord.LFGField{Name: "Starting life", Value: "30"}) {
		t.Errorf("settings on the post = %+v", p.Settings)
	}
}

func TestLFGPostGoesWhenTheGameStarts(t *testing.T) {
	h := newLFGHarness(t, nil)
	meta, alice := h.table("FNM")
	_, bob, _ := h.l.Join(meta.ID, meta.InviteToken, "Bob")
	for _, p := range []uuid.UUID{alice, bob} {
		uploadDummyDeck(t, h.l, meta.ID, p)
	}
	if _, err := h.l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if live := h.fake.live(); len(live) != 0 || h.message(meta.ID) != "" {
		t.Errorf("after the start: posts %v, column %q", live, h.message(meta.ID))
	}
}

func TestLFGPostGoesWithTheHostsToggleAndTheTable(t *testing.T) {
	h := newLFGHarness(t, nil)
	meta, _ := h.table("Friday pod")
	if _, err := h.l.SetLFG(meta.ID, false); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if len(h.fake.live()) != 0 || mustGet(t, h.l, meta.ID).LFGStatus != "" {
		t.Fatal("withdrawn post stayed")
	}
	if _, err := h.l.SetLFG(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if len(h.fake.live()) != 1 {
		t.Fatal("turning it back on did not post")
	}
	if h.fake.pinged() != 1 {
		t.Errorf("a second post inside the host's cooldown pinged (%d pings)", h.fake.pinged())
	}

	archived, _ := h.table("Archived")
	if _, err := h.l.SetArchived(archived.ID, true); err != nil {
		t.Fatal(err)
	}
	if err := h.l.Delete(meta.ID); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if live := h.fake.live(); len(live) != 0 {
		t.Errorf("posts after archive and delete: %v", live)
	}
}

func TestLFGPingCooldownIsPerHost(t *testing.T) {
	h := newLFGHarness(t, nil)
	h.table("One")
	h.table("Two")
	if h.fake.pinged() != 1 {
		t.Fatalf("two posts by one host inside 30 minutes: %d pings, want 1", h.fake.pinged())
	}
	h.clock = h.clock.Add(lfgPingCooldown)
	h.table("Three")
	if h.fake.pinged() != 2 {
		t.Errorf("after the cooldown: %d pings, want 2", h.fake.pinged())
	}
	// Another host is not held to Alice's cooldown.
	meta, _ := h.l.Create("Bob's")
	if _, _, err := h.l.JoinWithIdentity(meta.ID, meta.InviteToken, "Bob", DiscordIdentity{ID: "502", Username: "bob"}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.l.SetLFG(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if h.fake.pinged() != 3 {
		t.Errorf("a second host's first post: %d pings, want 3", h.fake.pinged())
	}
}

func TestLFGRepostsADeletedPostWithoutAPing(t *testing.T) {
	h := newLFGHarness(t, nil)
	meta, _ := h.table("Friday pod")
	h.clock = h.clock.Add(lfgPingCooldown + time.Minute) // the cooldown is not why
	h.fake.mu.Lock()
	h.fake.msgs = map[string]discord.LFGPost{}
	h.fake.mu.Unlock()
	if _, _, err := h.l.Join(meta.ID, meta.InviteToken, "Gus"); err != nil {
		t.Fatal(err)
	}
	h.later()
	p, ok := h.post(meta.ID)
	if !ok || p.Seated != 2 {
		t.Fatalf("no repost: %v", h.fake.log())
	}
	if h.fake.pinged() != 1 {
		t.Errorf("the repost pinged (%d pings)", h.fake.pinged())
	}
}

func TestLFGFailures(t *testing.T) {
	h := newLFGHarness(t, nil)
	meta, _ := h.l.Create("Friday pod")
	if _, _, err := h.l.JoinWithIdentity(meta.ID, meta.InviteToken, "Alice", DiscordIdentity{ID: "501", Username: "alice"}); err != nil {
		t.Fatal(err)
	}

	// A 403 is not retried: unavailable, the table unaffected.
	h.fake.failNext("post", discord.ErrLFGForbidden)
	if _, err := h.l.SetLFG(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if m := mustGet(t, h.l, meta.ID); m.LFGStatus != LFGStatusUnavailable || !m.LFG {
		t.Fatalf("after a 403: %+v", m.LFGStatus)
	}
	h.later()
	if len(h.fake.live()) != 0 {
		t.Fatal("an unavailable post was retried without the host asking")
	}
	// The host's toggle tries again.
	if _, err := h.l.SetLFG(meta.ID, false); err != nil {
		t.Fatal(err)
	}
	if _, err := h.l.SetLFG(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if len(h.fake.live()) != 1 || mustGet(t, h.l, meta.ID).LFGStatus != LFGStatusPosted {
		t.Fatalf("the toggle did not post again: %v", h.fake.log())
	}

	// Transient failures are retried up to lfgMaxAttempts, then show
	// unavailable.
	if _, _, err := h.l.Join(meta.ID, meta.InviteToken, "Gus"); err != nil {
		t.Fatal(err)
	}
	h.fake.failNext("edit", discord.ErrVoiceUnavailable, discord.ErrVoiceUnavailable, discord.ErrVoiceUnavailable)
	for range lfgMaxAttempts {
		h.later()
	}
	if s := mustGet(t, h.l, meta.ID).LFGStatus; s != LFGStatusUnavailable {
		t.Errorf("after %d transient failures: %q", lfgMaxAttempts, s)
	}
}

func TestLFGTransientFailureRecovers(t *testing.T) {
	h := newLFGHarness(t, nil)
	meta, _ := h.l.Create("Friday pod")
	if _, _, err := h.l.JoinWithIdentity(meta.ID, meta.InviteToken, "Alice", DiscordIdentity{ID: "501", Username: "alice"}); err != nil {
		t.Fatal(err)
	}
	h.fake.failNext("post", &discord.RateLimitError{RetryAfter: time.Second})
	if _, err := h.l.SetLFG(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if s := mustGet(t, h.l, meta.ID).LFGStatus; s != LFGStatusPending {
		t.Fatalf("after a 429: %q, want pending", s)
	}
	h.later()
	if s := mustGet(t, h.l, meta.ID).LFGStatus; s != LFGStatusPosted {
		t.Errorf("after the retry: %q, want posted", s)
	}
}

func TestSetLFGRefusals(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("Friday pod")
	if _, err := l.SetLFG(meta.ID, true); !errors.Is(err, ErrLFGOff) {
		t.Errorf("board off: %v", err)
	}
	h := newLFGHarness(t, l)

	if _, err := l.SetLFG(uuid.New(), true); !errors.Is(err, ErrGameNotFound) {
		t.Errorf("unknown table: %v", err)
	}
	// Nobody seated, nobody named: no host to name.
	if _, err := l.SetLFG(meta.ID, true); !errors.Is(err, ErrLFGGuestHost) {
		t.Errorf("no host: %v", err)
	}
	// A guest host.
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Gus"); err != nil {
		t.Fatal(err)
	}
	if _, err := l.SetLFG(meta.ID, true); !errors.Is(err, ErrLFGGuestHost) {
		t.Errorf("guest host: %v", err)
	}

	full, _ := h.l.Create("Full")
	if _, _, err := l.JoinWithIdentity(full.ID, full.InviteToken, "Alice", DiscordIdentity{ID: "501", Username: "alice"}); err != nil {
		t.Fatal(err)
	}
	for i := 1; i < game.MaxPlayers; i++ {
		if _, _, err := l.Join(full.ID, full.InviteToken, fmt.Sprintf("P%d", i)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := l.SetLFG(full.ID, true); !errors.Is(err, ErrLFGFull) {
		t.Errorf("full table: %v", err)
	}

	archived, _ := h.l.Create("Archived")
	if _, _, err := l.JoinWithIdentity(archived.ID, archived.InviteToken, "Alice", DiscordIdentity{ID: "501", Username: "alice"}); err != nil {
		t.Fatal(err)
	}
	if _, err := l.SetArchived(archived.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := l.SetLFG(archived.ID, true); !errors.Is(err, ErrLFGUnavailable) {
		t.Errorf("archived: %v", err)
	}
	// Turning it off is always allowed.
	if _, err := l.SetLFG(archived.ID, false); err != nil {
		t.Errorf("off on an archived table: %v", err)
	}

	// The invite link lost to a restart, and no post carrying it.
	lost, _ := h.l.Create("Lost link")
	if _, _, err := l.JoinWithIdentity(lost.ID, lost.InviteToken, "Alice", DiscordIdentity{ID: "501", Username: "alice"}); err != nil {
		t.Fatal(err)
	}
	l.mu.Lock()
	l.games[lost.ID].meta.InviteToken = ""
	l.mu.Unlock()
	if _, err := l.SetLFG(lost.ID, true); !errors.Is(err, ErrInviteTokenUnavailable) {
		t.Errorf("no invite link: %v", err)
	}
	if n := len(h.fake.log()); n != 0 {
		t.Errorf("a refused table reached Discord: %v", h.fake.log())
	}
}

func TestLFGSurvivesARestart(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	h := newLFGHarness(t, l)
	kept, _ := h.table("Kept")
	startsNow, alice := h.table("Starts during the restart")
	_, bob, _ := l.Join(startsNow.ID, startsNow.InviteToken, "Bob")
	keptMsg, goneMsg := h.message(kept.ID), h.message(startsNow.ID)
	if keptMsg == "" || goneMsg == "" {
		t.Fatalf("posts before the restart: %q %q", keptMsg, goneMsg)
	}
	// The second table's game starts while the manager is down: its
	// post must go at boot.
	l.mu.Lock()
	l.lfg = nil
	l.mu.Unlock()
	for _, p := range []uuid.UUID{alice, bob} {
		uploadDummyDeck(t, l, startsNow.ID, p)
	}
	if _, err := l.Start(startsNow.ID); err != nil {
		t.Fatal(err)
	}
	// A row whose table never came back: its post goes too.
	orphan, _ := l.Create("Orphan")
	l.setLFGMessage(orphan.ID, "m-orphan")
	h.fake.mu.Lock()
	h.fake.msgs["m-orphan"] = discord.LFGPost{Table: "Orphan"}
	h.fake.mu.Unlock()

	l2, mgr2 := newDurableLobby(t, dir)
	m2 := newLFGManager(h.fake, nil, lfgBase, quietLogger(), func() time.Time { return h.clock })
	l2.SetLFGManager(m2)
	// The orphan has no restore point (it never committed one), so it
	// stays unresident.
	mgr2.Delete(orphan.ID)
	if n := l2.RestoreFromDisk(quietLogger()); n < 2 {
		t.Fatalf("restored %d games, want at least 2", n)
	}
	if m := mustGet(t, l2, kept.ID); !m.LFG {
		t.Error("the host's choice did not survive the restart")
	}
	m2.boot(context.Background())

	live := h.fake.live()
	p, ok := live[keptMsg]
	if !ok || len(live) != 1 {
		t.Fatalf("after the boot: posts %v, want only the kept one", live)
	}
	// The restored lobby has no invite plaintext: the button keeps the
	// link it had, which still works.
	if p.JoinURL != inviteURL(lfgBase, kept.ID, kept.InviteToken) {
		t.Errorf("kept post's link = %q", p.JoinURL)
	}
	if m := mustGet(t, l2, kept.ID); m.LFGStatus != LFGStatusPosted {
		t.Errorf("kept post status %q", m.LFGStatus)
	}
	l2.mu.Lock()
	startedMsg := l2.games[startsNow.ID].lfgMessageID
	l2.mu.Unlock()
	if startedMsg != "" {
		t.Errorf("a started table kept its message id %q", startedMsg)
	}
	ctx, cancel := storeCtx()
	defer cancel()
	if rec, _, err := l2.store.LoadGame(ctx, orphan.ID); err != nil || rec.LFGMessageID != "" {
		t.Errorf("the orphan's column: %q, %v", rec.LFGMessageID, err)
	}
}

func TestPutLFGRoute(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	meta, _ := l.Create("FNM")
	admin := adminSession(t, a)
	path := "/games/" + meta.ID.String() + "/lfg"
	put := func(token string, body any) (int, GameMeta, string) {
		t.Helper()
		resp := doJSON(t, srv, http.MethodPut, path, token, body)
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		var m GameMeta
		_ = json.Unmarshal(raw, &m)
		return resp.StatusCode, m, string(raw)
	}
	on := map[string]bool{"enabled": true}

	if code, _, body := put(admin, on); code != http.StatusServiceUnavailable || !strings.Contains(body, discord.LFGChannelEnv) {
		t.Errorf("board off: %d %s, want a 503 naming the variable", code, body)
	}
	h := newLFGHarness(t, l)
	if code, _, _ := put(admin, map[string]any{}); code != http.StatusBadRequest {
		t.Errorf("no enabled field: %d, want 400", code)
	}
	if code, _, _ := put("", on); code != http.StatusUnauthorized {
		t.Errorf("anonymous: %d, want 401", code)
	}
	// A guest host is refused with the reason.
	gus := joinAs(t, srv, meta, "Gus")
	if code, _, body := put(gus.Token, on); code != http.StatusConflict || !strings.Contains(body, "Discord") {
		t.Errorf("guest host: %d %s", code, body)
	}
	bob := joinAs(t, srv, meta, "Bob")
	if code, _, _ := put(bob.Token, on); code != http.StatusForbidden {
		t.Errorf("non-host: %d, want 403", code)
	}

	// A Discord-hosted table: the admin posts it, the post goes up.
	posted, _ := l.Create("Posted")
	if _, _, err := l.JoinWithIdentity(posted.ID, posted.InviteToken, "Alice", DiscordIdentity{ID: "501", Username: "alice"}); err != nil {
		t.Fatal(err)
	}
	path = "/games/" + posted.ID.String() + "/lfg"
	code, m, _ := put(admin, on)
	if code != http.StatusOK || !m.LFG || m.LFGStatus != LFGStatusPending {
		t.Fatalf("admin posts it: %d lfg=%v status=%q", code, m.LFG, m.LFGStatus)
	}
	h.drain()
	if len(h.fake.live()) != 1 {
		t.Error("no post after PUT")
	}
	if code, m, _ := put(admin, map[string]bool{"enabled": false}); code != http.StatusOK || m.LFG {
		t.Errorf("admin withdraws it: %d %v", code, m.LFG)
	}
}

func TestCreateWithLFG(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	admin := adminSession(t, a)
	create := func(body createGameRequest) GameMeta {
		t.Helper()
		resp := postJSON(t, srv, "/games", admin, body)
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("create: %d %s", resp.StatusCode, raw)
		}
		var m GameMeta
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		return m
	}
	req := createGameRequest{Name: "posted", HostDiscordID: "501", LFG: true}

	// Board off: the ask is ignored, the table is still made.
	if m := create(req); m.LFG || mustGet(t, l, m.ID).LFG {
		t.Error("board off, but the table was posted")
	}
	h := newLFGHarness(t, l)
	m := create(req)
	if !m.LFG || m.LFGStatus != LFGStatusPending {
		t.Fatalf("board on: lfg %v status %q", m.LFG, m.LFGStatus)
	}
	h.drain()
	if p, ok := h.post(m.ID); !ok || p.HostID != "501" || p.Seated != 0 {
		t.Errorf("the created table's post = %+v, %v", p, ok)
	}
	// No host named: nobody to name, so not posted, and still created.
	if m := create(createGameRequest{Name: "nobody", LFG: true}); m.LFG {
		t.Error("a table with no host was posted")
	}
}

func TestCommunityReportsLFG(t *testing.T) {
	srv, l, _ := newTestHTTPStack(t)
	read := func() map[string]any {
		t.Helper()
		resp := doGet(t, srv, "/community", "")
		defer resp.Body.Close()
		var body map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		return body
	}
	if b := read(); b["lfg"] != false {
		t.Errorf("board off: lfg = %v", b["lfg"])
	}
	newLFGHarness(t, l)
	if b := read(); b["lfg"] != true {
		t.Errorf("board on: lfg = %v", b["lfg"])
	}
}
