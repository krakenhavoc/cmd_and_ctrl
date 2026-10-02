package lobby

// Tests for ADR 0110 section 5 items 1 and 2 (Delivery PR 7): the last
// setup is captured when a table starts, served by GET /me/setup, and
// applied by POST /games/{id}/setup and POST /games's "setup".

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/tablesetups"
)

func (s *tableStack) getSetup(t *testing.T, tok string) (int, mySetupResponse) {
	t.Helper()
	resp := doGet(t, s.srv, "/me/setup", tok)
	defer resp.Body.Close()
	var out mySetupResponse
	if resp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			t.Fatalf("decode /me/setup: %v", err)
		}
	}
	return resp.StatusCode, out
}

func (s *tableStack) uploadTestDeck(t *testing.T, seat sessionResponse, gameID uuid.UUID) {
	t.Helper()
	resp := postJSON(t, s.srv, "/games/"+gameID.String()+"/decks", seat.Token,
		uploadDeckRequest{Format: "text", Source: testDeckSource, PlayerID: seat.PlayerID})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("upload: %d %s", resp.StatusCode, body)
	}
}

func intp(v int) *int { return &v }

func TestSetupIsCapturedWhenTheTableStarts(t *testing.T) {
	s := newTableStack(t)
	alice, aliceID := s.signIn(t, "discord-99", "Alice")
	bob, bobID := s.signIn(t, "discord-77", "Bob")

	// Nothing yet: a null setup, not a 404.
	if code, got := s.getSetup(t, alice); code != http.StatusOK || got.Setup != nil {
		t.Fatalf("before any table: %d %+v", code, got)
	}

	table := s.mustCreate(t, alice, "Friday")
	aliceSeat := s.join(t, alice, table.GameMeta, "")
	bobSeat := s.join(t, bob, table.GameMeta, "")
	guestSeat := s.join(t, "", table.GameMeta, "Guest")
	resp := postJSON(t, s.srv, "/games/"+table.ID.String()+"/seats/bot", aliceSeat.Token,
		addBotRequest{Tier: "random", Deck: "test-mono-white", Name: "Robo"})
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("add bot: %d %s", resp.StatusCode, body)
	}
	resp.Body.Close()
	if got := status(t, s.srv, "PATCH", "/games/"+table.ID.String()+"/settings", aliceSeat.Token,
		game.SettingsPatch{StartingLife: intp(30)}); got != http.StatusOK {
		t.Fatalf("patch settings: %d", got)
	}
	for _, seat := range []sessionResponse{aliceSeat, bobSeat, guestSeat} {
		s.uploadTestDeck(t, seat, table.ID)
	}

	// No setup is captured before the start.
	if _, got := s.getSetup(t, alice); got.Setup != nil {
		t.Fatalf("a setup was captured before the start: %+v", got.Setup)
	}

	// Bob presses start; the setup is the creator's, Alice's.
	if got := status(t, s.srv, "POST", "/games/"+table.ID.String()+"/start", bobSeat.Token, nil); got != http.StatusOK {
		t.Fatalf("start: %d", got)
	}
	code, got := s.getSetup(t, alice)
	if code != http.StatusOK || got.Setup == nil {
		t.Fatalf("after the start: %d %+v", code, got)
	}
	if got.GameID == nil || *got.GameID != table.ID || got.UpdatedAt <= 0 {
		t.Errorf("game_id / updated_at: %+v", got)
	}
	var patch game.SettingsPatch
	if err := json.Unmarshal(got.Setup.Settings, &patch); err != nil {
		t.Fatalf("settings: %v (%s)", err, got.Setup.Settings)
	}
	def := game.DefaultTableSettings()
	if patch.StartingLife == nil || *patch.StartingLife != 30 ||
		patch.UndoLimit == nil || *patch.UndoLimit != def.UndoLimit ||
		patch.CommanderDamage == nil || patch.BotPace == nil || patch.UndoScope == nil || patch.AllowSpawn == nil {
		t.Errorf("the settings are not a complete patch: %s", got.Setup.Settings)
	}
	if len(got.Setup.Bots) != 1 || got.Setup.Bots[0] != (tablesetups.Bot{Tier: "random", DeckID: "test-mono-white", Name: "Robo"}) {
		t.Errorf("bots: %+v", got.Setup.Bots)
	}
	// The other signed-in human; not Alice herself, not the guest.
	if len(got.Setup.Tablemates) != 1 || got.Setup.Tablemates[0] != bobID {
		t.Errorf("tablemates: %v, want [%s] (alice is %s)", got.Setup.Tablemates, bobID, aliceID)
	}
	// Bob did not create it, so Bob has no setup.
	if _, bobs := s.getSetup(t, bob); bobs.Setup != nil {
		t.Errorf("the starter of someone else's table got a setup: %+v", bobs.Setup)
	}
}

