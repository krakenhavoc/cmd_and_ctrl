package lobby

// ADR 0212 phase B2: a seat carries its deck's commanders, resolved the
// way a saved deck's are, kept across a restart by reading the restored
// game, and served only to the table's own people (owner decision 3).

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// seatCmdDeckName is the seated deck's name, which is not the
// commander's, so a body that has the deck name but no commander is
// told from one that has both.
const seatCmdDeckName = "Wizard Tribal"

// seatCommanderIndex knows "Test Commander" (W/U) in two printings, both
// with art, the partner pair Tymna (W/B) and Kraum (U/R), and Plains.
// Every image URL is on example.test so a body can be searched for one.
func seatCommanderIndex(t *testing.T) (idx *cards.Index, first, second uuid.UUID) {
	t.Helper()
	idx = cards.NewIndex()
	legal := map[string]string{"commander": "legal"}
	art := func(id uuid.UUID) map[string]string {
		return map[string]string{"art_crop": "https://example.test/" + id.String() + ".jpg"}
	}
	oracle := uuid.New()
	first, second = uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{first, second} {
		idx.Put(cards.Card{ID: id, OracleID: oracle, Name: "Test Commander", TypeLine: "Legendary Creature — Human Wizard",
			ColorIdentity: []string{"W", "U"}, Legalities: legal, ImageURIs: art(id)})
	}
	for _, c := range []struct {
		name     string
		identity []string
	}{{"Tymna the Weaver", []string{"W", "B"}}, {"Kraum, Ludevic's Opus", []string{"U", "R"}}} {
		id := uuid.New()
		idx.Put(cards.Card{ID: id, OracleID: uuid.New(), Name: c.name, TypeLine: "Legendary Creature — Human Cleric",
			ColorIdentity: c.identity, Legalities: legal, ImageURIs: art(id)})
	}
	idx.Put(cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains",
		ColorIdentity: []string{"W"}, Legalities: legal})
	return idx, first, second
}

// seatDeckOf is a deck of game cards led by the named commanders, each
// on the printing given (uuid.Nil: none).
func seatDeckOf(printing uuid.UUID, commanders ...string) []game.Card {
	var out []game.Card
	for _, name := range commanders {
		c := game.NewCommander(name, uuid.Nil)
		if printing != uuid.Nil {
			c.ScryfallID = printing.String()
		}
		out = append(out, c)
	}
	for i := 0; i < 20; i++ {
		out = append(out, game.NewCard(fmt.Sprintf("Plains %d", i), uuid.Nil))
	}
	return out
}

func seatOf(t *testing.T, m GameMeta, id uuid.UUID) SeatInfo {
	t.Helper()
	for _, s := range m.Players {
		if s.PlayerID == id {
			return s
		}
	}
	t.Fatalf("no seat %s in %+v", id, m.Players)
	return SeatInfo{}
}

// revealed is what the table's own people get for a seat.
func revealed(s SeatInfo) SeatInfo {
	return withSeatCommanders(GameMeta{Players: []SeatInfo{s}}, true).Players[0]
}

