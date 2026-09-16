package game

import (
	"errors"
	"testing"
)

// exile_zone_exit_test.go — CR 400.7: a card that leaves exile is a
// new object. Its exile-play grant and its counters belong to the old
// object and must not follow it.

// An airbended card moved out of exile by any route other than a
// cast loses the grant, and a later hand cast pays the printed cost
// rather than airbend's {2}.
func TestAirbendGrantDoesNotSurviveLeavingExile(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := permanentFor(g, me, "Airbent Giant", "Creature — Giant", "{4}{G}{G}")
	g.WithWriteLock(func() {
		if err := g.ExileCardWithPermissionForEffect(id, airbendGrant()); err != nil {
			t.Fatalf("ExileCardWithPermissionForEffect: %v", err)
		}
	})
	if !exilePlayOf(g, id).WhileExiled {
		t.Fatalf("setup: airbend grant missing on the exiled card")
	}

	if err := g.MoveCardByID(ZoneRef{Kind: ZoneExile}, ZoneRef{Kind: ZoneHand, Owner: me.ID}, id); err != nil {
		t.Fatalf("MoveCardByID exile -> hand: %v", err)
	}
	var inHand Card
	for _, c := range me.Hand.Cards {
		if c.InstanceID == id {
			inHand = c
		}
	}
	if inHand.InstanceID != id {
		t.Fatalf("card did not reach hand")
	}
	if inHand.ExilePlay != (ExilePlayPermission{}) {
		t.Errorf("exile grant survived leaving exile: %+v", inHand.ExilePlay)
	}

	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
	err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "hand"})
	var im *InsufficientManaError
	if !errors.As(err, &im) {
		t.Fatalf("hand cast on two mana of a {4}{G}{G} card: got %v, want *InsufficientManaError", err)
	}
}

// The price half on its own: a grant still sitting on a card outside
// exile does not reprice the cast. Built by hand, because MoveCard no
// longer lets one get there.
func TestExileGrantPricesOnlyAnExileCast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Stale Grant", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{3}{R}"
	grant := airbendGrant()
	grant.Player = me.ID
	grant.AnyColor = true
	c.ExilePlay = grant
	me.Hand.PushTop(c)

	got := priceOf(t, g, me, c.InstanceID, CastSpellParams{FromZone: "hand"})
	if got.Generic != 3 || len(got.Required) != 1 {
		t.Errorf("hand cast with a stale exile grant: cost %+v, want the printed {3}{R}", got)
	}
}

// Counters placed on a card in exile (by hand today, by suspend
// later) are gone once it leaves.
func TestCountersDoNotSurviveLeavingExile(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	c := NewCard("Suspended Thing", me.ID)
	c.TypeLine = "Creature — Beast"
	c.Counters = map[string]int{"time": 3}
	g.Exile.PushTop(c)

	moved, err := MoveCard(g.Exile, me.Graveyard, c.InstanceID)
	if err != nil {
		t.Fatalf("MoveCard: %v", err)
	}
	if moved.Counters != nil {
		t.Errorf("counters survived leaving exile: %+v", moved.Counters)
	}
}
