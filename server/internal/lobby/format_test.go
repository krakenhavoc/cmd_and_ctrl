package lobby

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/tablesetups"
)

// format_test.go: a format on the table through the lobby (ADR 0222
// §§1–2, #2352): creating a Standard table, its seat cap, the deck
// check per format, saved decks and saved setups. Each test also pins
// the Commander side: a table that names no format is what it was.

// putStandardDecks adds every curated Standard deck's cards to idx as
// Standard-legal, and returns idx.
func putStandardDecks(idx *cards.Index) *cards.Index {
	for _, d := range decks.In(game.FormatStandard) {
		for _, c := range d.Mainboard {
			typeLine := "Creature — Test"
			if c.Basic {
				typeLine = "Basic Land — " + c.Name
			}
			card := cards.Card{
				ID:            uuid.New(),
				Name:          c.Name,
				TypeLine:      typeLine,
				ColorIdentity: []string{c.Identity},
				Legalities:    map[string]string{"standard": "legal", "commander": "legal"},
			}
			if c.OracleID != "" {
				card.OracleID = uuid.MustParse(c.OracleID)
			}
			idx.Put(card)
		}
	}
	return idx
}

// standardPaste is a legal sixty in the plain-text dialect: cards of the
// first curated Standard deck, pasted as a player would.
func standardPaste() string {
	return decks.In(game.FormatStandard)[0].Decklist()
}

func bodyOf(resp *http.Response) string {
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return string(raw)
}

// seatAt joins the table as a guest and returns the seat.
func seatAt(t *testing.T, s *tableStack, meta GameMeta, name string) sessionResponse {
	t.Helper()
	return s.join(t, "", meta, name)
}

func standardTable(t *testing.T, s *tableStack) createGameResponse {
	t.Helper()
	code, out, _, raw := s.create(t, s.admin(t), createGameRequest{Name: "FNM", Format: "standard"})
	if code != http.StatusCreated {
		t.Fatalf("create a Standard table: %d %s", code, raw)
	}
	return out
}

func newFormatStack(t *testing.T) *tableStack {
	t.Helper()
	s := newTableStack(t)
	putStandardDecks(s.cards)
	return s
}

func TestCreateAStandardTable(t *testing.T) {
	s := newFormatStack(t)
	out := standardTable(t, s)
	if out.Format != "standard" || out.MaxSeats != 2 {
		t.Errorf("created meta: format %q, max_seats %d; want standard, 2", out.Format, out.MaxSeats)
	}
	g, err := s.lobby.LookupGame(out.ID)
	if err != nil {
		t.Fatal(err)
	}
	ts := g.TableSettingsSnapshot()
	if ts.EffectiveFormat() != game.FormatStandard || ts.StartingLife != 20 {
		t.Errorf("settings: format %q, starting life %d; want standard, 20", ts.EffectiveFormat(), ts.StartingLife)
	}

	// The table list and the join page say so.
	listed, err := s.lobby.Get(out.ID)
	if err != nil {
		t.Fatal(err)
	}
	if listed.Format != "standard" || listed.MaxSeats != 2 {
		t.Errorf("listed meta: format %q, max_seats %d", listed.Format, listed.MaxSeats)
	}
	resp := doGet(t, s.srv, "/games/"+out.ID.String()+"/preview?t="+out.InviteToken, "")
	var preview previewResponse
	if err := json.NewDecoder(resp.Body).Decode(&preview); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if preview.MaxSeats != 2 || preview.Game.Format != "standard" {
		t.Errorf("preview: max_seats %d, format %q", preview.MaxSeats, preview.Game.Format)
	}

	// Two seats, and no third.
	seatAt(t, s, out.GameMeta, "Alice")
	seatAt(t, s, out.GameMeta, "Bob")
	resp = postJSON(t, s.srv, "/games/"+out.ID.String()+"/join", "", joinRequest{InviteToken: out.InviteToken, Name: "Carol"})
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("a third seat at a Standard table: %d %s, want 409", resp.StatusCode, bodyOf(resp))
	} else {
		resp.Body.Close()
	}
	for _, p := range g.Snapshot().Seats {
		if p.Life != 20 {
			t.Errorf("%s sat down at %d life, want 20", p.Name, p.Life)
		}
	}
}

