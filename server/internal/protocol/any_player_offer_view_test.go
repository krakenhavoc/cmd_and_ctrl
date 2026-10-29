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
		if err := queueAnyPlayerHooked(g, anyPlayerTestOffer{
			Source: source, Controller: controller, Cost: "{2}", Question: "Rhystic Thing — pay {2}?",
		}); err != nil {
			t.Fatalf("queueAnyPlayerHooked: %v", err)
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

// ADR 0150 §3 and §5: a life, damage or no-payment link names its
// payment on the wire (pay_life, pay_damage, pay_nothing) beside the
// card's words for the two answers, and round-trips.
func TestAnyPlayerPaymentKindsAreOnTheWire(t *testing.T) {
	for _, tc := range []struct {
		kind game.PayActionKind
		key  string
	}{
		{game.PayActionLife, `"pay_life":5`},
		{game.PayActionDamage, `"pay_damage":5`},
		{game.PayActionNothing, `"pay_nothing":true`},
	} {
		g := buildActiveGame(t)
		controller := g.Seats[1].ID
		source := uuid.New()
		g.Battlefield.PushTop(game.Card{InstanceID: source, Name: "Thing", TypeLine: "Enchantment", Owner: controller, Controller: controller})
		g.WithWriteLock(func() {
			if err := queueAnyPlayerHooked(g, anyPlayerTestOffer{
				Source: source, Controller: controller, Cost: "5", Question: "Thing?",
				Action:      &game.PayAction{Kind: tc.kind, Count: 5},
				AcceptLabel: "Take it", DeclineLabel: "Don't",
			}); err != nil {
				t.Fatalf("queue: %v", err)
			}
		})
		v := ViewOfGameFor(g, g.Seats[0].ID.String()).PendingChoices[0]
		raw, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(raw), tc.key) || v.AcceptLabel != "Take it" || v.DeclineLabel != "Don't" {
			t.Errorf("%s: wire form %s", tc.kind, raw)
		}
		if v.PayCards != nil || v.PayEnergy != nil {
			t.Errorf("%s: also sent a card or energy payment", tc.kind)
		}
		var back PendingChoiceView
		if err := json.Unmarshal(raw, &back); err != nil {
			t.Fatal(err)
		}
		if (back.PayLife != nil) != (tc.kind == game.PayActionLife) ||
			(back.PayDamage != nil) != (tc.kind == game.PayActionDamage) ||
			back.PayNothing != (tc.kind == game.PayActionNothing) {
			t.Errorf("%s: round trip %+v", tc.kind, back)
		}
	}
}

// Each payment reads as itself in the log.
func TestAnyPlayerAnswerTextNamesThePayment(t *testing.T) {
	for _, tc := range []struct {
		paid        bool
		label, want string
	}{
		{true, "{2}", "Alice paid {2} for X"},
		{false, "5 life", "Alice didn't pay 5 life for X"},
		{true, "6 damage", "Alice took 6 damage from X"},
		{false, "6 damage", "Alice didn't take 6 damage from X"},
		{true, "sacrifice a land", "Alice chose to sacrifice a land for X"},
		{false, "discard three cards", "Alice chose not to discard three cards for X"},
		{true, "sacrifice two lands or take 5 damage", "Alice paid for X (sacrifice two lands or take 5 damage)"},
		{false, "sacrifice two lands or take 5 damage", "Alice didn't pay for X (sacrifice two lands or take 5 damage)"},
		{true, "", "Alice said yes to X"},
		{false, "", "Alice said no to X"},
	} {
		if got := anyPlayerAnswerText(tc.paid, "Alice", tc.label, "X"); got != tc.want {
			t.Errorf("(%v, %q) = %q, want %q", tc.paid, tc.label, got, tc.want)
		}
	}
}
