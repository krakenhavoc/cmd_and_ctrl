package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// any_player_offer_view_test.go — ADR 0150 §5: an any-player link
// carries `any_player_offer` on every seat's view, and each answer is
// one public log line.

func queueAnyPlayerOfferView(t *testing.T, g *game.Game, controller uuid.UUID) uuid.UUID {
	t.Helper()
	source := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: source, Name: "Rhystic Thing", TypeLine: "Enchantment", Owner: controller, Controller: controller})
	g.WithWriteLock(func() {
		if err := g.QueueAnyPlayerPaysForEffect(game.AnyPlayerPays{
			Source: source, Controller: controller, Cost: "{2}", Question: "Rhystic Thing — pay {2}?",
		}); err != nil {
			t.Fatalf("QueueAnyPlayerPaysForEffect: %v", err)
		}
	})
	return source
}

func TestAnyPlayerOfferIsOnEverySeatsView(t *testing.T) {
	g := buildActiveGame(t)
	controller := g.Seats[1].ID
	queueAnyPlayerOfferView(t, g, controller)
	first := g.PendingChoices[0]
	if err := g.ResolvePayUnless(first.ID, first.Chooser, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	for _, viewer := range g.Seats {
		v := ViewOfGameFor(g, viewer.ID.String())
		if len(v.PendingChoices) != 1 {
			t.Fatalf("viewer %s sees %d pending choices, want 1", viewer.Name, len(v.PendingChoices))
		}
		o := v.PendingChoices[0].AnyPlayerOffer
		if o == nil {
			t.Fatalf("viewer %s: no any_player_offer", viewer.Name)
		}
		if o.Controller != controller.String() || len(o.Order) != len(g.Seats) || o.Order[0] != g.Seats[0].ID.String() {
			t.Errorf("viewer %s: offer %+v", viewer.Name, o)
		}
		if len(o.Declined) != 1 || o.Declined[0] != first.Chooser.String() {
			t.Errorf("viewer %s: declined %v, want the first seat", viewer.Name, o.Declined)
		}
	}
	raw, err := json.Marshal(ViewOfGameFor(g, g.Seats[len(g.Seats)-1].ID.String()).PendingChoices[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"any_player_offer":{"controller":`) {
		t.Errorf("wire form lacks any_player_offer: %s", raw)
	}
	var back PendingChoiceView
	if err := json.Unmarshal(raw, &back); err != nil || back.AnyPlayerOffer == nil || len(back.AnyPlayerOffer.Declined) != 1 {
		t.Errorf("round trip lost the offer: %+v (%v)", back.AnyPlayerOffer, err)
	}
}

// Every answer is a public log line naming the seat, the payment and
// the card.
func TestAnyPlayerAnswersAreLogLines(t *testing.T) {
	g := buildActiveGame(t)
	payer := g.Seats[1]
	payer.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	queueAnyPlayerOfferView(t, g, g.Seats[0].ID)
	c := g.PendingChoices[0]
	if err := g.ResolvePayUnless(c.ID, c.Chooser, false); err != nil {
		t.Fatal(err)
	}
	c = g.PendingChoices[0]
	if err := g.ResolvePayUnless(c.ID, payer.ID, true); err != nil {
		t.Fatal(err)
	}
	var declined, paid bool
	for _, e := range ViewOfGame(g).Log {
		switch e.Kind {
		case LogAnyPlayerDeclined:
			declined = strings.Contains(e.Text, "didn't pay {2} for Rhystic Thing")
		case LogAnyPlayerPaid:
			paid = strings.Contains(e.Text, payer.Name+" paid {2} for Rhystic Thing")
		}
	}
	if !declined || !paid {
		t.Errorf("log lines: declined %v, paid %v", declined, paid)
	}
}