func TestCreateRefusesAnUnknownFormat(t *testing.T) {
	s := newFormatStack(t)
	before := len(s.lobby.List())
	code, _, _, raw := s.create(t, s.admin(t), createGameRequest{Name: "x", Format: "vintage"})
	if code != http.StatusBadRequest || !strings.Contains(raw, "vintage") {
		t.Errorf("unknown format: %d %s, want 400 naming it", code, raw)
	}
	if got := len(s.lobby.List()); got != before {
		t.Errorf("a refused create left a table behind: %d tables, was %d", got, before)
	}
}

// A table that names no format, or names Commander, is the table it
// always was: four seats, 40 life, and no format on its wire.
func TestATableWithNoFormatIsACommanderTable(t *testing.T) {
	s := newFormatStack(t)
	for _, format := range []string{"", "commander", "Commander"} {
		t.Run("format="+format, func(t *testing.T) {
			code, out, _, raw := s.create(t, s.admin(t), createGameRequest{Name: "pod", Format: format})
			if code != http.StatusCreated {
				t.Fatalf("create: %d %s", code, raw)
			}
			if strings.Contains(raw, `"format"`) || strings.Contains(raw, `"max_seats"`) {
				t.Errorf("a Commander table's create answer names a format or a seat cap: %s", raw)
			}
			g, _ := s.lobby.LookupGame(out.ID)
			if ts := g.TableSettingsSnapshot(); ts != game.DefaultTableSettings() {
				t.Errorf("settings %+v, want the defaults", ts)
			}
			for i := 0; i < game.MaxPlayers; i++ {
				seatAt(t, s, out.GameMeta, fmt.Sprintf("P%d", i+1))
			}
			resp := postJSON(t, s.srv, "/games/"+out.ID.String()+"/join", "", joinRequest{InviteToken: out.InviteToken, Name: "P5"})
			if resp.StatusCode != http.StatusConflict {
				t.Errorf("a fifth seat: %d, want 409", resp.StatusCode)
			}
			resp.Body.Close()
			resp = doGet(t, s.srv, "/games/"+out.ID.String()+"/preview?t="+out.InviteToken, "")
			var preview previewResponse
			_ = json.NewDecoder(resp.Body).Decode(&preview)
			resp.Body.Close()
			if preview.MaxSeats != game.MaxPlayers {
				t.Errorf("preview max_seats %d, want %d", preview.MaxSeats, game.MaxPlayers)
			}
		})
	}
}

func uploadFor(t *testing.T, s *tableStack, gameID uuid.UUID, seat sessionResponse, body uploadDeckRequest) (int, string) {
	t.Helper()
	body.PlayerID = seat.PlayerID
	resp := postJSON(t, s.srv, "/games/"+gameID.String()+"/decks", seat.Token, body)
	return resp.StatusCode, bodyOf(resp)
}