// The seat's commanders are libraryCommanders' answer, the code GET
// /me/decks uses, with the seated printing as the hint: a deck that
// names a printing draws that one, a partner pair is two rows, and the
// seat's identity is their union in WUBRG order.
func TestSeatCommandersResolveAsASavedDecksDo(t *testing.T) {
	idx, first, second := seatCommanderIndex(t)

	got := resolveSeatCommanders(idx, seatDeckOf(second, "Test Commander"))
	want, wantIdentity := libraryCommanders(idx, []string{"Test Commander"},
		map[string]deck.Entry{"test commander": {PrintingID: second}})
	if got == nil || len(got.cards) != 1 || got.cards[0].ScryfallID != second.String() {
		t.Fatalf("commanders = %+v, want the seated printing %s", got, second)
	}
	if got.cards[0].Name != want[0].Name || got.cards[0].ScryfallID != want[0].ScryfallID ||
		!slices.Equal(got.cards[0].ColorIdentity, want[0].ColorIdentity) || !slices.Equal(got.identity, wantIdentity) {
		t.Errorf("commanders = %+v %v, want libraryCommanders' %+v %v", got.cards, got.identity, want, wantIdentity)
	}
	if !slices.Equal(got.identity, []string{"W", "U"}) {
		t.Errorf("identity = %v, want [W U]", got.identity)
	}
	if other := resolveSeatCommanders(idx, seatDeckOf(first, "Test Commander")); other.cards[0].ScryfallID != first.String() {
		t.Errorf("the other printing seated draws %q, want %s", other.cards[0].ScryfallID, first)
	}

	pair := resolveSeatCommanders(idx, seatDeckOf(uuid.Nil, "Tymna the Weaver", "Kraum, Ludevic's Opus"))
	if pair == nil || len(pair.cards) != 2 || pair.cards[0].Name != "Tymna the Weaver" || pair.cards[1].Name != "Kraum, Ludevic's Opus" {
		t.Fatalf("partner pair = %+v", pair)
	}
	if pair.cards[0].ScryfallID == "" || pair.cards[1].ScryfallID == "" {
		t.Errorf("a partner with art has no printing: %+v", pair.cards)
	}
	if !slices.Equal(pair.identity, []string{"W", "U", "B", "R"}) {
		t.Errorf("pair identity = %v, want [W U B R]", pair.identity)
	}

	// A commander the index does not know keeps its name, with no id and
	// no identity, as a saved deck's does; the tile washes.
	unknown := resolveSeatCommanders(idx, seatDeckOf(uuid.New(), "Nobody Knows Me"))
	if unknown == nil || unknown.cards[0].Name != "Nobody Knows Me" || unknown.cards[0].ScryfallID != "" || unknown.cards[0].ColorIdentity != nil {
		t.Errorf("unknown commander = %+v", unknown)
	}
	if resolveSeatCommanders(nil, seatDeckOf(first, "Test Commander")) != nil {
		t.Error("commanders with no card index")
	}
	if resolveSeatCommanders(idx, botDeck(5)[1:]) != nil {
		t.Error("commanders for a deck with none")
	}
}

// Every way a deck goes into a seat sets its commanders, a new deck
// replaces them, and the lobby's own copy of a seat never has them on
// the wire fields: only redactMetaFor puts them there.
func TestEverySeatingPathSetsTheSeatsCommanders(t *testing.T) {
	idx, first, second := seatCommanderIndex(t)
	l := newTestLobby(t)
	l.SetCardIndex(idx)
	meta, _ := l.Create("FNM")
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	if c := seatOf(t, mustGet(t, l, meta.ID), alice).commanders; c != nil {
		t.Errorf("the placeholder deck has commanders: %+v", c)
	}
	if _, err := l.SetDeck(meta.ID, alice, seatCmdDeckName, seatDeckOf(second, "Test Commander")); err != nil {
		t.Fatal(err)
	}
	_, bot, err := l.AddBot(meta.ID, "Bot 1", "random", "", "Partners", seatDeckOf(uuid.Nil, "Tymna the Weaver", "Kraum, Ludevic's Opus"))
	if err != nil {
		t.Fatal(err)
	}

	m := mustGet(t, l, meta.ID)
	if c := seatOf(t, m, alice).commanders; c == nil || c.cards[0].ScryfallID != second.String() {
		t.Errorf("a seated deck's commanders = %+v", c)
	}
	if c := seatOf(t, m, bot).commanders; c == nil || len(c.cards) != 2 {
		t.Errorf("a bot's commanders = %+v", c)
	}
	raw, _ := json.Marshal(m)
	if strings.Contains(string(raw), "commander_cards") || strings.Contains(string(raw), second.String()) {
		t.Errorf("the lobby's own meta carries the commanders on the wire: %s", raw)
	}

	// A new deck replaces them; a deck with no commander clears them.
	if _, err := l.SetDeck(meta.ID, alice, "Other", seatDeckOf(first, "Test Commander")); err != nil {
		t.Fatal(err)
	}
	if c := seatOf(t, mustGet(t, l, meta.ID), alice).commanders; c == nil || c.cards[0].ScryfallID != first.String() {
		t.Errorf("after a new deck: %+v", c)
	}
	if _, err := l.SetDeck(meta.ID, alice, "None", botDeck(5)[1:]); err != nil {
		t.Fatal(err)
	}
	if c := seatOf(t, mustGet(t, l, meta.ID), alice).commanders; c != nil {
		t.Errorf("a deck with no commander kept the old ones: %+v", c)
	}
}

