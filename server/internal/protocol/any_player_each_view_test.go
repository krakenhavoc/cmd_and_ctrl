package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// any_player_each_view_test.go — ADR 0187 §2/§8 (#3572): a leg of a
// per-object any-player offer is a choose_cards that carries
// `any_player_offer` (with `each_life`), `choose_suggested` and
// `earlier_choices` on every seat's view.
func TestPerObjectAnyPlayerOfferLegIsOnEverySeatsView(t *testing.T) {
	g := buildActiveGame(t)
	controller := g.Seats[1]
	source := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: source, Name: "Cleansing", TypeLine: "Sorcery", Owner: controller.ID, Controller: controller.ID})
	var lands [2][]uuid.UUID
	var all []uuid.UUID
	for i := 0; i < 2; i++ {
		for _, name := range []string{"Forest", "Island"} {
			id := uuid.New()
			g.Battlefield.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Basic Land — " + name, Owner: g.Seats[i].ID, Controller: g.Seats[i].ID})
			lands[i] = append(lands[i], id)
			all = append(all, id)
		}
	}
	g.WithWriteLock(func() {
		if err := g.QueueAnyPlayerPaysForEachForEffect(game.AnyPlayerPaysForEach{
			Source: source, Controller: controller.ID, Objects: all, EachLife: 1, Noun: "land",
		}); err != nil {
			t.Fatalf("QueueAnyPlayerPaysForEachForEffect: %v", err)
		}
	})
	first := g.PendingChoices[0]
	if err := g.ResolveChooseCards(first.ID, first.Chooser, []uuid.UUID{lands[0][0]}); err != nil {
		t.Fatalf("seat 0 answer: %v", err)
	}
	for _, viewer := range g.Seats {
		v := ViewOfGameFor(g, viewer.ID.String())
		if len(v.PendingChoices) != 1 {
			t.Fatalf("viewer %s sees %d pending choices, want 1", viewer.Name, len(v.PendingChoices))
		}
		c := v.PendingChoices[0]
		if c.Kind != string(game.PendingChoiceChooseCards) || c.Chooser != g.Seats[1].ID.String() {
			t.Fatalf("viewer %s: prompt %s for %s, want choose_cards for seat 1", viewer.Name, c.Kind, c.Chooser)
		}
		o := c.AnyPlayerOffer
		if o == nil || o.EachLife != 1 || o.Controller != controller.ID.String() || len(o.Order) != len(g.Seats) {
			t.Errorf("viewer %s: offer %+v, want each_life 1, controller seat 1 and the whole order", viewer.Name, o)
		}
		wantOwn := viewer.ID == g.Seats[1].ID
		if wantOwn && len(c.Options) != 4 {
			t.Errorf("viewer %s: %d lands offered, want all 4 (including the one seat 0 chose)", viewer.Name, len(c.Options))
		}
		if len(c.EarlierChoices) != 1 || c.EarlierChoices[0].PlayerID != g.Seats[0].ID.String() ||
			len(c.EarlierChoices[0].CardIDs) != 1 || c.EarlierChoices[0].CardIDs[0] != lands[0][0].String() {
			t.Errorf("viewer %s: earlier_choices %+v, want seat 0 paying for its Forest", viewer.Name, c.EarlierChoices)
		}
		// Seat 1's own lands, none of which an earlier player chose.
		// The suggestion and the options are the chooser's alone.
		switch {
		case wantOwn && (len(c.ChooseSuggested) != 2 || c.ChooseSuggested[0] != lands[1][0].String() || c.ChooseSuggested[1] != lands[1][1].String()):
			t.Errorf("chooser: choose_suggested %v, want seat 1's two lands", c.ChooseSuggested)
		case !wantOwn && (len(c.ChooseSuggested) != 0 || len(c.Options) != 0):
			t.Errorf("viewer %s sees suggestion %v and %d options of another seat's leg", viewer.Name, c.ChooseSuggested, len(c.Options))
		}
	}
	raw, err := json.Marshal(ViewOfGameFor(g, g.Seats[1].ID.String()).PendingChoices[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"each_life":1`) {
		t.Errorf("wire form lacks each_life: %s", raw)
	}
}