// A deck is checked against the format of the table it is seated at.
func TestDeckIsCheckedAgainstTheTablesFormat(t *testing.T) {
	s := newFormatStack(t)
	std := standardTable(t, s)
	alice := seatAt(t, s, std.GameMeta, "Alice")

	// A pasted sixty: accepted, no commander.
	code, raw := uploadFor(t, s, std.ID, alice, uploadDeckRequest{Format: "text", Source: standardPaste()})
	if code != http.StatusOK {
		t.Fatalf("a Standard list at a Standard table: %d %s", code, raw)
	}
	var ok uploadDeckResponse
	if err := json.Unmarshal([]byte(raw), &ok); err != nil {
		t.Fatal(err)
	}
	if ok.CardCount != 60 || len(ok.Commanders) != 0 {
		t.Errorf("seated %d cards with commanders %v; want 60 and none", ok.CardCount, ok.Commanders)
	}

	// A Commander list at a Standard table: refused, saying why.
	bob := seatAt(t, s, std.GameMeta, "Bob")
	code, raw = uploadFor(t, s, std.ID, bob, uploadDeckRequest{Format: "text", Source: testDeckSource})
	if code != http.StatusUnprocessableEntity || !strings.Contains(raw, "commander_not_allowed") {
		t.Errorf("a Commander list at a Standard table: %d %s", code, raw)
	}

	// A Standard list at a Commander table: refused as it always
	// would have been.
	code, cmd, _, _ := s.create(t, s.admin(t), createGameRequest{Name: "pod"})
	if code != http.StatusCreated {
		t.Fatal("create a Commander table")
	}
	carol := seatAt(t, s, cmd.GameMeta, "Carol")
	code, raw = uploadFor(t, s, cmd.ID, carol, uploadDeckRequest{Format: "text", Source: standardPaste()})
	if code != http.StatusUnprocessableEntity || !strings.Contains(raw, "missing_commander") {
		t.Errorf("a Standard list at a Commander table: %d %s", code, raw)
	}
	// And a Commander list there is fine.
	if code, raw = uploadFor(t, s, cmd.ID, carol, uploadDeckRequest{Format: "text", Source: testDeckSource}); code != http.StatusOK {
		t.Errorf("a Commander list at a Commander table: %d %s", code, raw)
	}
}

// A built-in deck is offered at tables of its format and refused at
// another's with one sentence, not a list of violations.
func TestBuiltInDeckMustFitTheTablesFormat(t *testing.T) {
	s := newFormatStack(t)
	std := standardTable(t, s)
	alice := seatAt(t, s, std.GameMeta, "Alice")
	standardID := decks.In(game.FormatStandard)[0].ID
	commanderID := decks.All()[0].ID

	code, raw := uploadFor(t, s, std.ID, alice, uploadDeckRequest{Deck: commanderID})
	if code != http.StatusUnprocessableEntity || !strings.Contains(raw, "Commander deck") || !strings.Contains(raw, "Standard table") {
		t.Errorf("a Commander built-in at a Standard table: %d %s", code, raw)
	}
	if code, raw = uploadFor(t, s, std.ID, alice, uploadDeckRequest{Deck: standardID}); code != http.StatusOK {
		t.Fatalf("a Standard built-in at a Standard table: %d %s", code, raw)
	}

	_, cmd, _, _ := s.create(t, s.admin(t), createGameRequest{Name: "pod"})
	bob := seatAt(t, s, cmd.GameMeta, "Bob")
	code, raw = uploadFor(t, s, cmd.ID, bob, uploadDeckRequest{Deck: standardID})
	if code != http.StatusUnprocessableEntity || !strings.Contains(raw, "Standard deck") || !strings.Contains(raw, "Commander table") {
		t.Errorf("a Standard built-in at a Commander table: %d %s", code, raw)
	}
	if code, raw = uploadFor(t, s, cmd.ID, bob, uploadDeckRequest{Deck: commanderID}); code != http.StatusOK {
		t.Errorf("a Commander built-in at a Commander table: %d %s", code, raw)
	}
}