// Kept across a restart by reading the restored game, not a column: in
// the lobby, and once the game is under way, with the printing seated.
func TestSeatCommandersComeBackAfterARestart(t *testing.T) {
	idx, _, second := seatCommanderIndex(t)
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	l.SetCardIndex(idx)
	meta, _ := l.Create("FNM")
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	_, bob, err := l.Join(meta.ID, meta.InviteToken, "Bob")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.SetDeck(meta.ID, alice, seatCmdDeckName, seatDeckOf(second, "Test Commander")); err != nil {
		t.Fatal(err)
	}
	_, bot, err := l.AddBot(meta.ID, "Bot 1", "random", "", "Partners", seatDeckOf(uuid.Nil, "Tymna the Weaver", "Kraum, Ludevic's Opus"))
	if err != nil {
		t.Fatal(err)
	}
	before := mustGet(t, l, meta.ID)

	check := func(t *testing.T, back GameMeta) {
		t.Helper()
		if c := seatOf(t, back, alice).commanders; c == nil || len(c.cards) != 1 || c.cards[0].ScryfallID != second.String() ||
			!slices.Equal(c.identity, []string{"W", "U"}) {
			t.Errorf("Alice's commanders after the restart = %+v, want Test Commander on %s", c, second)
		}
		got, want := revealed(seatOf(t, back, bot)), revealed(seatOf(t, before, bot))
		if len(got.CommanderCards) != 2 || len(want.CommanderCards) != 2 || got.CommanderCards[0].Name != want.CommanderCards[0].Name ||
			got.CommanderCards[1].ScryfallID != want.CommanderCards[1].ScryfallID || !slices.Equal(got.ColorIdentity, want.ColorIdentity) {
			t.Errorf("the bot's commanders after the restart = %+v, before %+v", got, want)
		}
		if c := seatOf(t, back, bob).commanders; c != nil {
			t.Errorf("Bob has no deck in but came back with commanders %+v", c)
		}
	}

	l2, _ := newDurableLobby(t, dir)
	l2.SetCardIndex(idx)
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}
	check(t, mustGet(t, l2, meta.ID))

	// Under way: Bob seats a deck, the table starts, and the server goes
	// down again. The commanders are wherever the game has them.
	if _, err := l2.SetDeck(meta.ID, bob, "Bob's", botDeck(20)); err != nil {
		t.Fatal(err)
	}
	if _, err := l2.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	l3, _ := newDurableLobby(t, dir)
	l3.SetCardIndex(idx)
	if n := l3.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games after the start, want 1", n)
	}
	back := mustGet(t, l3, meta.ID)
	if back.State != string(game.StateActive) {
		t.Fatalf("state %q after the restart, want active", back.State)
	}
	if c := seatOf(t, back, alice).commanders; c == nil || c.cards[0].ScryfallID != second.String() {
		t.Errorf("Alice's commanders in a restored game under way = %+v", c)
	}
	if c := seatOf(t, back, bot).commanders; c == nil || len(c.cards) != 2 {
		t.Errorf("the bot's commanders in a restored game under way = %+v", c)
	}
}

