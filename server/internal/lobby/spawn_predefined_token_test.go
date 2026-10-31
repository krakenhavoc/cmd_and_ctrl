package lobby

// spawn_predefined_token_test.go — #4738. A host who searched "Treasure"
// on the spawner's Cards tab picked a Scryfall token PRINTING, and the
// spawn made that printing: an object with the printing's oracle ID and
// no catalog entry, so it did not run its own "{T}, Sacrifice this token:
// Add one mana of any color" and the board flagged it unimplemented. A
// predefined token is defined by the rules, not by a printing (CR
// 111.10), so whichever way the spawner is asked for one it must make
// the catalog's template: the object a card's trigger makes.

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/appenv"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// treasurePrinting is the Scryfall record a search for "Treasure"
// returns (the dump's last English Treasure token, tblb), as the index
// holds it.
var treasurePrinting = cards.Card{
	ID:           uuid.MustParse("ff0357fb-d90c-49c7-b16a-8b52ac686c3b"),
	OracleID:     uuid.MustParse("3c549374-6c37-42e0-8d88-a8555d46732d"),
	Name:         "Treasure",
	Layout:       "token",
	SetCode:      "tblb",
	SetType:      "token",
	Lang:         "en",
	TypeLine:     "Token Artifact — Treasure",
	OracleText:   "{T}, Sacrifice this token: Add one mana of any color.",
	ProducedMana: []string{"B", "G", "R", "U", "W"},
}

func predefinedTokenIndex(t *testing.T, printings ...cards.Card) *cards.Index {
	t.Helper()
	idx := devIndex(t)
	for _, c := range printings {
		idx.Put(c)
	}
	return idx
}

// spawnOne posts one spawn onto Alice's battlefield and returns the new
// object as the engine holds it, plus the route's response.
func (f *spawnFixture) spawnOne(field string) (game.Card, spawnResponse) {
	f.t.Helper()
	resp := f.post(f.aliceTok, f.body(","+field))
	if resp.StatusCode != http.StatusOK {
		f.t.Fatalf("spawn %s: status %d (%s)", field, resp.StatusCode, errorMessage(f.t, resp))
	}
	var out spawnResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		f.t.Fatalf("decode: %v", err)
	}
	if len(out.Spawned) != 1 {
		f.t.Fatalf("spawn %s: made %d objects, want 1", field, len(out.Spawned))
	}
	id := uuid.MustParse(out.Spawned[0])
	g := f.lobby.RoomOf(f.gameID).Game
	var got game.Card
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				got, found = c, true
			}
		}
	})
	if !found {
		f.t.Fatalf("spawn %s: %s is not on the battlefield", field, id)
	}
	return got, out
}

// unimplementedInView is the flag the table is shown for one object:
// the board, the client's badge and the MCP seat's "(unimplemented: …)"
// note all read it.
func (f *spawnFixture) unimplementedInView(id uuid.UUID) bool {
	f.t.Helper()
	view, _, err := f.lobby.RoomOf(f.gameID).Snapshot()
	if err != nil {
		f.t.Fatalf("Snapshot: %v", err)
	}
	for _, c := range view.Battlefield.Cards {
		if c.InstanceID == id.String() {
			return c.Unimplemented
		}
	}
	f.t.Fatalf("%s is not in the view's battlefield", id)
	return false
}

