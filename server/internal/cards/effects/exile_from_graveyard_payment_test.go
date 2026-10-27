package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_from_graveyard_payment_test.go — ADR 0187 §3 (#3570): the card
// side of an any-player offer paid by exiling from your graveyard. No
// card uses it yet (Carrion Rats and Carrion Wurm are slice 3), so the
// offer is applied straight from the effect.

func TestExileFromGraveyardPaymentWords(t *testing.T) {
	for n, want := range map[int]string{1: "Exile a card from your graveyard", 3: "Exile three cards from your graveyard"} {
		p := ExileFromGraveyardPayment(n)
		if p.Words != want {
			t.Errorf("n=%d: words %q, want %q", n, p.Words, want)
		}
		if p.Action == nil || p.Action.Kind != game.PayActionExileFromGraveyard || p.Action.Count != n {
			t.Errorf("n=%d: action %+v", n, p.Action)
		}
	}
}

// The offer asks the active player first with the payment's words, a
// payer with too few cards is moved past, and the first player who can
// pay exiles their own cards and ends the offer.
func TestExileFromGraveyardOfferMovesPastAShortGraveyardAndExiles(t *testing.T) {
	g := newCatalogGame(t)
	short, payer := g.Seats[0], g.Seats[1]
	fodder := func(p *game.Player, n int) []uuid.UUID {
		var ids []uuid.UUID
		for i := 0; i < n; i++ {
			c := game.NewCard("Fodder", p.ID)
			c.TypeLine = "Sorcery"
			p.Graveyard.PushTop(c)
			ids = append(ids, c.InstanceID)
		}
		return ids
	}
	fodder(short, 2)
	cards := fodder(payer, 3)
	resetAnyPlayerProbe()
	item := &game.StackItem{Controller: g.Seats[2].ID, SourceCardID: uuid.New()}
	g.WithWriteLock(func() {
		err := AnyPlayerPays{
			Pay:      ExileFromGraveyardPayment(3),
			Question: "Carrion Thing — exile three cards from your graveyard?",
			Then:     testAnyPlayerProbe,
		}.Apply(NewContext(g, item))
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
	link := appOpenLink(g)
	if link == nil || link.Chooser != short.ID {
		t.Fatalf("open link %+v, want the active player's", link)
	}
	if link.PayCost != "exile three cards from your graveyard" {
		t.Errorf("link words %q", link.PayCost)
	}
	if err := g.ResolvePayUnlessWithCards(link.ID, link.Chooser, true, nil); err != nil {
		t.Fatalf("short player's yes: %v", err)
	}
	link = appOpenLink(g)
	if link == nil || link.Chooser != payer.ID {
		t.Fatalf("open link %+v, want seat 1", link)
	}
	if err := g.ResolvePayUnlessWithCards(link.ID, link.Chooser, true, cards); err != nil {
		t.Fatalf("paying: %v", err)
	}
	for _, id := range cards {
		if !g.Exile.Contains(id) {
			t.Errorf("card %v not exiled", id)
		}
	}
	if paid, nobody := testAnyPlayerPaid, testAnyPlayerNobody; paid != 1 || nobody != 0 || appOpenLink(g) != nil {
		t.Errorf("paid %d, nobody %d, open link %v; want 1, 0, none", paid, nobody, appOpenLink(g))
	}
}