// Owner decision 3: before Start, and after, a seat's commanders go to
// the table's own people (its seats, its spectators, its creator and the
// admin) and to nobody else, while deck_name stays for everyone.
func TestOnlyTheTablesOwnPeopleSeeItsCommanders(t *testing.T) {
	idx, _, _ := seatCommanderIndex(t)
	l := newTestLobby(t)
	l.SetCardIndex(idx)
	creator, bobUser := uuid.New(), uuid.New()
	meta, err := l.CreateBy("FNM", creator)
	if err != nil {
		t.Fatal(err)
	}
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	_, bob, err := l.JoinAs(meta.ID, meta.InviteToken, "Bob", DiscordIdentity{ID: "9001", Username: "bob"}, bobUser)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.SetDeck(meta.ID, alice, seatCmdDeckName, seatDeckOf(uuid.Nil, "Test Commander")); err != nil {
		t.Fatal(err)
	}
	if _, err := l.SetDeck(meta.ID, bob, "Partners", seatDeckOf(uuid.Nil, "Tymna the Weaver", "Kraum, Ludevic's Opus")); err != nil {
		t.Fatal(err)
	}
	other, _ := l.Create("Another table")
	_, otherSeat, err := l.Join(other.ID, other.InviteToken, "Carol")
	if err != nil {
		t.Fatal(err)
	}

	id := meta.ID
	cases := []struct {
		name   string
		p      auth.Principal
		admin  bool
		reveal bool
	}{
		{"the admin", auth.Principal{Role: auth.RoleAdmin}, true, true},
		{"the creator, not seated", auth.Principal{Role: auth.RoleIdentified, UserID: creator}, false, true},
		{"a guest seat", auth.Principal{Role: auth.RolePlayer, GameID: id, PlayerID: alice}, false, true},
		{"a signed-in seat, through its own session", auth.Principal{Role: auth.RolePlayer, GameID: id, PlayerID: bob, UserID: bobUser}, false, true},
		{"a signed-in seat, signed in elsewhere", auth.Principal{Role: auth.RoleIdentified, UserID: bobUser}, false, true},
		{"a signed-in seat, seated at another table", auth.Principal{Role: auth.RolePlayer, GameID: other.ID, PlayerID: otherSeat, UserID: bobUser}, false, true},
		{"a spectator", auth.Principal{Role: auth.RoleSpectator, GameID: id, SpectatorID: uuid.New()}, false, true},
		{"an unrelated signed-in person", auth.Principal{Role: auth.RoleIdentified, UserID: uuid.New()}, false, false},
		{"a seat at another table", auth.Principal{Role: auth.RolePlayer, GameID: other.ID, PlayerID: otherSeat}, false, false},
		{"a spectator of another table", auth.Principal{Role: auth.RoleSpectator, GameID: other.ID, SpectatorID: uuid.New()}, false, false},
		{"a session for a seat that left", auth.Principal{Role: auth.RolePlayer, GameID: id, PlayerID: uuid.New()}, false, false},
		{"an identified session naming this game", auth.Principal{Role: auth.RoleIdentified, GameID: id, UserID: uuid.New()}, false, false},
	}
	for _, started := range []bool{false, true} {
		if started {
			if _, err := l.Start(id); err != nil {
				t.Fatal(err)
			}
		}
		raw := mustGet(t, l, id)
		for _, tc := range cases {
			t.Run(fmt.Sprintf("%s/started=%v", tc.name, started), func(t *testing.T) {
				out := redactMetaFor(tc.p, tc.admin, id, raw)
				for _, s := range out.Players {
					if s.DeckName == "" {
						t.Errorf("%s lost its deck name", s.Name)
					}
					if has := len(s.CommanderCards) > 0 && len(s.ColorIdentity) > 0; has != tc.reveal {
						t.Errorf("%s: commanders %+v identity %v, want shown=%v", s.Name, s.CommanderCards, s.ColorIdentity, tc.reveal)
					}
				}
				body, _ := json.Marshal(out)
				if !tc.reveal && (strings.Contains(string(body), "commander_cards") || strings.Contains(string(body), "color_identity")) {
					t.Errorf("body names the commanders: %s", body)
				}
			})
		}
		// The caller's copy is never written to.
		if raw.Players[0].CommanderCards != nil {
			t.Error("redactMetaFor wrote the commanders onto its caller's seats")
		}
	}
}

