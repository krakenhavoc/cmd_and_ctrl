package lobby

// voice_room_test.go: ADR 0159's lobby half that does not talk to
// Discord: the host's choice, the voice stamp on the game view, and PUT
// /games/{id}/voice.

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// voiceTable is a lobby table with a Discord seat, a guest and a bot.
func voiceTable(t *testing.T, l *Lobby) (meta GameMeta, discordSeat, guest, bot uuid.UUID) {
	t.Helper()
	meta, err := l.Create("Voice night")
	if err != nil {
		t.Fatal(err)
	}
	_, discordSeat, err = l.JoinWithIdentity(meta.ID, meta.InviteToken, "Alice",
		DiscordIdentity{ID: "501", Username: "alice"})
	if err != nil {
		t.Fatalf("join with identity: %v", err)
	}
	_, guest, err = l.Join(meta.ID, meta.InviteToken, "Gus")
	if err != nil {
		t.Fatal(err)
	}
	_, bot, err = l.AddBot(meta.ID, "Bot", "random", "", "d", botDeck(5))
	if err != nil {
		t.Fatal(err)
	}
	return meta, discordSeat, guest, bot
}

func TestVoiceStampIsNilWithVoiceOff(t *testing.T) {
	l := newTestLobby(t)
	meta, _, _, _ := voiceTable(t, l)
	if v := l.RoomOf(meta.ID).Voice(); v != nil {
		t.Errorf("voice off, stamp = %+v; want nil", v)
	}
	if _, err := l.SetVoiceRoom(meta.ID, true); !errors.Is(err, ErrVoiceRoomsOff) {
		t.Errorf("SetVoiceRoom with voice off: %v", err)
	}
	view, _, err := l.RoomOf(meta.ID).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if view.Voice != nil {
		t.Errorf("game view carries voice with the feature off: %+v", view.Voice)
	}
}

func TestVoiceStampGivesEachSeatItsAccess(t *testing.T) {
	l := newTestLobby(t)
	l.SetVoiceRoomsEnabled(true)
	meta, alice, gus, bot := voiceTable(t, l)

	v := l.RoomOf(meta.ID).Voice()
	if v == nil || v.Enabled || v.Status != protocol.VoiceStatusOff {
		t.Fatalf("before the host asks: %+v; want enabled false, status off", v)
	}
	want := map[string]string{
		alice.String(): protocol.VoiceAccessPending,
		gus.String():   protocol.VoiceAccessNoDiscord,
		bot.String():   protocol.VoiceAccessNoDiscord,
	}
	for id, a := range want {
		if v.Access[id] != a {
			t.Errorf("access[%s] = %q, want %q", id, v.Access[id], a)
		}
	}

	got, err := l.SetVoiceRoom(meta.ID, true)
	if err != nil || !got.VoiceRoom {
		t.Fatalf("SetVoiceRoom: %+v %v", got.VoiceRoom, err)
	}
	v = l.RoomOf(meta.ID).Voice()
	if !v.Enabled || v.Status != protocol.VoiceStatusPending {
		t.Errorf("after the host asks: %+v; want enabled, pending", v)
	}

	// The per-viewer view: Alice sees her own access and no link yet.
	view := l.RoomOf(meta.ID).ViewFor(alice.String())
	if view.Voice == nil || view.Voice.You != protocol.VoiceAccessPending || view.Voice.URL != "" || view.Voice.Access != nil {
		t.Errorf("Alice's view: %+v", view.Voice)
	}
	if view := l.RoomOf(meta.ID).ViewFor(gus.String()); view.Voice.You != protocol.VoiceAccessNoDiscord {
		t.Errorf("Gus's view: %+v", view.Voice)
	}

	if got, err := l.SetVoiceRoom(meta.ID, false); err != nil || got.VoiceRoom {
		t.Fatalf("turning it off: %+v %v", got.VoiceRoom, err)
	}
	if v := l.RoomOf(meta.ID).Voice(); v.Enabled || v.Status != protocol.VoiceStatusOff {
		t.Errorf("after turning it off: %+v", v)
	}
}

func TestVoiceStampFollowsTheSeats(t *testing.T) {
	l := newTestLobby(t)
	l.SetVoiceRoomsEnabled(true)
	meta, _, _, _ := voiceTable(t, l)
	_, bob, err := l.JoinWithIdentity(meta.ID, meta.InviteToken, "Bob", DiscordIdentity{ID: "502", Username: "bob"})
	if err != nil {
		t.Fatal(err)
	}
	if a := l.RoomOf(meta.ID).Voice().Access[bob.String()]; a != protocol.VoiceAccessPending {
		t.Errorf("a Discord seat that just joined: %q, want pending", a)
	}
}