// The bug, from every way the spawner can be asked for a Treasure: the
// Tokens tab's key, the Cards tab's Scryfall id (what the client sends
// for a search result) and a bare name. Each must be the Treasure a
// card's trigger makes (effects.TreasureToken), not the printing.
func TestSpawnedTreasureIsThePredefinedTokenHoweverItIsAsked(t *testing.T) {
	trigger := effects.TreasureToken()
	for _, tc := range []struct{ name, field string }{
		{"token key (Tokens tab)", `"token":"Treasure"`},
		{"scryfall id (Cards tab)", `"scryfall_id":"` + treasurePrinting.ID.String() + `"`},
		{"card name", `"name":"Treasure"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newSpawnFixtureWith(t, predefinedTokenIndex(t, treasurePrinting))
			f.allowSpawn(true)
			got, out := f.spawnOne(tc.field)

			if key, want := game.CatalogKey(got), game.CatalogKey(trigger); key != want {
				t.Errorf("catalog key = %q, want %q (the trigger-made Treasure's); oracle id %q, token key %q",
					key, want, got.OracleID, got.TokenKey)
			}
			if got.OracleID != "" {
				t.Errorf("oracle id = %q: the object is the Scryfall printing, not the predefined token", got.OracleID)
			}
			if got.Name != trigger.Name || got.TypeLine != trigger.TypeLine {
				t.Errorf("spawned %q (%q), want %q (%q)", got.Name, got.TypeLine, trigger.Name, trigger.TypeLine)
			}
			if abs := game.ManaAbilitiesForCard(got); len(abs) != 1 || abs[0].Produced != "{W|U|B|R|G}" {
				t.Errorf("mana abilities = %+v, want the Treasure's one any-colour ability", abs)
			}
			if f.unimplementedInView(got.InstanceID) {
				t.Error("the view flags the spawned Treasure unimplemented")
			}
			if !out.Token || out.ScryfallID != "" || out.Name != "Treasure" {
				t.Errorf("response = %+v, want a token answer naming Treasure with no printing", out)
			}

			// It cracks on its own, with no Goldspan Dragon granting it
			// anything: the sacrifice is paid and a colour pick queued.
			g := f.lobby.RoomOf(f.gameID).Game
			if err := g.ActivateManaAbility(f.alice, got.InstanceID, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("ActivateManaAbility: %v", err)
			}
			g.ReadSnapshot(func() {
				for _, c := range g.Battlefield.Cards {
					if c.InstanceID == got.InstanceID {
						t.Error("the Treasure is still on the battlefield; its own sacrifice-for-mana did not run")
					}
				}
			})
		})
	}
}

// tokenPrinting is a Scryfall token record as the index holds it.
func tokenPrinting(id, oracle, name, typeLine, text string) cards.Card {
	return cards.Card{
		ID:         uuid.MustParse(id),
		OracleID:   uuid.MustParse(oracle),
		Name:       name,
		Layout:     "token",
		SetType:    "token",
		Lang:       "en",
		TypeLine:   typeLine,
		OracleText: text,
	}
}

// The other predefined tokens the catalog has a template for, each from
// its own Scryfall printing (ids and text from the dump), and the
// double-faced Incubator matched on its front face. Each must come out
// as the template under its name, the one the Tokens tab makes.
func TestEveryPredefinedTokenPrintingSpawnsAsItsTemplate(t *testing.T) {
	incubator := cards.Card{
		ID:       uuid.MustParse("d0f953c3-09b9-42c7-816d-59fb0f8b541c"),
		OracleID: uuid.MustParse("dd3bf75e-9ef0-4ded-9b65-1a60a78a6253"),
		Name:     "Incubator // Phyrexian",
		Layout:   "double_faced_token",
		SetType:  "token",
		Lang:     "en",
		TypeLine: "Token Artifact — Incubator // Token Artifact Creature — Phyrexian",
		CardFaces: []cards.CardFace{
			{Name: "Incubator", TypeLine: "Token Artifact — Incubator", OracleText: "{2}: Transform this artifact."},
			{Name: "Phyrexian", TypeLine: "Token Artifact Creature — Phyrexian"},
		},
	}
	printings := []cards.Card{
		tokenPrinting("fe954ed4-af64-4997-ae5d-089eaa439d1f", "a468338f-635e-4206-89d6-72d723071d45", "Food",
			"Token Artifact — Food", "{2}, {T}, Sacrifice this token: You gain 3 life."),
		tokenPrinting("f8b8298e-5ce6-4ff6-9563-fcd158ce841e", "5aa757d8-db7a-4a60-b63e-c9777c141953", "Gold",
			"Token Artifact — Gold", "Sacrifice this token: Add one mana of any color."),
		tokenPrinting("f2c859e1-181e-44d1-afbd-bbd6e52cf42a", "496e1083-c792-40a4-adf4-fec1d559cd5e", "Clue",
			"Token Artifact — Clue", "{2}, Sacrifice this token: Draw a card."),
		tokenPrinting("f90d1006-e8f9-4f62-b58b-6e711129e992", "03f45075-9423-454f-a256-94dcafb2a779", "Blood",
			"Token Artifact — Blood", "{1}, {T}, Discard a card, Sacrifice this token: Draw a card."),
		tokenPrinting("d45fe4b6-aeaf-4f84-b660-c7b482ed8512", "91da73fe-d028-43d7-bf75-f7ef30b45664", "Powerstone",
			"Token Artifact — Powerstone", "{T}: Add {C}. This mana can't be spent to cast a nonartifact spell."),
		tokenPrinting("91613970-c25c-4959-b5b0-b6e2a7657634", "c050f054-1ccf-4819-bc30-a928aed60c56", "Map",
			"Token Artifact — Map", "{1}, {T}, Sacrifice this artifact: Target creature you control explores. Activate only as a sorcery."),
		tokenPrinting("c1d0baf8-db86-4a98-84fa-43aa32e0980f", "d7b177cf-59fd-44dc-9c56-1c4bcacb8468", "Lander",
			"Token Artifact — Lander", "{2}, {T}, Sacrifice this token: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle."),
		tokenPrinting("7f9e9b3a-c515-449a-95da-ee3f5175b74f", "32edf0e6-a495-4de8-b85f-42a2d4ef71b1", "Vibranium",
			"Token Artifact — Vibranium", "Indestructible\n{T}: Add {C}. This mana can't be spent to cast a nonartifact spell."),
		tokenPrinting("eb4bf635-04d7-4994-9f70-4ff573a767e7", "ed1d8118-b182-413b-99ba-52844d35a35b", "Heartwood",
			"Token Artifact — Heartwood", "{T}: Add {R} or {G}."),
		incubator,
	}
	f := newSpawnFixtureWith(t, predefinedTokenIndex(t, printings...))
	f.allowSpawn(true)
	lib := effects.Tokens()
	for _, p := range printings {
		name := p.Name
		if len(p.CardFaces) > 0 {
			name = p.CardFaces[0].Name
		}
		want, ok := lib.PredefinedToken(name)
		if !ok {
			t.Fatalf("no predefined template for %q", name)
		}
		got, out := f.spawnOne(`"scryfall_id":"` + p.ID.String() + `"`)
		if game.CatalogKey(got) != game.CatalogKey(want) || got.OracleID != "" {
			t.Errorf("%s: spawned with catalog key %q (oracle id %q), want the template's %q",
				p.Name, game.CatalogKey(got), got.OracleID, game.CatalogKey(want))
		}
		if got.Name != want.Name || got.TypeLine != want.TypeLine {
			t.Errorf("%s: spawned %q (%q), want %q (%q)", p.Name, got.Name, got.TypeLine, want.Name, want.TypeLine)
		}
		if f.unimplementedInView(got.InstanceID) {
			t.Errorf("%s: the view flags the spawned token unimplemented", p.Name)
		}
		if !out.Token || out.Name != want.Name {
			t.Errorf("%s: response = %+v, want a token answer naming %q", p.Name, out, want.Name)
		}
		if via, ok := lib.TokenTemplate(want.Name); !ok || game.CatalogKey(via) != game.CatalogKey(want) {
			t.Errorf("%s: the Tokens tab does not offer the same template under %q", p.Name, want.Name)
		}
	}
}

// What is NOT swapped. A predefined token the catalog has no template
// for (Junk, Shard) is spawned as its printing, and the table is told
// the engine does not run its text, which is the truth. A record that is
// not a token printing keeps its own identity even when it is named
// "Treasure" (the dump's front_card placeholder of that name).
func TestPrintingsWithNoTemplateStayPrintings(t *testing.T) {
	junk := tokenPrinting("62c4a6f6-6425-4c0c-b35a-880fcab42aad", "7034bca4-8ba0-47a6-886e-ecbd829ddaec", "Junk",
		"Token Artifact — Junk", "{T}, Sacrifice this artifact: Exile the top card of your library. You may play that card this turn. Activate only as a sorcery.")
	shard := tokenPrinting("df826c7d-5508-4e21-848c-91bc3e3f447a", "08734dc6-71d9-46c1-a116-187a92ce3867", "Shard",
		"Token Enchantment — Shard", "{2}, Sacrifice this enchantment: Scry 1, then draw a card.")
	frontCard := cards.Card{
		ID:         uuid.MustParse("5e138adb-32ba-4b63-8964-5447bc353161"),
		OracleID:   uuid.MustParse("3365a79e-3e2d-4f4b-861c-a970ae23345b"),
		Name:       "Treasure",
		Layout:     "front_card",
		SetType:    "memorabilia",
		Lang:       "en",
		TypeLine:   "Card",
		OracleText: "(Theme color: {R})",
	}
	f := newSpawnFixtureWith(t, predefinedTokenIndex(t, junk, shard, frontCard))
	f.allowSpawn(true)

	for _, p := range []cards.Card{junk, shard} {
		got, out := f.spawnOne(`"scryfall_id":"` + p.ID.String() + `"`)
		if got.OracleID != p.OracleID.String() {
			t.Errorf("%s: oracle id %q, want the printing's %q", p.Name, got.OracleID, p.OracleID)
		}
		if !f.unimplementedInView(got.InstanceID) {
			t.Errorf("%s: not flagged, but the engine does not run its text", p.Name)
		}
		if out.Token || out.ScryfallID != p.ID.String() {
			t.Errorf("%s: response = %+v, want the printing echoed back", p.Name, out)
		}
	}

	got, _ := f.spawnOne(`"scryfall_id":"` + frontCard.ID.String() + `"`)
	if got.OracleID != frontCard.OracleID.String() || got.TokenKey != "" {
		t.Errorf("a non-token record named Treasure became oracle %q, token key %q; want it left alone",
			got.OracleID, got.TokenKey)
	}
}

// The dev dock's spawner resolves the same printings, so it makes the
// same template.
func TestDevSpawnOfATreasurePrintingIsThePredefinedToken(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	l := NewLobby(ws.NewRoomManager(log, ""))
	a := auth.NewMemoryAuthenticator()
	srv := httptest.NewServer(Handler(Config{
		Lobby:    l,
		Auth:     a,
		Cards:    predefinedTokenIndex(t, treasurePrinting),
		Tokens:   effects.Tokens(),
		Env:      appenv.EnvDev,
		Features: appenv.LoadFeatures(appenv.EnvDev),
	}))
	t.Cleanup(srv.Close)
	meta, err := l.Create("dev")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, alice, _ := l.Join(meta.ID, meta.InviteToken, "Alice")
	_, bob, _ := l.Join(meta.ID, meta.InviteToken, "Bob")
	uploadDummyDeck(t, l, meta.ID, alice)
	uploadDummyDeck(t, l, meta.ID, bob)
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	tok, _, err := a.Issue(context.Background(), auth.Principal{
		Role: auth.RolePlayer, GameID: meta.ID, PlayerID: alice, Name: "Alice",
	}, time.Hour)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	body := `{"player_id":"` + alice.String() + `","zone":"battlefield","scryfall_id":"` + treasurePrinting.ID.String() + `"}`
	resp := req(t, srv, http.MethodPost, "/games/"+meta.ID.String()+"/dev/spawn", tok, body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("dev spawn: status %d (%s)", resp.StatusCode, errorMessage(t, resp))
	}
	var out devSpawnResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Spawned) != 1 || out.ScryfallID != "" || out.Name != "Treasure" {
		t.Fatalf("response = %+v, want one Treasure with no printing", out)
	}
	g := l.RoomOf(meta.ID).Game
	want := game.CatalogKey(effects.TreasureToken())
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID.String() == out.Spawned[0] && game.CatalogKey(c) != want {
				t.Errorf("dev-spawned Treasure keys on %q, want %q", game.CatalogKey(c), want)
			}
		}
	})
}