// Every curated Standard deck installs, and two of them start a game at
// 20 life with no commander anywhere.
func TestStandardBuiltInsStartAGame(t *testing.T) {
	s := newFormatStack(t)
	std := standardTable(t, s)
	ids := decks.In(game.FormatStandard)
	for i, d := range ids[:2] {
		seat := seatAt(t, s, std.GameMeta, fmt.Sprintf("P%d", i+1))
		if code, raw := uploadFor(t, s, std.ID, seat, uploadDeckRequest{Deck: d.ID}); code != http.StatusOK {
			t.Fatalf("%s: %d %s", d.ID, code, raw)
		}
	}
	if _, err := s.lobby.Start(std.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	g, _ := s.lobby.LookupGame(std.ID)
	snap := g.Snapshot()
	if len(snap.Seats) != 2 {
		t.Fatalf("%d seats", len(snap.Seats))
	}
	for _, p := range snap.Seats {
		if p.Life != 20 {
			t.Errorf("%s starts at %d life, want 20", p.Name, p.Life)
		}
		if n := p.Command.Size(); n != 0 {
			t.Errorf("%s has %d command-zone cards", p.Name, n)
		}
	}
	if g.HasCommander() {
		t.Error("a Standard game reports a commander")
	}
}

// The prebuilt list names each deck's format, the Standard decks after
// the Commander ones.
func TestPrebuiltDecksListTheStandardDecks(t *testing.T) {
	s := newFormatStack(t)
	resp := doGet(t, s.srv, "/decks", s.admin(t))
	var out prebuiltDecksResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	var standard []prebuiltDeckInfo
	for _, d := range out.Decks {
		if d.Format == "standard" {
			standard = append(standard, d)
		}
	}
	if len(standard) < 2 {
		t.Fatalf("%d Standard decks listed, want at least 2", len(standard))
	}
	for _, d := range standard {
		if d.CardCount != 60 || d.Commander != "" || len(d.CommanderCards) != 0 {
			t.Errorf("%s: card_count %d, commander %q, commander_cards %v", d.ID, d.CardCount, d.Commander, d.CommanderCards)
		}
		if d.Coverage.Cards == 0 || d.Coverage.Full != d.Coverage.Cards {
			t.Errorf("%s: coverage %+v, want every card full", d.ID, d.Coverage)
		}
		total := 0
		for _, c := range d.Cards {
			total += c.Count
		}
		if total != 60 {
			t.Errorf("%s: cards hold %d copies, want 60", d.ID, total)
		}
	}
}

// The format is chosen when the table is made: it may change in the
// lobby until a seat has a deck, and not after.
func TestFormatCannotChangeOnceASeatHasADeck(t *testing.T) {
	s := newFormatStack(t)
	std := standardTable(t, s)
	commander, standard := game.FormatCommander, game.FormatStandard

	// Nobody has a deck: the host may still change their mind.
	if _, err := s.lobby.UpdateSettings(std.ID, uuid.Nil, game.SettingsPatch{Format: &commander}); err != nil {
		t.Fatalf("a format change with no decks loaded: %v", err)
	}
	if m, _ := s.lobby.Get(std.ID); m.Format != "" || m.MaxSeats != 0 {
		t.Errorf("back at Commander the meta still says %q, %d seats", m.Format, m.MaxSeats)
	}
	if _, err := s.lobby.UpdateSettings(std.ID, uuid.Nil, game.SettingsPatch{Format: &standard}); err != nil {
		t.Fatal(err)
	}

	alice := seatAt(t, s, std.GameMeta, "Alice")
	if code, raw := uploadFor(t, s, std.ID, alice, uploadDeckRequest{Format: "text", Source: standardPaste()}); code != http.StatusOK {
		t.Fatalf("upload: %d %s", code, raw)
	}
	req, _ := http.NewRequest(http.MethodPatch, s.srv.URL+"/games/"+std.ID.String()+"/settings", strings.NewReader(`{"format":"commander"}`))
	req.Header.Set("Authorization", "Bearer "+s.admin(t))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("a format change with a deck loaded: %d %s, want 422", resp.StatusCode, bodyOf(resp))
	} else {
		resp.Body.Close()
	}
	if f, _ := s.lobby.Format(std.ID); f != game.FormatStandard {
		t.Errorf("the refused change changed the format to %q", f)
	}
	// Naming the format it has is not a change, and other settings
	// still patch.
	if _, err := s.lobby.UpdateSettings(std.ID, uuid.Nil, game.SettingsPatch{Format: &standard, UndoLimit: intp(3)}); err != nil {
		t.Errorf("a patch naming the current format: %v", err)
	}
}

