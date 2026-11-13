package game

import (
	"testing"

	"github.com/google/uuid"
)

// minus_counter_payment_test.go — #5746: who put a permanent onto the
// battlefield rides its EventETB (Event.PutBy), and "put a -1/-1 counter
// on a creature you control" is a pay-unless payment
// (PayActionMinusOneCounter). The card side is Thelon's Chant and
// Tourach's Chant (cards/effects).

// etbPutBy is the PutBy of the EventETB that announced `id`.
func etbPutBy(t *testing.T, g *Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	for i := len(g.Events) - 1; i >= 0; i-- {
		if ev := g.Events[i]; ev.Kind == EventETB && ev.CardID == id {
			return ev.PutBy
		}
	}
	t.Fatalf("no EventETB for %s", id)
	return uuid.Nil
}

// EventETB names the player who put the permanent there, with the
// reading the Arboria tally takes: a land play's player, an effect's put
// by the resolving controller even for a permanent that enters under
// someone else, and the player an instruction names inside
// PutByEnteringPlayerForEffect. A token's creator puts it too.
func TestETBEventNamesThePlayerWhoPutThePermanent(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	toMainPhase(t, g)
	land := handPermanentForTest(a, "Wastes", "Basic Land")
	if err := g.CastSpell(a.ID, land, CastSpellParams{}); err != nil {
		t.Fatalf("land play: %v", err)
	}
	if got := etbPutBy(t, g, land); got != a.ID {
		t.Errorf("a land play's PutBy = %s, want its player %s (CR 305.1)", got, a.ID)
	}

	theirs := handPermanentForTest(b, "Bear", "Creature — Bear")
	resolveTriggerForTest(t, g, a, func(g *Game, _ *StackItem) error {
		return g.PutCardsFromHandOntoBattlefieldKeyedForEffect([]uuid.UUID{theirs}, HandEntryOptions{}, KeyedTail{})
	})
	if c := findBattlefieldCard(g, theirs); c == nil || c.Controller != b.ID {
		t.Fatal("setup: the Bear did not enter under its owner")
	}
	if got := etbPutBy(t, g, theirs); got != a.ID {
		t.Errorf("an effect's put: PutBy = %s, want the ability's controller %s (CR 608.2c)", got, a.ID)
	}

	named := handPermanentForTest(b, "Elk", "Creature — Elk")
	resolveTriggerForTest(t, g, a, func(g *Game, _ *StackItem) error {
		return g.PutByEnteringPlayerForEffect(func() error {
			return g.PutCardsFromHandOntoBattlefieldKeyedForEffect([]uuid.UUID{named}, HandEntryOptions{}, KeyedTail{})
		})
	})
	if got := etbPutBy(t, g, named); got != b.ID {
		t.Errorf("an instruction naming the player: PutBy = %s, want %s (CR 110.2a)", got, b.ID)
	}

	resolveTriggerForTest(t, g, a, func(g *Game, _ *StackItem) error {
		return g.CreateTokenForEffect(a.ID, Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1}, 1)
	})
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" {
			token = c.InstanceID
		}
	}
	if got := etbPutBy(t, g, token); got != a.ID {
		t.Errorf("a token's PutBy = %s, want its creator %s", got, a.ID)
	}

	// It rides a snapshot verbatim, as every event fact does.
	_, restored := roundTrip(t, g)
	if got := etbPutBy(t, restored, theirs); got != a.ID {
		t.Errorf("after a restore, PutBy = %s, want %s", got, a.ID)
	}
}

// battlefieldCreatureForTest puts a vanilla creature onto the
// battlefield under p.
func battlefieldCreatureForTest(g *Game, p *Player, name string, power, toughness int) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine, c.Controller = "Creature — Bear", p.ID
	c.Power, c.Toughness = power, toughness
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// The payment's options are the payer's own permanents of the clause's
// kind; paying puts one -1/-1 counter on each named one, and a creature
// that dies of it has still paid (CR 118.12, CR 704.5f).
func TestMinusOneCounterPaymentOffersYourCreaturesAndPutsTheCounter(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	bear := battlefieldCreatureForTest(g, a, "Bear", 2, 2)
	squire := battlefieldCreatureForTest(g, a, "Squire", 1, 1)
	theirs := battlefieldCreatureForTest(g, b, "Their Bear", 2, 2)
	rock := NewCard("Rock", a.ID)
	rock.TypeLine, rock.Controller = "Artifact", a.ID
	g.Battlefield.PushTop(rock)

	pay := &PayAction{Kind: PayActionMinusOneCounter, Count: 1, Of: PermanentQuery{Types: []string{"creature"}}}
	var opts []uuid.UUID
	var can, cannot bool
	g.WithWriteLock(func() {
		opts = g.PayActionOptionsForEffect(a.ID, pay)
		can = g.PayActionCanPayForEffect(a.ID, pay)
		cannot = g.PayActionCanPayForEffect(g.Seats[2].ID, pay)
	})
	if len(opts) != 2 || !containsID(opts, bear) || !containsID(opts, squire) || containsID(opts, theirs) || containsID(opts, rock.InstanceID) {
		t.Fatalf("options = %v, want only a's two creatures", opts)
	}
	if !can || cannot {
		t.Errorf("payable: controller of creatures %v, controller of none %v; want true, false (CR 118.3)", can, cannot)
	}
	if pay.NamesNothing() {
		t.Error("a counter payment names the creature: it is a card pick")
	}

	var err error
	g.WithWriteLock(func() {
		if err = g.validatePayActionLocked(a.ID, pay, []uuid.UUID{theirs}); err == nil {
			return
		}
		if err = g.validatePayActionLocked(a.ID, pay, []uuid.UUID{bear}); err != nil {
			return
		}
		err = g.payActionLocked(a.ID, uuid.Nil, pay, []uuid.UUID{bear}, KeyedTail{})
	})
	if err != nil {
		t.Fatalf("validate or pay: %v (another player's creature must be refused, a's own accepted)", err)
	}
	if c := findBattlefieldCard(g, bear); c == nil || c.Counters[CounterMinusOne] != 1 {
		t.Fatalf("the Bear's -1/-1 counters = %v, want 1", c)
	}

	g.WithWriteLock(func() {
		err = g.payActionLocked(a.ID, uuid.Nil, pay, []uuid.UUID{squire}, KeyedTail{})
		g.runStateChecksLocked()
	})
	if err != nil {
		t.Fatalf("pay with the Squire: %v", err)
	}
	if findBattlefieldCard(g, squire) != nil {
		t.Error("a 1/1 given a -1/-1 counter survived the state-based check")
	}
}