// The routes that list and read a table, and the invite preview a
// visitor reads before they have a session: the table's own people get
// the commanders, and an unrelated signed-in session and a signed-out
// caller never do.
func TestGamesRoutesKeepCommandersToTheTable(t *testing.T) {
	idx, _, _ := seatCommanderIndex(t)
	srv, l, a, _ := newTestHTTPStackWithCards(t, idx)
	creator := uuid.New()
	meta, err := l.CreateBy("FNM", creator)
	if err != nil {
		t.Fatal(err)
	}
	alice := joinAs(t, srv, meta, "Alice")
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", alice.Token, uploadDeckRequest{
		Format: "text", Source: "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n", PlayerID: alice.PlayerID,
	})
	raw, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upload: %d %s", resp.StatusCode, raw)
	}
	// The seat's own answer carries them, so its tile turns at once.
	var up uploadDeckResponse
	if err := json.Unmarshal(raw, &up); err != nil {
		t.Fatal(err)
	}
	cmd := seatOf(t, up.Game, alice.PlayerID)
	if len(cmd.CommanderCards) != 1 || cmd.CommanderCards[0].Name != "Test Commander" || cmd.CommanderCards[0].ScryfallID == "" ||
		!slices.Equal(cmd.ColorIdentity, []string{"W", "U"}) {
		t.Fatalf("the seat's upload answer: %+v", cmd)
	}
	printing := cmd.CommanderCards[0].ScryfallID
	if strings.Contains(string(raw), "example.test") {
		t.Errorf("the upload answer carries an image URL: %s", raw)
	}

	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/spectate", "", spectateRequest{InviteToken: meta.SpectatorInvite, Name: "Watcher"})
	var spec sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&spec)
	resp.Body.Close()
	if spec.Token == "" {
		t.Fatal("no spectator session")
	}

	get := func(path, tok string) (int, string) {
		t.Helper()
		resp := doGet(t, srv, path, tok)
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(b)
	}
	shows := func(body string) bool {
		return strings.Contains(body, `"commander_cards"`) && strings.Contains(body, printing)
	}
	hides := func(body string) bool {
		return !strings.Contains(body, "commander_cards") && !strings.Contains(body, "color_identity") && !strings.Contains(body, printing)
	}

	members := map[string]string{
		"the seat":      alice.Token,
		"its spectator": spec.Token,
		"its creator":   identifiedToken(t, a, creator),
		"the admin":     adminSession(t, a),
	}
	stranger := identifiedToken(t, a, uuid.New())
	for _, path := range []string{"/games", "/games/" + meta.ID.String()} {
		for who, tok := range members {
			if code, body := get(path, tok); code != http.StatusOK || !shows(body) {
				t.Errorf("GET %s as %s: %d, want the commanders: %s", path, who, code, body)
			}
		}
		code, body := get(path, stranger)
		if code != http.StatusOK || !hides(body) {
			t.Errorf("GET %s as an unrelated signed-in session: %d, want no commanders: %s", path, code, body)
		}
		if !strings.Contains(body, `"deck_uploaded":true`) {
			t.Errorf("GET %s as an unrelated signed-in session lost the seat's readiness: %s", path, body)
		}
		if code, body := get(path, ""); code != http.StatusUnauthorized || !hides(body) {
			t.Errorf("GET %s signed out: %d, want 401 and no commanders: %s", path, code, body)
		}
	}

	// The invite preview is read before a session exists (ADR 0092): it
	// never carries them, whoever asks.
	for who, tok := range map[string]string{"signed out": "", "an unrelated signed-in session": stranger, "the seat": alice.Token} {
		code, body := get("/games/"+meta.ID.String()+"/preview?t="+meta.InviteToken, tok)
		if code != http.StatusOK || !hides(body) {
			t.Errorf("GET preview as %s: %d, want no commanders: %s", who, code, body)
		}
	}
}