// A saved deck records the format of the table it was saved from, an
// existing deck is a Commander deck, and a saved deck is seated only at
// tables of its format.
func TestSavedDeckRecordsItsFormat(t *testing.T) {
	s := newFormatStack(t)
	alice, aliceID := s.signIn(t, "discord-99", "Alice")
	ctx := context.Background()

	// A deck saved before formats: the store's Upsert names none.
	old, err := s.library.Upsert(ctx, aliceID, "Old faithful", "text", testDeckSource, []string{"Test Commander"}, 100)
	if err != nil {
		t.Fatal(err)
	}
	if old.Format != "commander" {
		t.Errorf("a deck saved with no format reads as %q, want commander", old.Format)
	}

	std := standardTable(t, s)
	seat := s.join(t, alice, std.GameMeta, "Alice")
	code, raw := uploadFor(t, s, std.ID, seat, uploadDeckRequest{Format: "text", Source: standardPaste()})
	if code != http.StatusOK {
		t.Fatalf("upload: %d %s", code, raw)
	}
	saved, err := s.library.List(ctx, aliceID)
	if err != nil {
		t.Fatal(err)
	}
	var standardDeck, commanderDeck string
	for _, d := range saved {
		switch d.Format {
		case "standard":
			standardDeck = d.ID.String()
			if d.CardCount != 60 || len(d.Commanders) != 0 {
				t.Errorf("the Standard deck was saved as %d cards with commanders %v", d.CardCount, d.Commanders)
			}
		case "commander":
			commanderDeck = d.ID.String()
		default:
			t.Errorf("%s has format %q", d.Name, d.Format)
		}
	}
	if standardDeck == "" || commanderDeck != old.ID.String() {
		t.Fatalf("library after the upload: %+v", saved)
	}

	// GET /me/decks names the Standard deck's format and not the
	// Commander deck's, and checks each against its own format.
	resp := doGet(t, s.srv, "/me/decks", alice)
	listRaw := bodyOf(resp)
	var mine myDecksResponse
	if err := json.Unmarshal([]byte(listRaw), &mine); err != nil {
		t.Fatalf("%v: %s", err, listRaw)
	}
	for _, d := range mine.Decks {
		want := ""
		if d.ID == standardDeck {
			want = "standard"
		}
		if d.Format != want {
			t.Errorf("%s: format %q on the wire, want %q", d.Name, d.Format, want)
		}
	}
	if n := strings.Count(listRaw, `"format"`); n != 1 {
		t.Errorf("%d decks name a format on the wire, want only the Standard one: %s", n, listRaw)
	}

	// The Commander deck is not seated at the Standard table...
	resp = postJSON(t, s.srv, "/games/"+std.ID.String()+"/decks/"+commanderDeck, seat.Token, struct{}{})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("a saved Commander deck at a Standard table: %d %s, want 422", resp.StatusCode, bodyOf(resp))
	} else if body := bodyOf(resp); !strings.Contains(body, "Commander deck") {
		t.Errorf("the refusal should say why: %s", body)
	}
	// ...and the Standard deck is.
	resp = postJSON(t, s.srv, "/games/"+std.ID.String()+"/decks/"+standardDeck, seat.Token, struct{}{})
	if resp.StatusCode != http.StatusOK {
		t.Errorf("a saved Standard deck at a Standard table: %d %s", resp.StatusCode, bodyOf(resp))
	} else {
		resp.Body.Close()
	}
}

