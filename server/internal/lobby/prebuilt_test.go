package lobby

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
)

// prebuilt_test.go covers the human half of the pre-built deck flow:
// GET /decks, and installing one of them through
// POST /games/{id}/decks with `deck` instead of `source`.
//
// These tests use the REAL deck registry rather than a fake, on
// purpose. The whole feature is a claim about those four specific
// decks — that they are legal and that the coverage number shown next
// to each one is this build's — and a fake source would test the
// plumbing while leaving the claim unverified.

// buildPrebuiltDeckIndex synthesises just enough of a Scryfall index
// to resolve one pre-built deck: every declared card, with the colour
// identity the deck declares and a type line that makes basics basic
// and everything else not.
//
// It is not a substitute for the real dump — the
// CMDCTRL_SCRYFALL_DUMP-gated test in internal/decks checks the
// declarations against Scryfall itself. What it buys is that the
// LOBBY's install path is exercised end to end in CI, where there is
// no dump, so a broken `deck` field fails here rather than in
// production.
func buildPrebuiltDeckIndex(t *testing.T, d decks.Deck) *cards.Index {
	t.Helper()
	idx := cards.NewIndex()
	put := func(c decks.Card, commander bool) {
		identity := make([]string, 0, len(c.Identity))
		for _, ch := range c.Identity {
			identity = append(identity, string(ch))
		}
		typeLine := "Artifact"
		switch {
		case c.Basic:
			typeLine = "Basic Land — " + c.Name
		case commander:
			typeLine = "Legendary Creature — Test Commander"
		}
		card := cards.Card{
			ID:            uuid.New(),
			Name:          c.Name,
			TypeLine:      typeLine,
			ColorIdentity: identity,
			Legalities:    map[string]string{"commander": "legal"},
		}
		if c.OracleID != "" {
			if parsed, err := uuid.Parse(c.OracleID); err == nil {
				card.OracleID = parsed
			}
		}
		idx.Put(card)
	}
	put(d.Commander, true)
	for _, c := range d.Mainboard {
		put(c, false)
	}
	return idx
}

func TestPrebuiltDecksListsEveryDeckWithItsCoverage(t *testing.T) {
	srv, _, a, _ := newTestHTTPStackWithCards(t, buildMinimalDeckIndex(t))
	token := adminSession(t, a)

	resp := doGet(t, srv, "/decks", token)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET /decks: got %d (body=%s)", resp.StatusCode, body)
	}
	var out prebuiltDecksResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out.Decks) != len(decks.All()) {
		t.Fatalf("listed %d decks, registry has %d", len(out.Decks), len(decks.All()))
	}
	for i, got := range out.Decks {
		want := decks.All()[i]
		if got.ID != want.ID {
			t.Errorf("deck %d: id %q, want %q (picker order is the registry's)", i, got.ID, want.ID)
		}
		if got.Name == "" || got.Summary == "" || got.Commander == "" {
			t.Errorf("%s: the picker has nothing to render: %+v", got.ID, got)
		}
		if got.CardCount != 100 {
			t.Errorf("%s: card_count %d, want 100", got.ID, got.CardCount)
		}
		// The coverage profile is the point of the route. A zeroed
		// one would render as "0 cards" and quietly promise nothing.
		if got.Coverage.Cards == 0 {
			t.Errorf("%s: no coverage profile", got.ID)
		}
		if got.Coverage.Unregistered != 0 {
			t.Errorf("%s: %d unregistered cards", got.ID, got.Coverage.Unregistered)
		}
		if n := got.Coverage.Full + got.Coverage.Caveats + got.Coverage.Unreviewed; n != got.Coverage.Cards {
			t.Errorf("%s: coverage does not add up: %+v", got.ID, got.Coverage)
		}
		// Caveats without their text is the failure mode this whole
		// route exists to avoid: a number that says "something is
		// imperfect" and nothing a player can weigh.
		if got.Coverage.Caveats > 0 && len(got.Coverage.Imperfect) == 0 {
			t.Errorf("%s: %d caveat cards and no list of them", got.ID, got.Coverage.Caveats)
		}
	}
}

func TestPrebuiltDecksRequiresASession(t *testing.T) {
	srv, _, _, _ := newTestHTTPStackWithCards(t, buildMinimalDeckIndex(t))
	resp := doGet(t, srv, "/decks", "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", resp.StatusCode)
	}
}