// A pre-built deck and a bot from the catalog get their commanders the
// same way, by the code GET /decks and GET /bot/options draw them with.
func TestPrebuiltAndBotSeatsCarryTheirCommanders(t *testing.T) {
	d := decks.All()[0]
	idx := buildPrebuiltDeckIndex(t, d)
	cmd, ok := idx.FindByName(d.Commander.Name)
	if !ok {
		t.Fatalf("index has no %s", d.Commander.Name)
	}
	cmd.ImageURIs = map[string]string{"art_crop": "https://example.test/c.jpg"}
	idx.Put(cmd)
	srv, l, _, _ := newTestHTTPStackWithCards(t, idx)
	meta, _ := l.Create("FNM")
	alice := joinAs(t, srv, meta, "Alice")
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", alice.Token, uploadDeckRequest{Deck: d.ID, PlayerID: alice.PlayerID})
	var up uploadDeckResponse
	_ = json.NewDecoder(resp.Body).Decode(&up)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("install %s: %d", d.ID, resp.StatusCode)
	}
	got := seatOf(t, up.Game, alice.PlayerID).CommanderCards
	want := catalogCommanderCards(idx, d.Commander.Name)
	if len(got) != 1 || got[0].Name != want[0].Name || got[0].ScryfallID != want[0].ScryfallID || got[0].ScryfallID != cmd.ID.String() {
		t.Errorf("a pre-built seat's commanders = %+v, want GET /decks' %+v", got, want)
	}

	bsrv, bl, _, _ := newBotHTTPStack(t, buildBotCommanderIndex(t), fakeDeckSource{})
	bmeta, _ := bl.Create("FNM")
	bob := joinAs(t, bsrv, bmeta, "Bob")
	resp = postJSON(t, bsrv, "/games/"+bmeta.ID.String()+"/seats/bot", bob.Token, addBotRequest{Tier: "random", Deck: "test-mono-white"})
	var added addBotResponse
	_ = json.NewDecoder(resp.Body).Decode(&added)
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("add bot: %d", resp.StatusCode)
	}
	if c := seatOf(t, added.Game, added.PlayerID).CommanderCards; len(c) != 1 || c[0].Name != "Test Commander" || c[0].ScryfallID == "" {
		t.Errorf("a bot seat's commanders = %+v", c)
	}
}

// buildBotCommanderIndex is buildMinimalDeckIndex with art on the
// commander, so a seated bot has a printing to draw.
func buildBotCommanderIndex(t *testing.T) *cards.Index {
	t.Helper()
	idx := buildMinimalDeckIndex(t)
	c, ok := idx.FindByName("Test Commander")
	if !ok {
		t.Fatal("no Test Commander")
	}
	c.ImageURIs = map[string]string{"art_crop": "https://example.test/tc.jpg"}
	idx.Put(c)
	return idx
}

// The #looking-for-game post is read by the whole Discord server: it
// names the table, the host and the seat count, never a seat's deck.
func TestLFGPostNamesNoCommander(t *testing.T) {
	idx, _, second := seatCommanderIndex(t)
	l := newTestLobby(t)
	l.SetCardIndex(idx)
	h := newLFGHarness(t, l)
	meta, alice := h.table("Friday pod")
	if _, err := l.SetDeck(meta.ID, alice, seatCmdDeckName, seatDeckOf(second, "Test Commander")); err != nil {
		t.Fatal(err)
	}
	h.later()
	p, ok := h.post(meta.ID)
	if !ok {
		t.Fatalf("no post; calls %v", h.fake.log())
	}
	body, _ := json.Marshal(p)
	for _, leak := range []string{"Test Commander", second.String(), seatCmdDeckName} {
		if strings.Contains(string(body), leak) {
			t.Errorf("the #looking-for-game post names %q: %s", leak, body)
		}
	}
}

// A signed-out visitor never reaches a seat's commanders through the
// session routes either: joining or spectating answers with the table,
// and that answer is not redactMetaFor's, so it carries none.
func TestSessionAnswersCarryNoCommanders(t *testing.T) {
	idx, _, second := seatCommanderIndex(t)
	srv, l, _, _ := newTestHTTPStackWithCards(t, idx)
	meta, _ := l.Create("FNM")
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.SetDeck(meta.ID, alice, seatCmdDeckName, seatDeckOf(second, "Test Commander")); err != nil {
		t.Fatal(err)
	}
	for path, body := range map[string]any{
		"/games/" + meta.ID.String() + "/join":     joinRequest{InviteToken: meta.InviteToken, Name: "Visitor"},
		"/games/" + meta.ID.String() + "/spectate": spectateRequest{InviteToken: meta.SpectatorInvite, Name: "Watcher"},
	} {
		resp := postJSON(t, srv, path, "", body)
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST %s: %d %s", path, resp.StatusCode, raw)
		}
		if strings.Contains(string(raw), "commander_cards") || strings.Contains(string(raw), second.String()) {
			t.Errorf("POST %s answers with the commanders: %s", path, raw)
		}
	}
}