// TestSetupOfACreatorlessTableIsTheStarters: a token-created table has
// no creator, so the person who pressed start owns its setup.
func TestSetupOfACreatorlessTableIsTheStarters(t *testing.T) {
	s := newTableStack(t)
	alice, _ := s.signIn(t, "discord-99", "Alice")
	table := s.mustCreate(t, s.admin(t), "bot-made")
	aliceSeat := s.join(t, alice, table.GameMeta, "")
	guest := s.join(t, "", table.GameMeta, "Guest")
	s.uploadTestDeck(t, aliceSeat, table.ID)
	s.uploadTestDeck(t, guest, table.ID)
	if got := status(t, s.srv, "POST", "/games/"+table.ID.String()+"/start", aliceSeat.Token, nil); got != http.StatusOK {
		t.Fatalf("start: %d", got)
	}
	if _, got := s.getSetup(t, alice); got.Setup == nil || len(got.Setup.Tablemates) != 0 {
		t.Errorf("the starter's setup: %+v", got.Setup)
	}
}

func TestGetMySetupRefusesAGuest(t *testing.T) {
	s := newTableStack(t)
	table := s.mustCreate(t, s.admin(t), "t")
	guest := s.join(t, "", table.GameMeta, "Guest")
	for _, who := range []struct {
		name, token string
		want        int
	}{
		{"no credential", "", http.StatusUnauthorized},
		{"guest seat", guest.Token, http.StatusForbidden},
		{"admin token", s.admin(t), http.StatusForbidden},
	} {
		if got := status(t, s.srv, "GET", "/me/setup", who.token, nil); got != who.want {
			t.Errorf("%s: %d, want %d", who.name, got, who.want)
		}
	}
}

// putSetup stores a setup for user directly.
func (s *tableStack) putSetup(t *testing.T, user uuid.UUID, setup tablesetups.Setup) {
	t.Helper()
	if err := s.setups.Put(context.Background(), user, uuid.Nil, setup); err != nil {
		t.Fatal(err)
	}
}

