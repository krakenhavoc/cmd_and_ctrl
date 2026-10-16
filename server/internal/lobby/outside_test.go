package lobby

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// outside_test.go — ADR 0198 §3: the search behind a pick from the card
// database answers the seat that owes the pick, with the cards the pick
// would accept, and nothing else.

type lobbyCollection struct{ cards []game.Card }

func (f lobbyCollection) Search(q string, limit int) []game.Card {
	var out []game.Card
	for _, c := range f.cards {
		if strings.Contains(strings.ToLower(c.Name), strings.ToLower(q)) && len(out) < limit {
			out = append(out, c)
		}
	}
	return out
}

func (f lobbyCollection) Named(name string) (game.Card, bool) {
	for _, c := range f.cards {
		if strings.EqualFold(c.Name, name) {
			return c, true
		}
	}
	return game.Card{}, false
}

func (f lobbyCollection) Suggest(fn func(game.Card) bool) {
	for _, c := range f.cards {
		if !fn(c) {
			return
		}
	}
}

func TestOutsideCardSearchAnswersTheChooserWithMatchingCards(t *testing.T) {
	sorcery := game.Card{Name: "Wishful Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{R}",
		ScryfallID: uuid.NewSHA1(uuid.NameSpaceURL, []byte("ws")).String()}
	instant := game.Card{Name: "Wishful Instant", TypeLine: "Instant", ManaCost: "{U}"}
	prev := game.OutsideCollection
	game.OutsideCollection = lobbyCollection{cards: []game.Card{sorcery, instant}}
	t.Cleanup(func() { game.OutsideCollection = prev })

	f := newPreviewFixture(t)
	g, err := f.lobby.LookupGame(f.gameID)
	if err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if !g.QueueOutsidePickForEffect(game.OutsidePickPrompt{
			Player: f.alice, Question: "Burning Wish", Match: func(c game.Card) bool { return c.IsSorcery() },
		}) {
			t.Fatal("no collection pick for a seat with no sideboard")
		}
	})
	var choiceID uuid.UUID
	g.ReadSnapshot(func() { choiceID = g.PendingChoices[0].ID })

	base := "/games/" + f.gameID.String() + "/outside/cards"
	resp := f.get(base + "?choice=" + choiceID.String() + "&q=wishful")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	var body struct {
		Cards []outsideCardResult `json:"cards"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Cards) != 1 || body.Cards[0].Name != "Wishful Sorcery" || body.Cards[0].ID != sorcery.ScryfallID ||
		body.Cards[0].TypeLine != "Sorcery" || body.Cards[0].ManaCost != "{1}{R}" {
		t.Errorf("rows %+v, want the sorcery alone", body.Cards)
	}

	for path, want := range map[string]int{
		base + "?choice=" + uuid.NewString() + "&q=wishful": http.StatusNotFound,
		base + "?choice=nope&q=wishful":                     http.StatusBadRequest,
	} {
		r := f.get(path)
		r.Body.Close()
		if r.StatusCode != want {
			t.Errorf("GET %s: status %d, want %d", path, r.StatusCode, want)
		}
	}
}
