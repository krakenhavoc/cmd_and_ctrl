package lobby

// voice_manager_test.go: ADR 0159 §2–§5 against an in-memory Discord.
// The worker is never started: each test drives drain, sweep and boot
// itself, on a clock it controls, so every step is deterministic.

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// fakeDiscord is the community server: who is in it, the rooms in the
// voice category with their overwrites, and who is connected where.
type fakeDiscord struct {
	mu        sync.Mutex
	members   map[string]bool
	rooms     map[string]map[string]bool // channel → granted user ids
	names     map[string]string
	connected map[string]string // user → channel
	muted     map[string]bool
	next      int
	// failCreate, when set, is returned by the next CreateRoom calls.
	failCreate []error
	calls      []string
}

func newFakeDiscord(members ...string) *fakeDiscord {
	f := &fakeDiscord{
		members:   map[string]bool{},
		rooms:     map[string]map[string]bool{},
		names:     map[string]string{},
		connected: map[string]string{},
		muted:     map[string]bool{},
	}
	for _, m := range members {
		f.members[m] = true
	}
	return f
}

func (f *fakeDiscord) log(format string, args ...any) {
	f.calls = append(f.calls, fmt.Sprintf(format, args...))
}

func (f *fakeDiscord) BotUserID(context.Context, *http.Client) (string, error) { return "bot", nil }

func (f *fakeDiscord) CreateRoom(_ context.Context, _ *http.Client, name, botUserID string, members []string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.failCreate) > 0 {
		err := f.failCreate[0]
		f.failCreate = f.failCreate[1:]
		return "", err
	}
	f.next++
	ch := fmt.Sprintf("c%d", f.next)
	f.rooms[ch] = map[string]bool{botUserID: true}
	f.names[ch] = name
	for _, m := range members {
		f.rooms[ch][m] = true
	}
	f.log("create %s %v", ch, members)
	return ch, nil
}

func (f *fakeDiscord) GrantMember(_ context.Context, _ *http.Client, ch, user string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rooms[ch]
	if !ok {
		return discord.ErrVoiceNotFound
	}
	r[user] = true
	f.log("grant %s %s", ch, user)
	return nil
}

func (f *fakeDiscord) RevokeMember(_ context.Context, _ *http.Client, ch, user string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r, ok := f.rooms[ch]; ok {
		delete(r, user)
	}
	f.log("revoke %s %s", ch, user)
	return nil
}

func (f *fakeDiscord) DeleteRoom(_ context.Context, _ *http.Client, ch string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.rooms, ch)
	for u, c := range f.connected {
		if c == ch {
			delete(f.connected, u)
		}
	}
	f.log("delete %s", ch)
	return nil
}