func settingsJSON(t *testing.T, p game.SettingsPatch) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestCreateWithTheLastSetup(t *testing.T) {
	s := newTableStack(t)
	alice, aliceID := s.signIn(t, "discord-99", "Alice")
	s.putSetup(t, aliceID, tablesetups.Setup{
		Settings: settingsJSON(t, completeSettingsPatch(game.TableSettings{
			UndoLimit: 2, UndoScope: game.UndoScopeOwn, StartingLife: 25,
			CommanderDamage: 21, BotPace: game.BotPaceFast, AllowSpawn: true,
		})),
		Bots: []tablesetups.Bot{
			{Tier: "random", DeckID: "test-mono-white", Name: "Robo"},
			{Tier: "galaxy-brain", DeckID: "test-mono-white", Name: "Smart"},
			{Tier: "random", DeckID: "retired-deck", Name: "Old"},
			{Tier: "random", DeckID: "", Name: "Pasted"},
		},
	})

	code, out, _, raw := s.create(t, alice, createGameRequest{Name: "again", Setup: "last"})
	if code != http.StatusCreated {
		t.Fatalf("create: %d %s", code, raw)
	}
	if out.Setup == nil || !out.Setup.Settings || out.Setup.BotsAdded != 1 {
		t.Fatalf("setup result: %+v", out.Setup)
	}
	skipped := map[string]bool{}
	for _, sk := range out.Setup.Skipped {
		if sk.Reason == "" {
			t.Errorf("%s skipped with no reason", sk.Name)
		}
		skipped[sk.Name] = true
	}
	for _, name := range []string{"Smart", "Old", "Pasted"} {
		if !skipped[name] {
			t.Errorf("%s was not named as skipped: %+v", name, out.Setup.Skipped)
		}
	}
	if out.InviteToken == "" || len(out.Players) != 1 || !out.Players[0].IsBot || out.Players[0].Name != "Robo" {
		t.Errorf("the new table: %+v", out.GameMeta)
	}
	g, _ := s.lobby.LookupGame(out.ID)
	if ts := g.TableSettingsSnapshot(); ts.StartingLife != 25 || ts.BotPace != game.BotPaceFast || !ts.AllowSpawn || ts.UndoLimit != 2 {
		t.Errorf("settings not applied: %+v", ts)
	}

	// No setup yet for Bob: the table is still created, and says so.
	bob, _ := s.signIn(t, "discord-77", "Bob")
	code, out, _, raw = s.create(t, bob, createGameRequest{Name: "fresh", Setup: "last"})
	if code != http.StatusCreated || out.Setup == nil || len(out.Setup.Skipped) != 1 || out.Setup.BotsAdded != 0 {
		t.Errorf("create with no setup: %d %+v %s", code, out.Setup, raw)
	}
	// An unknown setup source is refused before anything is created.
	if code, _, _, _ := s.create(t, bob, createGameRequest{Name: "x", Setup: "favourite"}); code != http.StatusBadRequest {
		t.Errorf("unknown setup: %d, want 400", code)
	}
}

