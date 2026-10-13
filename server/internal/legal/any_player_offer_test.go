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
// ADR 0150 §3: a life or damage link offers "Pay" only when the engine
// would take it, names the payment ("pay 5 life", "take 6 damage") and
// carries a life price on the move's cost.
func TestAnyPlayerLifeAndDamageLinksNameTheirPayment(t *testing.T) {
	for _, tc := range []struct {
		kind  game.PayActionKind
		words string
		life  int
		want  string
		pays  bool
	}{
		{game.PayActionLife, "5 life", 40, "pay 5 life to stop", true},
		{game.PayActionLife, "5 life", 4, "", false},
		{game.PayActionDamage, "6 damage", 1, "take 6 damage to stop", true},
	} {
		g := newTable(t)
		controller := g.Seats[1]
		source := uuid.New()
		known := map[uuid.UUID]bool{}
		for _, s := range g.Seats {
			known[s.ID] = true
		}
		g.Battlefield.PushTop(game.Card{InstanceID: source, Name: "Thing", TypeLine: "Enchantment", Owner: controller.ID, Controller: controller.ID, KnownBy: known})
		g.WithWriteLock(func() {
			if err := g.QueueAnyPlayerPaysForEffect(game.AnyPlayerPays{
				Source: source, Controller: controller.ID, Cost: tc.words, Question: "Thing?",
				Action: &game.PayAction{Kind: tc.kind, Count: 5},
			}); err != nil {
				t.Fatalf("queue: %v", err)
			}
		})
		c := g.PendingChoices[0]
		for _, s := range g.Seats {
			if s.ID == c.Chooser {
				s.Life = tc.life
			}
		}
		var pay *legal.Move
		moves := legal.EnumerateFor(g, c.Chooser)
		for i := range moves {
			if strings.Contains(moves[i].Label, "to stop") {
				pay = &moves[i]
			}
		}
		if (pay != nil) != tc.pays {
			t.Fatalf("%s at %d life: offered %v", tc.words, tc.life, labels(moves))
		}
		if pay == nil {
			continue
		}
		if !strings.Contains(pay.Label, tc.want) {
			t.Errorf("%s: pay move %q, want it to contain %q", tc.words, pay.Label, tc.want)
		}
		if tc.kind == game.PayActionLife && (pay.Cost == nil || pay.Cost.Life != 5) {
			t.Errorf("%s: the move's cost is %+v, want 5 life", tc.words, pay.Cost)
		}
	}
}

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
