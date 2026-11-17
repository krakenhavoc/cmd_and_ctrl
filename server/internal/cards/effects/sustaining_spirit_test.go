package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sustaining_spirit_test.go — #5770: a printed "damage that would reduce
// your life total to less than 1 reduces it to 1 instead" on a permanent,
// proved through the engine with the real catalog row.

const sustainingSpiritOracle = "683c81ae-38cd-4db1-b7e7-a09d34b6431e"

func pushSustainingSpirit(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Sustaining Spirit", "Creature — Angel Spirit", sustainingSpiritOracle, 0, 3)
}

// pushLifelinker puts a 2/2 with lifelink onto the battlefield.
func pushLifelinker(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Lifelinker", TypeLine: "Creature — Test",
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		Keywords: []string{"lifelink"},
	})
	return id
}

// The floor holds while the Spirit is out, the damage is dealt in full
// (lifelink gains all of it, the event carries all of it), it is only its
// controller's, and it goes away when the Spirit leaves.
func TestSustainingSpiritFloorsDamageToItsController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	spirit := pushSustainingSpirit(g, me.ID)
	linker := pushLifelinker(g, opp.ID)

	me.Life = 5
	oppLife := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(linker, me.ID, 20) })
	if me.Life != 1 {
		t.Fatalf("life after 20 damage from 5 under the Spirit = %d, want 1", me.Life)
	}
	if ev, ok := lastDamageTo(g, me.ID); !ok || ev.Amount != 20 || ev.DamageNotLifeLoss != 16 {
		t.Errorf("damage event = %d dealt, %d cost no life; want 20 and 16", ev.Amount, ev.DamageNotLifeLoss)
	}
	if opp.Life != oppLife+20 {
		t.Errorf("lifelink gained %d, want the full 20", opp.Life-oppLife)
	}

	// At 1, more damage costs nothing more.
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(linker, me.ID, 3) })
	if me.Life != 1 {
		t.Errorf("life at the floor after more damage = %d, want 1", me.Life)
	}

	// Another seat is not protected.
	other.Life = 2
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(linker, other.ID, 5) })
	if other.Life != -3 {
		t.Errorf("a seat without the Spirit is at %d, want -3", other.Life)
	}

	// The Spirit leaves: the floor goes with it.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(spirit); err != nil {
			t.Fatalf("destroy the Spirit: %v", err)
		}
	})
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(linker, me.ID, 3) })
	if me.Life != -2 {
		t.Errorf("life after the Spirit left = %d, want -2", me.Life)
	}
	if n := len(me.Statics); n != 0 {
		t.Errorf("the printed floor wrote %d entries to Player.Statics; it must write none", n)
	}
}

// Only damage is floored: "lose life" and paying life take the total below
// 1 as normal (the card's 2008-10-01 ruling).
func TestSustainingSpiritDoesNotFloorLifeLossOrLifePayments(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushSustainingSpirit(g, me.ID)

	me.Life = 3
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -5); err != nil {
			t.Fatalf("lose 5 life: %v", err)
		}
	})
	if me.Life != -2 {
		t.Errorf("life after losing 5 from 3 under the Spirit = %d, want -2", me.Life)
	}

	me.Life = 2
	var err error
	g.WithWriteLock(func() { err = g.PayLifeForEffect(uuid.Nil, me.ID, 2) })
	if err != nil {
		t.Fatalf("pay 2 life from 2: %v", err)
	}
	if me.Life != 0 {
		t.Errorf("life after paying 2 from 2 under the Spirit = %d, want 0", me.Life)
	}
}

// The floor is its CONTROLLER's: gaining control of the Spirit moves it.
func TestSustainingSpiritFloorFollowsControl(t *testing.T) {
	g := newCatalogGame(t)
	me, thief := g.Seats[0], g.Seats[1]
	spirit := pushSustainingSpirit(g, me.ID)
	src := pushVanillaCreature(g, g.Seats[2].ID, "Their Burner", 2, 2)

	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), spirit, thief.ID, g.UntilEndOfTurnDuration(), "test — steal") {
			t.Fatal("GainControlForEffect refused the Spirit")
		}
		g.RecomputeLayersIfStaleLocked()
	})
	if c, ok := battlefieldCard(g, spirit); !ok || c.Controller != thief.ID {
		t.Fatal("setup: the Spirit did not change control")
	}

	me.Life, thief.Life = 2, 2
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 5)
		_ = g.DealDamageToPlayerForEffect(src, thief.ID, 5)
	})
	if me.Life != -3 {
		t.Errorf("the Spirit's owner, no longer its controller, is at %d, want -3", me.Life)
	}
	if thief.Life != 1 {
		t.Errorf("the Spirit's new controller is at %d, want 1", thief.Life)
	}
}

// Cumulative upkeep {1}{W}: the first upkeep asks {1}{W}, and declining
// sacrifices the Spirit and takes the floor with it.
func TestSustainingSpiritCumulativeUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	spirit := pushSustainingSpirit(g, owner.ID)
	src := pushVanillaCreature(g, g.Seats[2].ID, "Their Burner", 2, 2)

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := ageCountersOn(t, g, spirit); got != 1 {
		t.Fatalf("age counters %d, want 1", got)
	}
	choice := remoraUpkeepPrompt(t, g, owner.ID)
	parsed, err := game.ParseCost(choice.PayCost)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Generic != 1 || len(parsed.Required) != 1 || !slices.Equal(parsed.Required[0].Options, []string{"W"}) {
		t.Fatalf("upkeep cost %q, want {1}{W}", choice.PayCost)
	}
	if err := g.ResolvePayUnless(choice.ID, owner.ID, false); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(spirit) {
		t.Fatal("the Spirit survived an unpaid upkeep")
	}
	owner.Life = 2
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, owner.ID, 5) })
	if owner.Life != -3 {
		t.Errorf("life after the Spirit was sacrificed = %d, want -3", owner.Life)
	}
}