func TestVoiceRoomRefusedForArchivedAndUnknownTables(t *testing.T) {
	l := newTestLobby(t)
	l.SetVoiceRoomsEnabled(true)
	meta, _, _, _ := voiceTable(t, l)
	if _, err := l.SetVoiceRoom(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := l.SetArchived(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := l.SetVoiceRoom(meta.ID, true); !errors.Is(err, ErrVoiceRoomUnavailable) {
		t.Errorf("turning voice on for an archived table: %v", err)
	}
	if _, err := l.SetVoiceRoom(meta.ID, false); err != nil {
		t.Errorf("turning it off is always allowed: %v", err)
	}
	if _, err := l.SetVoiceRoom(uuid.New(), true); !errors.Is(err, ErrGameNotFound) {
		t.Errorf("unknown table: %v", err)
	}
}

func TestVoiceRoomRefusedOnceTheGameHasEnded(t *testing.T) {
	l := newTestLobby(t)
	l.SetVoiceRoomsEnabled(true)
	meta, alice, _, bob, carol := startedTable(t, l)
	if _, err := l.SetVoiceRoom(meta.ID, true); err != nil {
		t.Fatalf("an active table: %v", err)
	}
	// Every human concedes: the bot is the last seat standing.
	for _, p := range []uuid.UUID{alice, bob, carol} {
		concede(t, l, meta.ID, p)
	}
	if m := mustGet(t, l, meta.ID); m.State != "ended" {
		t.Fatalf("state %q after every human conceded, want ended", m.State)
	}
	if _, err := l.SetVoiceRoom(meta.ID, false); err != nil {
		t.Errorf("turning it off after the end: %v", err)
	}
	if _, err := l.SetVoiceRoom(meta.ID, true); !errors.Is(err, ErrVoiceRoomUnavailable) {
		t.Errorf("turning it on after the end: %v", err)
	}
}

func TestVoiceRoomSurvivesARestart(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	l.SetVoiceRoomsEnabled(true)
	meta, alice, _, _ := voiceTable(t, l)
	if _, err := l.SetVoiceRoom(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	// The voice manager's half (slice 3b) writes the live room's id.
	l.mu.Lock()
	l.games[meta.ID].voiceChannelID = "444"
	l.persistGameLocked(l.games[meta.ID])
	l.mu.Unlock()

	l2, _ := newDurableLobby(t, dir)
	l2.SetVoiceRoomsEnabled(true)
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}
	if m := mustGet(t, l2, meta.ID); !m.VoiceRoom {
		t.Error("the host's choice did not survive the restart")
	}
	l2.mu.Lock()
	channel := l2.games[meta.ID].voiceChannelID
	l2.mu.Unlock()
	if channel != "444" {
		t.Errorf("voice_channel_id after the restart = %q, want 444", channel)
	}
	v := l2.RoomOf(meta.ID).Voice()
	if v == nil || !v.Enabled || v.Status != protocol.VoiceStatusPending || v.Access[alice.String()] != protocol.VoiceAccessPending {
		t.Errorf("restored stamp = %+v", v)
	}
}

func TestPutVoiceRoomRoute(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	meta, _ := l.Create("FNM")
	alice := joinAs(t, srv, meta, "Alice")
	bob := joinAs(t, srv, meta, "Bob")
	admin := adminSession(t, a)
	path := "/games/" + meta.ID.String() + "/voice"

	put := func(token string, body any) (int, GameMeta) {
		t.Helper()
		resp := doJSON(t, srv, http.MethodPut, path, token, body)
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		var m GameMeta
		_ = json.Unmarshal(raw, &m)
		return resp.StatusCode, m
	}
	on := map[string]bool{"enabled": true}

	if code, _ := put(alice.Token, on); code != http.StatusServiceUnavailable {
		t.Errorf("voice off: %d, want 503", code)
	}

	l.SetVoiceRoomsEnabled(true)
	if code, _ := put(alice.Token, map[string]any{}); code != http.StatusBadRequest {
		t.Errorf("no enabled field: %d, want 400", code)
	}
	if code, _ := put(bob.Token, on); code != http.StatusForbidden {
		t.Errorf("non-host: %d, want 403", code)
	}
	if code, _ := put("", on); code != http.StatusUnauthorized {
		t.Errorf("anonymous: %d, want 401", code)
	}
	if mustGet(t, l, meta.ID).VoiceRoom {
		t.Fatal("a refused request changed the table")
	}
	code, m := put(alice.Token, on)
	if code != http.StatusOK || !m.VoiceRoom {
		t.Fatalf("host turns it on: %d %+v", code, m.VoiceRoom)
	}
	if code, m := put(admin, map[string]bool{"enabled": false}); code != http.StatusOK || m.VoiceRoom {
		t.Errorf("admin turns it off: %d %+v", code, m.VoiceRoom)
	}

	if _, err := l.SetArchived(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	if code, _ := put(admin, on); code != http.StatusConflict {
		t.Errorf("archived table: %d, want 409", code)
	}
}

func TestCreateWithVoiceRoom(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	admin := adminSession(t, a)
	create := func() GameMeta {
		t.Helper()
		resp := postJSON(t, srv, "/games", admin, createGameRequest{Name: "voiced", VoiceRoom: true})
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

	// Voice off: the ask is ignored, the table is still made.
	if m := create(); m.VoiceRoom || mustGet(t, l, m.ID).VoiceRoom {
		t.Error("voice off, but the table has a voice room")
	}
	l.SetVoiceRoomsEnabled(true)
	m := create()
	if !m.VoiceRoom || !mustGet(t, l, m.ID).VoiceRoom {
		t.Error("voice on, but the create's voice_room was dropped")
	}
}
