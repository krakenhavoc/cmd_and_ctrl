package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// any_player_offer_test.go — ADR 0150 §6: each link of an any-player
// offer is offered to its chooser only, as an ordinary pay / decline,
// labelled with whose effect the payment is about; and the link blocks
// the table, so nobody else is offered anything.
func TestAnyPlayerLinkIsOfferedToItsChooserOnly(t *testing.T) {
	g := newTable(t)
	controller := g.Seats[1]
	source := uuid.New()
	known := map[uuid.UUID]bool{}
	for _, s := range g.Seats {
		known[s.ID] = true
	}
	g.Battlefield.PushTop(game.Card{InstanceID: source, Name: "Rhystic Thing", TypeLine: "Enchantment", Owner: controller.ID, Controller: controller.ID, KnownBy: known})
	g.WithWriteLock(func() {
		if err := g.QueueAnyPlayerPaysForEffect(game.AnyPlayerPays{
			Source: source, Controller: controller.ID, Cost: "{0}", Question: "Rhystic Thing — pay {0}?",
		}); err != nil {
			t.Fatalf("queue: %v", err)
		}
	})
	asked := g.PendingChoices[0].Chooser
	for _, s := range g.Seats {
		moves := legal.EnumerateFor(g, s.ID)
		if s.ID != asked {
			if len(moves) != 0 {
				t.Errorf("seat %s was offered %v while another seat's link is open", s.Name, labels(moves))
			}
			continue
		}
		var pay, decline bool
		for _, m := range moves {
			switch {
			case strings.Contains(m.Label, "pay {0} to stop "+controller.Name+"'s Rhystic Thing"):
				pay = true
			case strings.Contains(m.Label, "decline (let "+controller.Name+"'s Rhystic Thing happen)"):
				decline = true
			default:
				t.Errorf("unexpected move %q", m.Label)
			}
		}
		if !pay || !decline {
			t.Errorf("the chooser was offered %v; want pay and decline naming the effect", labels(moves))
		}
	}
}