func TestApplySetupToATable(t *testing.T) {
	s := newTableStack(t)
	alice, aliceID := s.signIn(t, "discord-99", "Alice")
	bob, _ := s.signIn(t, "discord-77", "Bob")
	carol, _ := s.signIn(t, "discord-55", "Carol")
	s.putSetup(t, aliceID, tablesetups.Setup{
		Settings: settingsJSON(t, game.SettingsPatch{StartingLife: intp(20)}),
		Bots: []tablesetups.Bot{
			{Tier: "random", DeckID: "test-mono-white", Name: "B1"},
			{Tier: "random", DeckID: "test-mono-white", Name: "B2"},
			{Tier: "random", DeckID: "test-mono-white", Name: "B3"},
			{Tier: "random", DeckID: "test-mono-white", Name: "B4"},
		},
	})

	table := s.mustCreate(t, alice, "apply here")
	aliceSeat := s.join(t, alice, table.GameMeta, "")
	bobSeat := s.join(t, bob, table.GameMeta, "")
	path := "/games/" + table.ID.String() + "/setup"
	body := applySetupRequest{From: "last"}

	// Who may apply it: the creator, yes; a seated non-host (Alice is
	// the named host), an unrelated signed-in person and a guest, no.
	elsewhere := s.mustCreate(t, s.admin(t), "elsewhere")
	guest := s.join(t, "", elsewhere.GameMeta, "Guest")
	for _, who := range []struct {
		name, token string
		want        int
	}{
		{"no credential", "", http.StatusUnauthorized},
		{"guest seat", guest.Token, http.StatusForbidden},
		{"a seated non-host", bobSeat.Token, http.StatusForbidden},
		{"an unrelated person", carol, http.StatusForbidden},
	} {
		if got := status(t, s.srv, "POST", path, who.token, body); got != who.want {
			t.Errorf("%s: %d, want %d", who.name, got, who.want)
		}
	}
	if got := status(t, s.srv, "POST", path, alice, applySetupRequest{From: "best"}); got != http.StatusBadRequest {
		t.Errorf("unknown from: %d, want 400", got)
	}

	resp := postJSON(t, s.srv, path, alice, body)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("apply: %d %s", resp.StatusCode, raw)
	}
	var out applySetupResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	// Two humans are seated, so two of the four bots fit.
	if !out.Settings || out.BotsAdded != 2 || len(out.Skipped) != 2 || len(out.Game.Players) != 4 {
		t.Errorf("apply result: %+v", out)
	}
	for _, sk := range out.Skipped {
		if sk.Reason != "the table has no free seat" {
			t.Errorf("skip reason: %+v", sk)
		}
	}
	if !out.Game.IsCreator || out.Game.InviteToken == "" {
		t.Errorf("the creator's view of the table: %+v", out.Game)
	}

	// Somebody with no setup: 404.
	if got := status(t, s.srv, "POST", "/games/"+s.mustCreate(t, carol, "carol's").ID.String()+"/setup", carol, body); got != http.StatusNotFound {
		t.Errorf("no setup: %d, want 404", got)
	}

	// A started table: 409.
	for _, seat := range []sessionResponse{aliceSeat, bobSeat} {
		s.uploadTestDeck(t, seat, table.ID)
	}
	if got := status(t, s.srv, "POST", "/games/"+table.ID.String()+"/start", bobSeat.Token, nil); got != http.StatusOK {
		t.Fatalf("start: %d", got)
	}
	if got := status(t, s.srv, "POST", path, alice, body); got != http.StatusConflict {
		t.Errorf("a started table: %d, want 409", got)
	}
}

// TestApplySetupByTheHost: a seated host who did not create the table
// may apply their own setup; the admin token has none to apply.
func TestApplySetupByTheHost(t *testing.T) {
	s := newTableStack(t)
	alice, aliceID := s.signIn(t, "discord-99", "Alice")
	s.putSetup(t, aliceID, tablesetups.Setup{Bots: []tablesetups.Bot{{Tier: "random", DeckID: "test-mono-white", Name: "B"}}})
	table := s.mustCreate(t, s.admin(t), "token table")
	seat := s.join(t, alice, table.GameMeta, "")
	if got := status(t, s.srv, "POST", "/games/"+table.ID.String()+"/setup", seat.Token, applySetupRequest{From: "last"}); got != http.StatusOK {
		t.Errorf("the host applying: %d", got)
	}
	if got := status(t, s.srv, "POST", "/games/"+table.ID.String()+"/setup", s.admin(t), applySetupRequest{From: "last"}); got != http.StatusForbidden {
		t.Errorf("the admin token: %d, want 403 (it is not a person and has no setup)", got)
	}
}

func TestBuildSetupOrdersBotsBySeat(t *testing.T) {
	owner, mate := uuid.New(), uuid.New()
	meta := GameMeta{Players: []SeatInfo{
		{Seat: 2, IsBot: true, BotTier: "random", BotDeck: "b", Name: "Second"},
		{Seat: 0, UserID: owner.String()},
		{Seat: 3, UserID: mate.String()},
		{Seat: 1, IsBot: true, BotTier: "heuristic", BotDeck: "a", Name: "First"},
		{Seat: 4, UserID: mate.String()},
	}}
	setup, err := buildSetup(meta, game.DefaultTableSettings(), owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(setup.Bots) != 2 || setup.Bots[0].Name != "First" || setup.Bots[1].Name != "Second" {
		t.Errorf("bots: %+v", setup.Bots)
	}
	if len(setup.Tablemates) != 1 || setup.Tablemates[0] != mate {
		t.Errorf("tablemates: %v", setup.Tablemates)
	}
}
