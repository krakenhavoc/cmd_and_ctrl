package lobby

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
)

// companion_upload_test.go — ADR 0172: a deck's companion is accepted,
// named in the upload answer and seated outside the game; a deck that
// breaks the condition is refused; a lone sideboard companion is a
// warning.

func companionDeckIndex() *cards.Index {
	idx := cards.NewIndex()
	wubrg := []string{"W", "U", "B", "R", "G"}
	legal := map[string]string{"commander": "legal"}
	idx.Put(cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: "Five Colour Commander", TypeLine: "Legendary Creature — Dragon", ManaCost: "{W}{U}{B}{R}{G}", ColorIdentity: wubrg, Legalities: legal})
	idx.Put(cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains", ColorIdentity: []string{"W"}, Legalities: legal})
	idx.Put(cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: "Doubled Spell", TypeLine: "Instant", ManaCost: "{W}{W}", ColorIdentity: []string{"W"}, Legalities: legal})
	idx.Put(cards.Card{
		ID: uuid.New(), OracleID: uuid.MustParse("f0587c55-06c1-4930-a62e-668f37464ce9"),
		Name: "Jegantha, the Wellspring", TypeLine: "Legendary Creature — Elemental Elk", ManaCost: "{4}{R/G}",
		Keywords: []string{"Companion"}, ColorIdentity: wubrg, Legalities: legal,
	})
	return idx
}

func uploadAs(t *testing.T, source string) (int, uploadDeckResponse, string, *Lobby, uuid.UUID, uuid.UUID) {
	t.Helper()
	srv, l, _, _ := newTestHTTPStackWithCards(t, companionDeckIndex())
	meta, _ := l.Create("Companions")
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	var joined sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&joined)
	resp.Body.Close()
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", joined.Token,
		uploadDeckRequest{Format: "text", Source: source, PlayerID: joined.PlayerID})
	raw, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var out uploadDeckResponse
	if resp.StatusCode == http.StatusOK {
		_ = json.Unmarshal(raw, &out)
	}
	return resp.StatusCode, out, string(raw), l, meta.ID, joined.PlayerID
}

func TestUploadDeckWithACompanion(t *testing.T) {
	code, out, raw, l, gameID, playerID := uploadAs(t, "Commander:\n1 Five Colour Commander\nCompanion:\n1 Jegantha, the Wellspring\nMainboard:\n99 Plains\n")
	if code != http.StatusOK {
		t.Fatalf("upload: %d %s", code, raw)
	}
	if out.Companion != "Jegantha, the Wellspring" || out.CardCount != 100 {
		t.Errorf("companion %q, card_count %d", out.Companion, out.CardCount)
	}
	g, err := l.LookupGame(gameID)
	if err != nil {
		t.Fatal(err)
	}
	var seated bool
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			if p.ID == playerID {
				c, ok := p.CompanionCard()
				seated = ok && c.Name == "Jegantha, the Wellspring" && p.Library.Size() == 99
			}
		}
	})
	if !seated {
		t.Error("the companion was not seated outside the game")
	}
}

func TestUploadDeckBreakingTheCompanionConditionIsRefused(t *testing.T) {
	code, _, raw, _, _, _ := uploadAs(t, "Commander:\n1 Five Colour Commander\nCompanion:\n1 Jegantha, the Wellspring\nMainboard:\n1 Doubled Spell\n98 Plains\n")
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("upload: %d %s, want 422", code, raw)
	}
	var body struct {
		Violations []deck.Violation `json:"violations"`
	}
	_ = json.Unmarshal([]byte(raw), &body)
	for _, v := range body.Violations {
		if v.Code == deck.CodeCompanionConditionNotMet && v.Card == "Doubled Spell" {
			return
		}
	}
	t.Errorf("no condition violation naming Doubled Spell: %s", raw)
}

func TestUploadDeckReadsALoneSideboardCompanionWithAWarning(t *testing.T) {
	code, out, raw, _, _, _ := uploadAs(t, "Commander:\n1 Five Colour Commander\nMainboard:\n99 Plains\nSideboard:\n1 Jegantha, the Wellspring\n")
	if code != http.StatusOK {
		t.Fatalf("upload: %d %s", code, raw)
	}
	if out.Companion != "Jegantha, the Wellspring" {
		t.Errorf("companion %q", out.Companion)
	}
	warned := false
	for _, w := range out.Warnings {
		if w.Code == deck.CodeCompanionFromSideboard {
			warned = true
		}
	}
	if !warned {
		t.Errorf("warnings %+v, want %s", out.Warnings, deck.CodeCompanionFromSideboard)
	}
}