func (f *fakeDiscord) Rooms(context.Context, *http.Client) ([]discord.Room, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []discord.Room
	for ch := range f.rooms {
		out = append(out, discord.Room{ID: ch, Name: f.names[ch]})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (f *fakeDiscord) IsMember(_ context.Context, _ *http.Client, user string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.members[user], nil
}

func (f *fakeDiscord) ConnectedChannel(_ context.Context, _ *http.Client, user string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.connected[user], nil
}

func (f *fakeDiscord) VoiceStateOf(_ context.Context, _ *http.Client, user string) (discord.VoiceState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return discord.VoiceState{ChannelID: f.connected[user], Muted: f.muted[user]}, nil
}

func (f *fakeDiscord) MoveTo(_ context.Context, _ *http.Client, ch, user string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.connected[user] = ch
	f.log("move %s %s", ch, user)
	return nil
}

func (f *fakeDiscord) DisconnectFrom(_ context.Context, _ *http.Client, ch, user string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.connected[user] != ch {
		return false, nil
	}
	delete(f.connected, user)
	f.log("disconnect %s %s", ch, user)
	return true, nil
}

func (f *fakeDiscord) RoomURL(ch string) string { return "https://discord.test/" + ch }

func (f *fakeDiscord) live() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []string
	for ch := range f.rooms {
		out = append(out, ch)
	}
	sort.Strings(out)
	return out
}

func (f *fakeDiscord) granted(ch, user string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.rooms[ch][user]
}

// voiceHarness is a lobby with a voice manager over a fake Discord and
// a clock the test moves.
type voiceHarness struct {
	t     *testing.T
	l     *Lobby
	vm    *VoiceManager
	fake  *fakeDiscord
	clock time.Time
}

func newVoiceHarness(t *testing.T, members ...string) *voiceHarness {
	t.Helper()
	h := &voiceHarness{t: t, fake: newFakeDiscord(members...), clock: time.Date(2026, 10, 13, 20, 0, 0, 0, time.UTC)}
	h.l = newTestLobby(t)
	h.vm = newVoiceManager(h.fake, nil, quietLogger(), func() time.Time { return h.clock })
	h.l.SetVoiceManager(h.vm)
	return h
}

func (h *voiceHarness) drain() { h.vm.drain(context.Background()) }

func (h *voiceHarness) sweep() {
	h.vm.sweep(context.Background())
	h.drain()
}

func (h *voiceHarness) stamp(id uuid.UUID) *protocol.VoiceView {
	h.t.Helper()
	return h.l.RoomOf(id).Voice()
}

func (h *voiceHarness) channel(id uuid.UUID) string {
	h.l.mu.Lock()
	defer h.l.mu.Unlock()
	if e, ok := h.l.games[id]; ok {
		return e.voiceChannelID
	}
	return ""
}

// table makes a voice table with Alice (Discord 501) seated.
func (h *voiceHarness) table() (GameMeta, uuid.UUID) {
	h.t.Helper()
	meta, err := h.l.Create("Friday pod")
	if err != nil {
		h.t.Fatal(err)
	}
	if _, err := h.l.SetVoiceRoom(meta.ID, true); err != nil {
		h.t.Fatal(err)
	}
	_, alice, err := h.l.JoinWithIdentity(meta.ID, meta.InviteToken, "Alice", DiscordIdentity{ID: "501", Username: "alice"})
	if err != nil {
		h.t.Fatal(err)
	}
	h.drain()
	return meta, alice
}

func TestVoiceRoomWaitsForTheFirstDiscordSeat(t *testing.T) {
	h := newVoiceHarness(t, "501")
	meta, _ := h.l.Create("Friday pod")
	if _, err := h.l.SetVoiceRoom(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, _, err := h.l.Join(meta.ID, meta.InviteToken, "Gus"); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if live := h.fake.live(); len(live) != 0 {
		t.Fatalf("a room with no Discord seat: %v", live)
	}
	if s := h.stamp(meta.ID); s.Status != protocol.VoiceStatusPending {
		t.Errorf("status %q, want pending", s.Status)
	}

	_, alice, err := h.l.JoinWithIdentity(meta.ID, meta.InviteToken, "Alice", DiscordIdentity{ID: "501", Username: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	h.drain()
	live := h.fake.live()
	if len(live) != 1 || !h.fake.granted(live[0], "501") {
		t.Fatalf("after Alice sits down: rooms %v", live)
	}
	if h.channel(meta.ID) != live[0] {
		t.Errorf("voice_channel_id = %q, want %q", h.channel(meta.ID), live[0])
	}
	s := h.stamp(meta.ID)
	if s.Status != protocol.VoiceStatusReady || s.URL != "https://discord.test/"+live[0] || s.Access[alice.String()] != protocol.VoiceAccessGranted {
		t.Errorf("stamp = %+v", s)
	}
	if v := h.l.RoomOf(meta.ID).ViewFor(alice.String()).Voice; v.URL == "" || v.You != protocol.VoiceAccessGranted {
		t.Errorf("Alice's view = %+v", v)
	}
}

func TestVoiceRoomLetsInOnlyServerMembers(t *testing.T) {
	h := newVoiceHarness(t, "501")
	meta, _ := h.table()
	_, bob, err := h.l.JoinWithIdentity(meta.ID, meta.InviteToken, "Bob", DiscordIdentity{ID: "502", Username: "bob"})
	if err != nil {
		t.Fatal(err)
	}
	h.drain()
	ch := h.channel(meta.ID)
	if h.fake.granted(ch, "502") {
		t.Fatal("Bob is not in the server but was let in")
	}
	if a := h.stamp(meta.ID).Access[bob.String()]; a != protocol.VoiceAccessNotInServer {
		t.Errorf("Bob's access %q, want not_in_server", a)
	}
	if v := h.l.RoomOf(meta.ID).ViewFor(bob.String()).Voice; v.URL != "" {
		t.Error("a player not let in got the link")
	}

	// Bob joins the server and asks again.
	h.fake.mu.Lock()
	h.fake.members["502"] = true
	h.fake.mu.Unlock()
	if err := h.l.RefreshVoice(meta.ID, bob); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if !h.fake.granted(ch, "502") || h.stamp(meta.ID).Access[bob.String()] != protocol.VoiceAccessGranted {
		t.Error("Bob was not let in after joining the server")
	}
}

func TestVoiceRoomFollowsASeatThatLeaves(t *testing.T) {
	h := newVoiceHarness(t, "501", "502")
	meta, alice := h.table()
	_, bob, _ := h.l.JoinWithIdentity(meta.ID, meta.InviteToken, "Bob", DiscordIdentity{ID: "502", Username: "bob"})
	h.drain()
	ch := h.channel(meta.ID)
	h.fake.mu.Lock()
	h.fake.connected["502"] = ch
	h.fake.connected["501"] = "elsewhere"
	h.fake.mu.Unlock()

	if _, err := h.l.RemoveSeat(meta.ID, bob); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if h.fake.granted(ch, "502") {
		t.Error("Bob left but keeps the overwrite")
	}
	h.fake.mu.Lock()
	bobIn, aliceWhere := h.fake.connected["502"], h.fake.connected["501"]
	h.fake.mu.Unlock()
	if bobIn != "" {
		t.Error("Bob left the table but is still in its room")
	}
	if aliceWhere != "elsewhere" {
		t.Error("a player in another channel was moved")
	}
	if !h.fake.granted(ch, "501") || h.stamp(meta.ID).Access[alice.String()] != protocol.VoiceAccessGranted {
		t.Error("Alice lost her access")
	}
}

func TestVoiceRoomGoesWhenTheHostTurnsItOff(t *testing.T) {
	h := newVoiceHarness(t, "501")
	meta, _ := h.table()
	if len(h.fake.live()) != 1 {
		t.Fatal("no room")
	}
	if _, err := h.l.SetVoiceRoom(meta.ID, false); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if live := h.fake.live(); len(live) != 0 {
		t.Errorf("rooms after turning voice off: %v", live)
	}
	if h.channel(meta.ID) != "" || h.stamp(meta.ID).Status != protocol.VoiceStatusOff {
		t.Errorf("channel %q, status %q", h.channel(meta.ID), h.stamp(meta.ID).Status)
	}
	if _, err := h.l.SetVoiceRoom(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if len(h.fake.live()) != 1 || h.stamp(meta.ID).Status != protocol.VoiceStatusReady {
		t.Error("turning it back on made no room")
	}
}

func TestVoiceRoomOutlivesTheGameByTheGrace(t *testing.T) {
	h := newVoiceHarness(t, "501")
	meta, _ := h.l.Create("FNM")
	if _, err := h.l.SetVoiceRoom(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	_, alice, _ := h.l.JoinWithIdentity(meta.ID, meta.InviteToken, "Alice", DiscordIdentity{ID: "501", Username: "alice"})
	_, bob, _ := h.l.Join(meta.ID, meta.InviteToken, "Bob")
	for _, p := range []uuid.UUID{alice, bob} {
		uploadDummyDeck(t, h.l, meta.ID, p)
	}
	if _, err := h.l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if len(h.fake.live()) != 1 {
		t.Fatal("no room for the active game")
	}
	concede(t, h.l, meta.ID, bob)
	if m := mustGet(t, h.l, meta.ID); m.State != "ended" {
		t.Fatalf("state %q, want ended", m.State)
	}
	h.l.mu.Lock()
	ended := *h.l.games[meta.ID].endedAt
	h.l.mu.Unlock()

	h.clock = ended.Add(voiceEndGrace - time.Minute)
	h.sweep()
	if len(h.fake.live()) != 1 {
		t.Fatal("the room went before the grace ran out")
	}
	h.clock = ended.Add(voiceEndGrace + time.Second)
	h.sweep()
	if live := h.fake.live(); len(live) != 0 {
		t.Errorf("rooms after the grace: %v", live)
	}
	if s := h.stamp(meta.ID); s.Status != protocol.VoiceStatusOff {
		t.Errorf("status %q after the grace, want off", s.Status)
	}
}

func TestVoiceRoomIdleRuleAndReopen(t *testing.T) {
	h := newVoiceHarness(t, "501")
	meta, alice := h.table()
	ch := h.channel(meta.ID)

	// Someone is in the room: two quiet hours do not remove it.
	h.fake.mu.Lock()
	h.fake.connected["501"] = ch
	h.fake.mu.Unlock()
	h.clock = h.clock.Add(voiceIdleAfter + time.Minute)
	h.sweep()
	if len(h.fake.live()) != 1 {
		t.Fatal("a room someone is in was removed")
	}

	// Nobody in it, two more quiet hours: removed, and the table says idle.
	h.fake.mu.Lock()
	delete(h.fake.connected, "501")
	h.fake.mu.Unlock()
	h.clock = h.clock.Add(voiceIdleAfter + time.Minute)
	h.sweep()
	if live := h.fake.live(); len(live) != 0 {
		t.Fatalf("an idle room stayed: %v", live)
	}
	if s := h.stamp(meta.ID); s.Status != protocol.VoiceStatusIdle || !s.Enabled {
		t.Errorf("stamp after idling = %+v", s)
	}

	// Alice asks for it back.
	if err := h.l.RefreshVoice(meta.ID, alice); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if len(h.fake.live()) != 1 || h.stamp(meta.ID).Status != protocol.VoiceStatusReady {
		t.Error("reopening made no room")
	}
}

func TestVoiceRoomGameActivityKeepsItAlive(t *testing.T) {
	h := newVoiceHarness(t, "501")
	meta, alice := h.table()
	// An action on the table, then most of the idle window later.
	h.clock = h.clock.Add(voiceIdleAfter - time.Minute)
	room := h.l.RoomOf(meta.ID)
	if _, _, err := room.ApplyExternal(func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	h.sweep()
	h.clock = h.clock.Add(voiceIdleAfter - time.Minute)
	h.sweep()
	if len(h.fake.live()) != 1 {
		t.Error("a table in play lost its room")
	}
	_ = alice
}

func TestVoiceRoomCap(t *testing.T) {
	old := voiceRoomCap
	voiceRoomCap = 1
	defer func() { voiceRoomCap = old }()
	h := newVoiceHarness(t, "501", "502")
	first, _ := h.table()
	second, _ := h.l.Create("Second pod")
	if _, err := h.l.SetVoiceRoom(second.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, _, err := h.l.JoinWithIdentity(second.ID, second.InviteToken, "Bob", DiscordIdentity{ID: "502", Username: "bob"}); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if s := h.stamp(second.ID); s.Status != protocol.VoiceStatusFull {
		t.Errorf("second table status %q, want full", s.Status)
	}
	// The first room goes; the second table's next change gets one.
	if _, err := h.l.SetVoiceRoom(first.ID, false); err != nil {
		t.Fatal(err)
	}
	h.drain()
	h.vm.notify(second.ID)
	h.drain()
	if s := h.stamp(second.ID); s.Status != protocol.VoiceStatusReady {
		t.Errorf("second table after a room freed up: %q", s.Status)
	}
}

func TestVoiceRoomFailures(t *testing.T) {
	h := newVoiceHarness(t, "501")
	h.fake.failCreate = []error{discord.ErrVoiceUnavailable, discord.ErrVoiceUnavailable, discord.ErrVoiceUnavailable}
	meta, _ := h.table()
	if s := h.stamp(meta.ID); s.Status != protocol.VoiceStatusPending {
		t.Fatalf("after one transient failure: %q, want pending (a retry is due)", s.Status)
	}
	for i := 0; i < 2; i++ {
		h.vm.notify(meta.ID)
		h.drain()
	}
	if s := h.stamp(meta.ID); s.Status != protocol.VoiceStatusUnavailable {
		t.Errorf("after %d transient failures: %q, want unavailable", voiceMaxAttempts, s.Status)
	}

	// A permanent failure is unavailable at once, and the host's toggle
	// starts over.
	h2 := newVoiceHarness(t, "501")
	h2.fake.failCreate = []error{discord.ErrVoiceForbidden}
	meta2, _ := h2.table()
	if s := h2.stamp(meta2.ID); s.Status != protocol.VoiceStatusUnavailable {
		t.Errorf("forbidden: %q, want unavailable", s.Status)
	}
	if _, err := h2.l.SetVoiceRoom(meta2.ID, false); err != nil {
		t.Fatal(err)
	}
	h2.drain()
	if _, err := h2.l.SetVoiceRoom(meta2.ID, true); err != nil {
		t.Fatal(err)
	}
	h2.drain()
	if s := h2.stamp(meta2.ID); s.Status != protocol.VoiceStatusReady {
		t.Errorf("after the host's toggle: %q, want ready", s.Status)
	}
}

func TestVoiceRoomGoesWithADeletedTable(t *testing.T) {
	h := newVoiceHarness(t, "501")
	meta, _ := h.table()
	if err := h.l.Delete(meta.ID); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if live := h.fake.live(); len(live) != 0 {
		t.Errorf("rooms after deleting the table: %v", live)
	}
}

func TestVoiceBootKeepsClaimedRoomsAndDeletesStrays(t *testing.T) {
	h := newVoiceHarness(t, "501")
	meta, _ := h.table()
	kept := h.channel(meta.ID)
	// A stray from an older process, and a table whose room vanished.
	h.fake.mu.Lock()
	h.fake.rooms["stray"] = map[string]bool{}
	h.fake.mu.Unlock()
	gone, _ := h.l.Create("Vanished")
	if _, err := h.l.SetVoiceRoom(gone.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, _, err := h.l.JoinWithIdentity(gone.ID, gone.InviteToken, "Alice2", DiscordIdentity{ID: "501", Username: "alice"}); err == nil {
		// Same Discord user at a second table is fine for this test.
		_ = err
	}
	h.l.setVoiceChannel(gone.ID, "deleted-on-discord")

	// A fresh manager over the same lobby and Discord: the restart.
	vm2 := newVoiceManager(h.fake, nil, quietLogger(), func() time.Time { return h.clock })
	h.l.SetVoiceManager(vm2)
	vm2.boot(context.Background())

	live := h.fake.live()
	for _, ch := range live {
		if ch == "stray" {
			t.Error("the stray room survived the boot")
		}
	}
	if !contains(live, kept) || h.channel(meta.ID) != kept {
		t.Errorf("the claimed room was not kept: live %v, channel %q", live, h.channel(meta.ID))
	}
	if ch := h.channel(gone.ID); ch == "deleted-on-discord" || ch == "" || !contains(live, ch) {
		t.Errorf("the vanished room was not replaced: channel %q, live %v", ch, live)
	}
	if len(live) != 2 {
		t.Errorf("live rooms %v, want the kept one and a replacement", live)
	}
}

func TestRefreshVoiceRoute(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	meta, _ := l.Create("FNM")
	guest := joinAs(t, srv, meta, "Gus")
	_, alice, err := l.JoinWithIdentity(meta.ID, meta.InviteToken, "Alice", DiscordIdentity{ID: "501", Username: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	aliceTok, _, err := a.Issue(context.Background(),
		auth.Principal{Role: auth.RolePlayer, GameID: meta.ID, PlayerID: alice, DiscordID: "501", Name: "Alice"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	other, _ := l.Create("Other")
	stranger := joinAs(t, srv, other, "Dana")
	path := "/games/" + meta.ID.String() + "/voice/refresh"
	post := func(tok string) int {
		t.Helper()
		resp := postJSON(t, srv, path, tok, map[string]any{})
		defer resp.Body.Close()
		return resp.StatusCode
	}

	if got := post(aliceTok); got != http.StatusServiceUnavailable {
		t.Errorf("voice off: %d, want 503", got)
	}
	fake := newFakeDiscord("501")
	vm := newVoiceManager(fake, nil, quietLogger(), time.Now)
	l.SetVoiceManager(vm)
	if got := post(aliceTok); got != http.StatusConflict {
		t.Errorf("table without voice: %d, want 409", got)
	}
	if _, err := l.SetVoiceRoom(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	if got := post(guest.Token); got != http.StatusConflict {
		t.Errorf("guest seat: %d, want 409", got)
	}
	if got := post(stranger.Token); got != http.StatusForbidden {
		t.Errorf("a seat at another table: %d, want 403", got)
	}
	if got := post(""); got != http.StatusUnauthorized {
		t.Errorf("anonymous: %d, want 401", got)
	}
	if got := post(aliceTok); got != http.StatusAccepted {
		t.Errorf("Alice: %d, want 202", got)
	}
	vm.drain(context.Background())
	if len(fake.live()) != 1 {
		t.Error("the refresh made no room")
	}
}

func contains(xs []string, x string) bool {
	for _, s := range xs {
		if s == x {
			return true
		}
	}
	return false
}

func (h *voiceHarness) poll() { h.vm.pollConnections(context.Background()) }

func TestVoiceRoomShowsWhoIsConnected(t *testing.T) {
	h := newVoiceHarness(t, "501")
	meta, alice := h.table()
	room := h.fake.live()[0]

	if s := h.stamp(meta.ID); s.Granted != 1 || s.InRoom != 0 || len(s.Connected) != 0 {
		t.Fatalf("before anyone connects: %+v", s)
	}
	h.fake.connected["501"] = room
	h.poll()
	s := h.stamp(meta.ID)
	if s.InRoom != 1 || s.Connected[alice.String()] != protocol.VoiceConnLive {
		t.Fatalf("connected: %+v", s)
	}
	h.fake.muted["501"] = true
	h.poll()
	if s := h.stamp(meta.ID); s.Connected[alice.String()] != protocol.VoiceConnMuted {
		t.Fatalf("muted: %+v", s)
	}
	// A seated viewer sees it; a spectator does not.
	if v := h.l.RoomOf(meta.ID).ViewFor(alice.String()).Voice; v.InRoom != 1 || v.Connected[alice.String()] != protocol.VoiceConnMuted {
		t.Errorf("Alice's view: %+v", v)
	}
	if v := h.l.RoomOf(meta.ID).ViewFor(protocol.SpectatorViewerID).Voice; v.InRoom != 0 || v.Connected != nil {
		t.Errorf("spectator's view: %+v", v)
	}
	delete(h.fake.connected, "501")
	h.poll()
	if s := h.stamp(meta.ID); s.InRoom != 0 || len(s.Connected) != 0 {
		t.Fatalf("left: %+v", s)
	}
}

func TestVoiceRoomPullsInOnlyWhenOnAndNeverFromTheRoom(t *testing.T) {
	h := newVoiceHarness(t, "501")
	_, _ = h.table()
	room := h.fake.live()[0]

	h.fake.connected["501"] = "lounge"
	h.poll()
	if h.fake.connected["501"] != "lounge" {
		t.Fatal("moved with pull-in off")
	}
	h.vm.pullIn = true
	h.poll()
	if h.fake.connected["501"] != room {
		t.Fatalf("not pulled in: %q", h.fake.connected["501"])
	}
	n := len(h.fake.calls)
	h.poll()
	if len(h.fake.calls) != n {
		t.Errorf("moved someone already in the room: %v", h.fake.calls[n:])
	}
}
