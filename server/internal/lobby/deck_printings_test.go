package lobby

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decklibrary"
)

func TestDeckPrintingPickerPreservesDeckAndRejectsWrongCards(t *testing.T) {
	st, owner, tok, idx := manageSetup(t)
	plains, _ := idx.FindByName("Plains")
	plains.OracleID, plains.SetCode, plains.CollectorNo, plains.Lang = uuid.New(), "one", "1", "en"
	idx.Put(plains)
	alternate := plains
	alternate.ID, alternate.SetCode, alternate.CollectorNo = uuid.New(), "two", "2"
	idx.Put(alternate)
	commander, _ := idx.FindByName("Test Commander")
	commander.SetCode, commander.CollectorNo = "cmd", "1"
	idx.Put(commander)
	text := "Commander:\n1 Test Commander\nCompanion:\n1 Unknown Companion\nMainboard:\n98 Plains (ONE) 1\n1 Unknown Card\nSideboard:\n1 Unknown Sideboard\n"
	d, err := st.library.UpsertFromLink(context.Background(), owner, "Squall test", "text", text, "https://moxfield.com/decks/example", []string{"Test Commander"}, 100)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.library.SetPowerLevel(context.Background(), d.ID, 3.5); err != nil {
		t.Fatal(err)
	}
	path := "/me/decks/" + d.ID.String() + "/printings"
	resp := doGet(t, st.srv, path, tok)
	var picker deckPrintingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&picker); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 || len(picker.Cards) != 5 {
		t.Fatalf("picker %d: %+v", resp.StatusCode, picker)
	}
	row := -1
	for _, c := range picker.Cards {
		if c.Name == "Plains" {
			row = c.Entry
			if c.PrintingID != plains.ID {
				t.Fatal("selected source printing lost")
			}
		}
	}
	if row < 0 {
		t.Fatal("no Plains row")
	}
	body := func(id uuid.UUID, revision string) any {
		return map[string]any{"revision": revision, "choices": []any{map[string]any{"entry": row, "printing_id": id}}}
	}
	for _, tc := range []struct {
		name   string
		body   any
		status int
	}{
		{"other card", body(commander.ID, picker.Revision), 400},
		{"unknown printing", body(uuid.New(), picker.Revision), 400},
		{"stale revision", body(alternate.ID, "stale"), 409},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := doJSON(t, st.srv, "PUT", path, tok, tc.body)
			defer resp.Body.Close()
			if resp.StatusCode != tc.status {
				t.Fatalf("got %d", resp.StatusCode)
			}
		})
	}
	resp = doJSON(t, st.srv, "PUT", path, tok, body(alternate.ID, picker.Revision))
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("save %d", resp.StatusCode)
	}
	saved, err := st.library.Get(context.Background(), d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.SourceURL != d.SourceURL || saved.Name != d.Name || saved.CardCount != d.CardCount || saved.PowerLevel != 3.5 || saved.Commanders[0] != d.Commanders[0] {
		t.Fatalf("art selection changed deck metadata: %+v", saved)
	}
	entries, err := deck.ParseText(saved.SourceText)
	if err != nil || len(entries) != 5 {
		t.Fatalf("lost cards: %v %s", err, saved.SourceText)
	}
	for _, e := range entries {
		switch e.Name {
		case "Plains":
			if e.Count != 98 || e.SetCode != "TWO" || e.CollectorNo != "2" {
				t.Fatalf("wrong Plains: %+v", e)
			}
		case "Test Commander":
			if !e.IsCommander {
				t.Fatal("lost commander")
			}
		case "Unknown Companion":
			if !e.IsCompanion {
				t.Fatal("lost companion")
			}
		case "Unknown Sideboard":
			if !e.IsSideboard {
				t.Fatal("lost sideboard")
			}
		}
	}
	resp = doJSON(t, st.srv, "PUT", path, tok, body(plains.ID, picker.Revision))
	resp.Body.Close()
	if resp.StatusCode != 409 {
		t.Fatalf("stale picker overwrote saved printings: %d", resp.StatusCode)
	}
	other := mustLibraryUser(t, st.db, "Other")
	theirs := saveText(t, st, other, "Other deck")
	for _, method := range []string{"GET", "PUT"} {
		resp := doJSON(t, st.srv, method, "/me/decks/"+theirs.ID.String()+"/printings", tok, body(alternate.ID, picker.Revision))
		resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("%s other owner: %d", method, resp.StatusCode)
		}
		resp = doJSON(t, st.srv, method, path, guestToken(t, st), body(alternate.ID, picker.Revision))
		resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("%s guest: %d", method, resp.StatusCode)
		}
	}
}

func TestCommanderAliasUsesTheExplicitPrintingInLibraryArt(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	commander, _ := idx.FindByName("Test Commander")
	commander.OracleID, commander.FlavorName, commander.SetCode, commander.CollectorNo = uuid.New(), "Alternate Commander", "one", "1"
	commander.ImageURIs = map[string]string{"normal": "https://example.test/one.jpg"}
	idx.Put(commander)
	other := commander
	other.ID, other.FlavorName, other.SetCode, other.CollectorNo = uuid.New(), "", "two", "2"
	idx.Put(other)
	d := decklibrary.Deck{Commanders: []string{"Alternate Commander"}, SourceFormat: "text", SourceText: "Commander:\n1 Test Commander (TWO) 2\nMainboard:\n99 Plains\n"}
	got := libraryDeckInfo(Config{Cards: idx}, d, nil)
	if len(got.CommanderCards) != 1 || got.CommanderCards[0].ScryfallID != other.ID.String() {
		t.Fatalf("commander alias ignored chosen printing: %+v", got.CommanderCards)
	}
}