// The end-to-end claim: a player picks a deck by ID and their seat is
// holding it, having gone through the same validator a pasted list
// goes through.
func TestUploadDeckInstallsAPrebuiltDeck(t *testing.T) {
	d := decks.All()[0]
	srv, l, _, _ := newTestHTTPStackWithCards(t, buildPrebuiltDeckIndex(t, d))

	meta, _ := l.Create("FNM")
	joined := joinAs(t, srv, meta, "Alice")

	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", joined.Token,
		uploadDeckRequest{Deck: d.ID, PlayerID: joined.PlayerID})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("install %s: got %d (body=%s)", d.ID, resp.StatusCode, body)
	}
	var out uploadDeckResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.CardCount != 100 {
		t.Errorf("card_count %d, want 100", out.CardCount)
	}
	if out.DeckID != d.ID {
		t.Errorf("deck_id %q, want %q", out.DeckID, d.ID)
	}
	// A plain-text decklist carries no name, so without the lobby
	// supplying one the seat would be labelled "deck ready" rather
	// than with the deck the player actually chose.
	if out.DeckName != d.Name {
		t.Errorf("deck_name %q, want %q", out.DeckName, d.Name)
	}
	if len(out.Commanders) != 1 || out.Commanders[0] != d.Commander.Name {
		t.Errorf("commanders %v, want [%s]", out.Commanders, d.Commander.Name)
	}
	// And the seat really holds it — the response is not the state.
	after, err := l.Get(meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	var seated bool
	for _, p := range after.Players {
		if p.PlayerID == joined.PlayerID {
			seated = p.DeckUploaded
		}
	}
	if !seated {
		t.Error("the seat is not holding a deck after a successful pre-built install")
	}
}

// Every pre-built deck has to survive the lobby's own pipeline, not
// only the one in internal/decks. A deck that 422s here is one a
// player can pick and not receive.
func TestEveryPrebuiltDeckInstalls(t *testing.T) {
	for _, d := range decks.All() {
		t.Run(d.ID, func(t *testing.T) {
			srv, l, _, _ := newTestHTTPStackWithCards(t, buildPrebuiltDeckIndex(t, d))
			meta, _ := l.Create("FNM")
			joined := joinAs(t, srv, meta, "Alice")
			resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", joined.Token,
				uploadDeckRequest{Deck: d.ID, PlayerID: joined.PlayerID})
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("got %d (body=%s)", resp.StatusCode, body)
			}
		})
	}
}

func TestUploadDeckRejectsAnUnknownPrebuiltDeck(t *testing.T) {
	srv, l, _, _ := newTestHTTPStackWithCards(t, buildMinimalDeckIndex(t))
	meta, _ := l.Create("FNM")
	joined := joinAs(t, srv, meta, "Alice")

	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", joined.Token,
		uploadDeckRequest{Deck: "no-such-deck", PlayerID: joined.PlayerID})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("got %d, want 422", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	// The error names what IS available: a client sending a deck ID
	// this build does not have should not have to guess.
	if !strings.Contains(string(body), decks.IDs()[0]) {
		t.Errorf("error does not list the available decks: %s", body)
	}
}

func TestUploadDeckRejectsBothDeckAndSource(t *testing.T) {
	srv, l, _, _ := newTestHTTPStackWithCards(t, buildMinimalDeckIndex(t))
	meta, _ := l.Create("FNM")
	joined := joinAs(t, srv, meta, "Alice")

	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", joined.Token,
		uploadDeckRequest{
			Deck:     decks.IDs()[0],
			Source:   "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n",
			PlayerID: joined.PlayerID,
		})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", resp.StatusCode)
	}
}

func TestUploadDeckStillRequiresOneOfThem(t *testing.T) {
	srv, l, _, _ := newTestHTTPStackWithCards(t, buildMinimalDeckIndex(t))
	meta, _ := l.Create("FNM")
	joined := joinAs(t, srv, meta, "Alice")

	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", joined.Token,
		uploadDeckRequest{PlayerID: joined.PlayerID})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", resp.StatusCode)
	}
}

// A pre-built deck must not become a way around the seat-ownership
// rule the pasted path enforces.
func TestPrebuiltDeckObeysSeatOwnership(t *testing.T) {
	d := decks.All()[0]
	srv, l, _, _ := newTestHTTPStackWithCards(t, buildPrebuiltDeckIndex(t, d))
	meta, _ := l.Create("FNM")
	alice := joinAs(t, srv, meta, "Alice")
	bob := joinAs(t, srv, meta, "Bob")

	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", alice.Token,
		uploadDeckRequest{Deck: d.ID, PlayerID: bob.PlayerID})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("got %d, want 403", resp.StatusCode)
	}
}

// Belt and braces on the "one legality path" claim: the bytes the
// lobby sends down the pipeline for a pre-built deck are the deck's
// own decklist text, in the dialect a player could paste.
func TestUploadDeckChoiceHandsOverDecklistText(t *testing.T) {
	d := decks.All()[0]
	format, source, id, err := uploadDeckChoice(uploadDeckRequest{Deck: d.ID})
	if err != nil {
		t.Fatalf("uploadDeckChoice: %v", err)
	}
	if format != "text" {
		t.Errorf("format %q, want text", format)
	}
	if id != d.ID {
		t.Errorf("deck id %q, want %q", id, d.ID)
	}
	if source != d.Decklist() {
		t.Error("the source sent down the pipeline is not the deck's own decklist text")
	}
	if !strings.Contains(source, d.Commander.Name) {
		t.Errorf("decklist does not name the commander: %q", source)
	}
}
