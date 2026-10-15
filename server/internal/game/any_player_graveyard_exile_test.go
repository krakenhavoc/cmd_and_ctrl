package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// any_player_graveyard_exile_test.go — ADR 0187 §3 (#3570): an
// any-player offer paid by exiling cards from the payer's own
// graveyard (Carrion Rats, Carrion Wurm).

// graveyardCards puts n plain cards into p's graveyard and returns their IDs.
func graveyardCards(p *Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		c := NewCard("Fodder", p.ID)
		c.TypeLine = "Sorcery"
		p.Graveyard.PushTop(c)
		out = append(out, c.InstanceID)
	}
	return out
}

func exilePayment(n int) func(*AnyPlayerPays) {
	return withPayment(PayAction{Kind: PayActionExileFromGraveyard, Count: n}, "exile cards from their graveyard")
}

// CR 118.3: a graveyard with fewer than Count cards can't pay, so a
// "yes" without picks is a decline and the offer moves on; the next
// player, who has the cards, pays and the cards go to exile.
func TestGraveyardExileNeedsCountCardsAndMovesOnWithout(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	short, payer := g.Seats[0], g.Seats[1]
	shortCards := graveyardCards(short, 2)
	payerCards := graveyardCards(payer, 3)
	var out offerOutcome
	queueOffer(t, g, g.Seats[2].ID, "three", &out, exilePayment(3))

	c := onlyPendingChoice(t, g)
	if a := c.PayAction(); a == nil || a.Kind != PayActionExileFromGraveyard || a.Count != 3 {
		t.Fatalf("the link's payment is %+v, want exile 3 from the graveyard", a)
	}
	var seen []uuid.UUID
	g.WithWriteLock(func() { seen = g.PayActionOptionsForEffect(c.Chooser, c.PayAction()) })
	if len(seen) != 2 {
		t.Fatalf("the short player's options are %d cards, want their 2", len(seen))
	}
	if err := g.ResolvePayUnlessWithCards(c.ID, c.Chooser, true, nil); err != nil {
		t.Fatalf("a yes without picks from a short graveyard: %v, want a decline", err)
	}
	if len(out.paidBy) != 0 || g.Exile.Contains(shortCards[0]) {
		t.Fatalf("the short player's decline paid: %v", out.paidBy)
	}
	c = onlyPendingChoice(t, g)
	if c.Chooser != payer.ID {
		t.Fatalf("asked %v second, want seat 1", c.Chooser)
	}
	if err := g.ResolvePayUnlessWithCards(c.ID, c.Chooser, true, payerCards); err != nil {
		t.Fatalf("paying with 3 cards: %v", err)
	}
	for _, id := range payerCards {
		if !g.Exile.Contains(id) || payer.Graveyard.Contains(id) {
			t.Errorf("card %v was not exiled from the payer's graveyard", id)
		}
	}
	if !sameIDs(out.paidBy, []uuid.UUID{payer.ID}) || out.nobody != 0 {
		t.Errorf("paid by %v, nobody %d; want seat 1 and never", out.paidBy, out.nobody)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts left open after the first payer", len(g.PendingChoices))
	}
}

// Exactly Count distinct cards from the payer's own graveyard are
// accepted; anything else is refused with the prompt kept.
func TestGraveyardExileRefusesWrongPicksAndKeepsThePrompt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, other := g.Seats[0], g.Seats[1]
	mine := graveyardCards(me, 3)
	theirs := graveyardCards(other, 3)
	var out offerOutcome
	queueOffer(t, g, g.Seats[2].ID, "two", &out, exilePayment(2))
	c := onlyPendingChoice(t, g)

	for name, picks := range map[string][]uuid.UUID{
		"another player's card": {mine[0], theirs[0]},
		"a duplicate":           {mine[0], mine[0]},
		"too few":               {mine[0]},
		"too many":              {mine[0], mine[1], mine[2]},
		"a made-up card":        {mine[0], uuid.New()},
	} {
		err := g.ResolvePayUnlessWithCards(c.ID, c.Chooser, true, picks)
		if !errors.Is(err, ErrInvalidParam) {
			t.Errorf("%s: %v, want ErrInvalidParam", name, err)
		}
		if got := onlyPendingChoice(t, g); got.ID != c.ID {
			t.Fatalf("%s: the prompt did not stay", name)
		}
	}
	if len(out.paidBy) != 0 || g.Exile.Contains(mine[0]) || g.Exile.Contains(theirs[0]) {
		t.Fatal("a refused answer moved a card or paid")
	}
	if err := g.ResolvePayUnlessWithCards(c.ID, c.Chooser, true, []uuid.UUID{mine[0], mine[2]}); err != nil {
		t.Fatalf("a legal pair: %v", err)
	}
	if !g.Exile.Contains(mine[0]) || !g.Exile.Contains(mine[2]) || !me.Graveyard.Contains(mine[1]) {
		t.Error("the wrong cards were exiled")
	}
	if !sameIDs(out.paidBy, []uuid.UUID{me.ID}) {
		t.Errorf("paid by %v, want the active player", out.paidBy)
	}
}

// OnPaid runs after the last card has landed, not before: it sees all
// of the cards in exile.
func TestGraveyardExileRunsOnPaidAfterTheLastCardLands(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	cards := graveyardCards(me, 3)
	inExileAtPaid := -1
	var out offerOutcome
	queueOffer(t, g, g.Seats[2].ID, "three", &out, exilePayment(3), func(p *AnyPlayerPays) {
		p.OnPaid = func(g *Game, payer uuid.UUID) error {
			inExileAtPaid = 0
			for _, id := range cards {
				if g.Exile.Contains(id) {
					inExileAtPaid++
				}
			}
			out.paidBy = append(out.paidBy, payer)
			return nil
		}
	})
	c := onlyPendingChoice(t, g)
	if err := g.ResolvePayUnlessWithCards(c.ID, c.Chooser, true, cards); err != nil {
		t.Fatal(err)
	}
	if inExileAtPaid != 3 {
		t.Errorf("OnPaid saw %d of 3 cards in exile", inExileAtPaid)
	}
}

// A commander among the exiled cards lands in exile like any card and
// its owner is offered CR 903.9a afterwards; OnPaid still runs once.
func TestGraveyardExileOffersACommanderTheCommandZoneAndPaysOnce(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	cmd := NewCommander("Test Commander", me.ID)
	cmd.TypeLine = "Legendary Creature — Bear"
	me.Graveyard.PushTop(cmd)
	rest := graveyardCards(me, 1)
	var out offerOutcome
	queueOffer(t, g, g.Seats[2].ID, "two", &out, exilePayment(2))
	c := onlyPendingChoice(t, g)
	if err := g.ResolvePayUnlessWithCards(c.ID, c.Chooser, true, []uuid.UUID{cmd.InstanceID, rest[0]}); err != nil {
		t.Fatal(err)
	}
	if len(out.paidBy) != 1 || out.nobody != 0 {
		t.Fatalf("paid by %v, nobody %d; want exactly one payment", out.paidBy, out.nobody)
	}
	if cardIn(g.Exile, cmd.InstanceID) == nil || cardIn(g.Exile, rest[0]) == nil {
		t.Fatal("not every picked card reached exile")
	}
	runChecks(g)
	prompts := commanderReturnPrompts(g)
	if len(prompts) != 1 || prompts[0].Chooser != me.ID || prompts[0].Source != cmd.InstanceID {
		t.Fatalf("commander_return prompts = %d, want one for the owner", len(prompts))
	}
	if len(out.paidBy) != 1 {
		t.Errorf("OnPaid ran %d times, want once", len(out.paidBy))
	}
}
