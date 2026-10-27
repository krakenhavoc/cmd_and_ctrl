package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0187 §3 (#3570): an exile-from-graveyard link carries pay_cards
// with the chooser's own graveyard as options, and every seat gets them
// — a graveyard is public, unlike a discard's hand.
func TestGraveyardExileLinkShowsItsOptionsToEverySeat(t *testing.T) {
	g := buildActiveGame(t)
	controller := g.Seats[1].ID
	source := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: source, Name: "Carrion Thing", TypeLine: "Creature — Rat", Owner: controller, Controller: controller})
	g.WithWriteLock(func() {
		if err := queueAnyPlayerHooked(g, anyPlayerTestOffer{
			Source: source, Controller: controller, Cost: "exile two cards from their graveyard",
			Question: "Carrion Thing?",
			Action:   &game.PayAction{Kind: game.PayActionExileFromGraveyard, Count: 2},
		}); err != nil {
			t.Fatalf("queueAnyPlayerHooked: %v", err)
		}
	})
	chooser := g.PendingChoices[0].Chooser
	var want []string
	for _, s := range g.Seats {
		if s.ID != chooser {
			continue
		}
		for i := 0; i < 3; i++ {
			c := game.NewCard("Fodder", s.ID)
			c.TypeLine = "Sorcery"
			s.Graveyard.PushTop(c)
			want = append(want, c.InstanceID.String())
		}
	}
	// A card in someone else's graveyard is never an option.
	for _, s := range g.Seats {
		if s.ID != chooser {
			s.Graveyard.PushTop(game.NewCard("Not yours", s.ID))
		}
	}

	for _, viewer := range g.Seats {
		v := ViewOfGameFor(g, viewer.ID.String())
		if len(v.PendingChoices) != 1 || v.PendingChoices[0].PayCards == nil {
			t.Fatalf("viewer %s: no pay_cards on %+v", viewer.Name, v.PendingChoices)
		}
		pc := v.PendingChoices[0].PayCards
		if pc.Action != "exile_graveyard" || pc.Count != 2 {
			t.Errorf("viewer %s: pay_cards %+v", viewer.Name, pc)
		}
		if len(pc.Options) != len(want) {
			t.Fatalf("viewer %s sees %d options, want the chooser's %d", viewer.Name, len(pc.Options), len(want))
		}
		got := map[string]bool{}
		for _, id := range pc.Options {
			got[id] = true
		}
		for _, id := range want {
			if !got[id] {
				t.Errorf("viewer %s: option %s missing", viewer.Name, id)
			}
		}
	}
}
