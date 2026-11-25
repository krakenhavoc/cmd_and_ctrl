package game

import (
	"testing"

	"github.com/google/uuid"
)

// life_payment_event_test.go — #6234, Font of Agonies: "Whenever you pay
// life, put that many blood counters on this enchantment." CR 119.4: "If
// a player pays life, the payment is subtracted from their life total;
// in other words, the player loses that much life." Every payment goes
// through payLifeAsCostLocked, whose EventChangeLife carries LifePayment;
// life lost any other way does not.

func lifeChanges(g *Game, from int, player uuid.UUID) []Event {
	var out []Event
	for _, ev := range g.Events[from:] {
		if ev.Kind == EventChangeLife && ev.Target == player {
			out = append(out, ev)
		}
	}
	return out
}

// Phyrexian mana paid with 2 life is a life payment of 2.
func TestPhyrexianLifeIsALifePayment(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := compleatedSage(t, g, me)
	for _, land := range []struct{ name, typeLine string }{
		{"Forest", "Basic Land — Forest"},
		{"Island", "Basic Land — Island"},
		{"Plains", "Basic Land — Plains"},
		{"Swamp", "Basic Land — Swamp"},
	} {
		pushBattlefieldForTest(g, me.ID, land.name, land.typeLine, "")
	}
	start := len(g.Events)
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, AutoTap: true, FromZone: "hand", PhyrexianLife: 1}); err != nil {
		t.Fatal(err)
	}
	evs := lifeChanges(g, start, me.ID)
	if len(evs) != 1 || !evs[0].LifePayment || evs[0].Amount != -PhyrexianLifePerSymbol {
		t.Fatalf("life events = %+v, want one payment of %d", evs, PhyrexianLifePerSymbol)
	}
}

// The door every activated ability's "Pay N life" cost goes through
// (activated.go, mutations.go, player_mana_ability.go) marks the
// payment; damage and an effect's life loss do not.
func TestOnlyAPaymentIsMarkedAsLifePaid(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	start := len(g.Events)
	g.WithWriteLock(func() {
		if err := g.PayLifeForEffect(uuid.Nil, me, 3); err != nil {
			t.Fatal(err)
		}
		_ = g.DealDamageToPlayerForEffect(g.Seats[1].ID, me, 2)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me, -1)
	})
	evs := lifeChanges(g, start, me)
	if len(evs) < 2 {
		t.Fatalf("life events = %d, want the payment and the losses", len(evs))
	}
	if !evs[0].LifePayment || evs[0].Amount != -3 {
		t.Errorf("the payment = amount %d paid=%v, want a life payment of 3", evs[0].Amount, evs[0].LifePayment)
	}
	for _, ev := range evs[1:] {
		if ev.LifePayment {
			t.Errorf("a loss of %d that was not paid is marked as paid", ev.Amount)
		}
	}
}
