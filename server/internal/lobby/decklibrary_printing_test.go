package lobby

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// TestMyDecksShowsTheCommandersChosenPrinting pins ADR 0192 §3: the
// Decks page draws the printing a saved deck's source names for its
// commander, not the index's default for the name.
func TestMyDecksShowsTheCommandersChosenPrinting(t *testing.T) {
	st, owner, tok, idx := manageSetup(t)
	oracle := uuid.New()
	plain, lair := uuid.New(), uuid.New()
	art := map[string]string{"art_crop": "https://example.test/a.jpg"}
	idx.Put(cards.Card{ID: plain, OracleID: oracle, Name: "Raffine, Scheming Seer", SetCode: "snc", CollectorNo: "213",
		SetType: "expansion", ReleasedAt: "2022-04-29", ImageURIs: art})
	idx.Put(cards.Card{ID: lair, OracleID: oracle, Name: "Raffine, Scheming Seer", SetCode: "slz", CollectorNo: "93",
		SetType: "box", ReleasedAt: "2026-09-02", ImageURIs: art})
	ctx := context.Background()
	if _, err := st.library.Upsert(ctx, owner, "Default", "text", "Commander:\n1 Raffine, Scheming Seer\n",
		[]string{"Raffine, Scheming Seer"}, 100); err != nil {
		t.Fatal(err)
	}
	if _, err := st.library.Upsert(ctx, owner, "Chosen", "text", "Commander:\n1 Raffine, Scheming Seer (SLZ) 93\n",
		[]string{"Raffine, Scheming Seer"}, 100); err != nil {
		t.Fatal(err)
	}

	resp := doGet(t, st.srv, "/me/decks", tok)
	defer resp.Body.Close()
	var got myDecksResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil || resp.StatusCode != 200 {
		t.Fatalf("status %d, decode %v", resp.StatusCode, err)
	}
	ids := map[string]string{}
	for _, d := range got.Decks {
		if len(d.CommanderCards) == 1 {
			ids[d.Name] = d.CommanderCards[0].ScryfallID
		}
	}
	if ids["Default"] != plain.String() {
		t.Errorf("a deck naming no printing drew %s, want the ordinary-set one %s", ids["Default"], plain)
	}
	if ids["Chosen"] != lair.String() {
		t.Errorf("a deck naming (SLZ) 93 drew %s, want %s", ids["Chosen"], lair)
	}
}
