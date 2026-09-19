package lobby

// settings_test.go — ADR 0075 §2.3, sub-PR 3. PATCH
// /games/{id}/settings: the gate, the partial-patch semantics, and
// the two ways a patch is refused.

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func patchJSON(t *testing.T, srv *httptest.Server, path, token string, body any) *http.Response {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req, err := http.NewRequest(http.MethodPatch, srv.URL+path, bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	return resp
}

// seatSession issues a player session for a seat that actually exists
// in the game, which is what CanManageTable compares against. The
// join route would do this too, but going through the lobby directly
// keeps the setup to the three lines the gate is about.
func seatSession(t *testing.T, a auth.Authenticator, gameID, playerID uuid.UUID, name string) string {
	t.Helper()
	tok, _, err := a.Issue(context.Background(), auth.Principal{
		Role: auth.RolePlayer, GameID: gameID, PlayerID: playerID, Name: name,
	}, time.Hour)
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	return tok
}

// settingsTable seats Alice (who therefore hosts) and Bob, and hands
// back a session for each.
func settingsTable(t *testing.T, l *Lobby, a auth.Authenticator) (GameMeta, uuid.UUID, string, string) {
	t.Helper()
	meta, err := l.Create("FNM")
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
	return meta, alice,
		seatSession(t, a, meta.ID, alice, "Alice"),
		seatSession(t, a, meta.ID, bob, "Bob")
}

func settingsOf(t *testing.T, l *Lobby, id uuid.UUID) game.TableSettings {
	t.Helper()
	return l.RoomOf(id).Game.TableSettingsSnapshot()
}

// decodeSettings reads the route's answer, which is the WHOLE settings
// struct after the patch — not an echo of the patch.
func decodeSettings(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	var body struct {
		Settings map[string]any `json:"settings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body.Settings
}

func TestPatchSettingsHostCanChangeTheTable(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	meta, _, aliceTok, _ := settingsTable(t, l, a)

	resp := patchJSON(t, srv, "/games/"+meta.ID.String()+"/settings", aliceTok,
		map[string]any{"undo_limit": 3, "allow_spawn": true})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("host patch: got %d, want 200", resp.StatusCode)
	}
	body := decodeSettings(t, resp)
	if body["undo_limit"] != float64(3) || body["allow_spawn"] != true {
		t.Errorf("response settings = %v", body)
	}
	// The whole struct comes back, so the client can render the panel
	// from the answer instead of waiting for the broadcast.
	if body["starting_life"] != float64(game.StartingLife) {
		t.Errorf("response omits a field the patch did not name: %v", body)
	}

	s := settingsOf(t, l, meta.ID)
	if s.UndoLimit != 3 || !s.AllowSpawn {
		t.Errorf("engine settings = %+v", s)
	}
	if s.StartingLife != game.StartingLife || s.CommanderDamage != game.CommanderDamageLethal {
		t.Errorf("a field the patch did not name moved: %+v", s)
	}
}

func TestPatchSettingsRefusesANonHostSeat(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	meta, _, _, bobTok := settingsTable(t, l, a)
	before := settingsOf(t, l, meta.ID)

	resp := patchJSON(t, srv, "/games/"+meta.ID.String()+"/settings", bobTok,
		map[string]any{"undo_limit": 9})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("non-host patch: got %d, want 403", resp.StatusCode)
	}
	if after := settingsOf(t, l, meta.ID); after != before {
		t.Errorf("a refused patch changed the table: %+v -> %+v", before, after)
	}
}

func TestPatchSettingsRefusesSpectatorsAndStrangers(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	meta, alice, _, _ := settingsTable(t, l, a)
	before := settingsOf(t, l, meta.ID)

	spectator, _, err := a.Issue(context.Background(), auth.Principal{
		Role: auth.RoleSpectator, GameID: meta.ID, Name: "Watcher",
	}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	// An identified (signed-in, unseated) session — ADR 0051's
	// Discord sign-in before the invite. Identity is not a seat.
	identified, _, err := a.Issue(context.Background(), auth.Principal{
		Role: auth.RoleIdentified, DiscordID: "1111", Name: "Luke",
	}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	// The host of this table, holding a session bound to ANOTHER
	// game. Hosting is per table.
	elsewhere := seatSession(t, a, uuid.New(), alice, "Alice")

	for name, tok := range map[string]string{
		"spectator":         spectator,
		"identified":        identified,
		"host of elsewhere": elsewhere,
	} {
		resp := patchJSON(t, srv, "/games/"+meta.ID.String()+"/settings", tok,
			map[string]any{"allow_spawn": true})
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("%s: got %d, want 403", name, resp.StatusCode)
		}
		resp.Body.Close()
	}
	if after := settingsOf(t, l, meta.ID); after != before {
		t.Errorf("a refused patch changed the table: %+v -> %+v", before, after)
	}
}

func TestPatchSettingsAdminNeedsNoSeat(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	meta, _, _, _ := settingsTable(t, l, a)

	resp := patchJSON(t, srv, "/games/"+meta.ID.String()+"/settings", adminSession(t, a),
		map[string]any{"bot_pace": "slow"})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("admin patch: got %d, want 200", resp.StatusCode)
	}
	if s := settingsOf(t, l, meta.ID); s.BotPace != game.BotPaceSlow {
		t.Errorf("BotPace = %q, want slow", s.BotPace)
	}
}

// TestPatchSettingsRangeErrorsAre400 — a value outside its range is a
// malformed request, and the whole patch is refused: the good field
// beside the bad one does not land.
func TestPatchSettingsRangeErrorsAre400(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	meta, _, aliceTok, _ := settingsTable(t, l, a)

	for name, patch := range map[string]map[string]any{
		"undo below unlimited": {"undo_limit": -2},
		"starting life 0":      {"starting_life": 0, "allow_spawn": true},
		"commander damage 500": {"commander_damage": 500},
		"unknown bot pace":     {"bot_pace": "glacial"},
		"unknown undo scope":   {"undo_scope": "anyone"},
	} {
		resp := patchJSON(t, srv, "/games/"+meta.ID.String()+"/settings", aliceTok, patch)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s: got %d, want 400", name, resp.StatusCode)
		}
		resp.Body.Close()
	}
	if s := settingsOf(t, l, meta.ID); s != game.DefaultTableSettings() {
		t.Errorf("a refused patch moved something: %+v", s)
	}
}

// TestPatchSettingsStartingLifeAfterStartIs422 — ADR 0075 §2.3's
// documented rejection. In the LOBBY the same change is fine and
// rewrites the seats' life totals, which is the contrast worth
// pinning: the field is not read-only, it is locked by state.
func TestPatchSettingsStartingLifeAfterStartIs422(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	meta, _, aliceTok, _ := settingsTable(t, l, a)
	path := "/games/" + meta.ID.String() + "/settings"

	resp := patchJSON(t, srv, path, aliceTok, map[string]any{"starting_life": 30})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("starting life in the lobby: got %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()
	if s := settingsOf(t, l, meta.ID); s.StartingLife != 30 {
		t.Fatalf("StartingLife = %d, want 30", s.StartingLife)
	}

	for _, p := range mustGet(t, l, meta.ID).Players {
		uploadDummyDeck(t, l, meta.ID, p.PlayerID)
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Now locked. The other field in the same patch is refused with
	// it — UpdateSettings state-checks the whole patch before it
	// writes any of it.
	resp = patchJSON(t, srv, path, aliceTok,
		map[string]any{"starting_life": 60, "commander_damage": 30})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("starting life after Start: got %d, want 422", resp.StatusCode)
	}
	s := settingsOf(t, l, meta.ID)
	if s.StartingLife != 30 {
		t.Errorf("StartingLife = %d, want 30", s.StartingLife)
	}
	if s.CommanderDamage != game.CommanderDamageLethal {
		t.Errorf("the other field in the refused patch landed: CommanderDamage = %d", s.CommanderDamage)
	}

	// …and a patch that leaves starting life out is still fine on a
	// live table: ADR 0075 §2.3 allows a mid-game change.
	resp2 := patchJSON(t, srv, path, aliceTok, map[string]any{"commander_damage": 30})
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("mid-game patch: got %d, want 200", resp2.StatusCode)
	}
	if s := settingsOf(t, l, meta.ID); s.CommanderDamage != 30 {
		t.Errorf("CommanderDamage = %d, want 30", s.CommanderDamage)
	}
}

// TestPatchSettingsUnknownGameIs404 keeps the route's shape honest:
// the gate runs AFTER the lookup, so a stranger probing for a game ID
// learns the same thing they would from any other route.
func TestPatchSettingsUnknownGameIs404(t *testing.T) {
	srv, _, a := newTestHTTPStack(t)
	resp := patchJSON(t, srv, "/games/"+uuid.New().String()+"/settings",
		adminSession(t, a), map[string]any{"allow_spawn": true})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("unknown game: got %d, want 404", resp.StatusCode)
	}
}