// A setup captured at a Standard table makes a Standard table; one
// saved before formats existed names none and is a Commander setup.
func TestSavedSetupCarriesTheFormat(t *testing.T) {
	s := newFormatStack(t)
	alice, aliceID := s.signIn(t, "discord-99", "Alice")

	// What a Standard table leaves behind.
	std := game.DefaultTableSettings()
	std.Format, std.StartingLife = game.FormatStandard, 20
	raw := settingsJSON(t, completeSettingsPatch(std))
	if !strings.Contains(string(raw), `"format":"standard"`) {
		t.Fatalf("a Standard table's setup does not name its format: %s", raw)
	}
	// And a Commander table's names Commander.
	if cmd := settingsJSON(t, completeSettingsPatch(game.DefaultTableSettings())); !strings.Contains(string(cmd), `"format":"commander"`) {
		t.Fatalf("a Commander table's setup does not name its format: %s", cmd)
	}

	s.putSetup(t, aliceID, tablesetups.Setup{Settings: raw})
	code, out, _, body := s.create(t, alice, createGameRequest{Name: "again", Setup: "last"})
	if code != http.StatusCreated || out.Setup == nil || !out.Setup.Settings {
		t.Fatalf("create from a Standard setup: %d %s", code, body)
	}
	if out.Format != "standard" || out.MaxSeats != 2 {
		t.Errorf("the new table: format %q, %d seats; want standard, 2", out.Format, out.MaxSeats)
	}
	g, _ := s.lobby.LookupGame(out.ID)
	if ts := g.TableSettingsSnapshot(); ts.EffectiveFormat() != game.FormatStandard || ts.StartingLife != 20 {
		t.Errorf("settings: %+v", ts)
	}

	// A setup stored before formats existed: every field but the
	// format. Asking for Standard and for that setup gives the setup's
	// table, a Commander one, at the setup's starting life.
	legacy := `{"undo_limit":2,"undo_scope":"own","starting_life":30,"commander_damage":21,"bot_pace":"fast","allow_spawn":false,"free_mulligans":0,"away_after":3}`
	s.putSetup(t, aliceID, tablesetups.Setup{Settings: json.RawMessage(legacy)})
	code, out, _, body = s.create(t, alice, createGameRequest{Name: "old setup", Setup: "last", Format: "standard"})
	if code != http.StatusCreated || out.Setup == nil || !out.Setup.Settings {
		t.Fatalf("create from a pre-format setup: %d %s", code, body)
	}
	if out.Format != "" || out.MaxSeats != 0 {
		t.Errorf("a pre-format setup made a %q table with %d seats; want Commander", out.Format, out.MaxSeats)
	}
	g, _ = s.lobby.LookupGame(out.ID)
	if ts := g.TableSettingsSnapshot(); ts.EffectiveFormat() != game.FormatCommander || ts.StartingLife != 30 || ts.UndoLimit != 2 {
		t.Errorf("settings: %+v", ts)
	}

	// The same setup with no format asked: a Commander table, as before
	// formats.
	code, out, _, body = s.create(t, alice, createGameRequest{Name: "as before", Setup: "last"})
	if code != http.StatusCreated || out.Format != "" {
		t.Fatalf("create from a pre-format setup: %d %s", code, body)
	}
	g, _ = s.lobby.LookupGame(out.ID)
	if ts := g.TableSettingsSnapshot(); ts.EffectiveFormat() != game.FormatCommander || ts.StartingLife != 30 {
		t.Errorf("settings: %+v", ts)
	}
}

// A Commander bot deck is not seated at a Standard table, by the bot
// route or by a setup, and the reason is one sentence.
func TestBotDeckMustFitTheTablesFormat(t *testing.T) {
	s := newFormatStack(t)
	std := standardTable(t, s)
	resp := postJSON(t, s.srv, "/games/"+std.ID.String()+"/seats/bot", s.admin(t),
		addBotRequest{Tier: "random", Deck: "test-mono-white", Name: "Robo"})
	body := bodyOf(resp)
	if resp.StatusCode != http.StatusUnprocessableEntity || !strings.Contains(body, "Commander deck") {
		t.Errorf("a Commander bot deck at a Standard table: %d %s", resp.StatusCode, body)
	}
	// A pasted Standard list seats a bot there.
	resp = postJSON(t, s.srv, "/games/"+std.ID.String()+"/seats/bot", s.admin(t),
		addBotRequest{Tier: "random", Source: standardPaste(), Name: "Robo"})
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Errorf("a bot with a Standard list at a Standard table: %d %s", resp.StatusCode, bodyOf(resp))
	} else {
		resp.Body.Close()
	}
	// The setup path names the same reason.
	reason := seatSetupBot(context.Background(), Config{
		Lobby: s.lobby, Bots: s.host, BotDecks: fakeDeckSource{}, Cards: s.cards,
	}, std.ID, "Robo2", tablesetups.Bot{Tier: "random", DeckID: "test-mono-white"})
	if !strings.Contains(reason, "Commander deck") || strings.Contains(reason, "lobby:") {
		t.Errorf("setup bot skip reason = %q", reason)
	}
}
