package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const jaredOracle = "b48da54c-002f-4afb-9fdf-2a3e1cf77383"

func jaredBarred(g *game.Game, p uuid.UUID) bool {
	var barred bool
	g.WithWriteLock(func() { barred = g.PlayerCantBecomeMonarchForEffect(p) })
	return barred
}

// #2039: the ETB crowns the opponent and bars its controller for the
// turn; the combat-damage steal then fails, and the bar is gone after
// cleanup.
func TestJaredCarthalionBarsTheCrownForTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	castAndResolveCreature(t, g, "Jared Carthalion, True Heir", "Legendary Creature — Human Warrior", jaredOracle)
	monSettle(t, g)
	pickPlayer(t, g, me.ID, them.ID)
	monSettle(t, g)
	if g.Monarch != them.ID {
		t.Fatalf("monarch = %v, want the opponent %v", g.Monarch, them.ID)
	}
	if !jaredBarred(g, me.ID) || jaredBarred(g, them.ID) {
		t.Fatal("only Jared's controller should be barred")
	}

	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	attackWith(t, g, them.ID, bear)
	monSettle(t, g)
	if g.Monarch != them.ID {
		t.Fatalf("combat damage moved the crown to %v despite the bar", g.Monarch)
	}

	// The manual sandbox set is a table correction and ignores the bar.
	monCrown(t, g, me.ID)
	if g.Monarch != me.ID {
		t.Fatalf("manual SetMonarch refused: monarch = %v", g.Monarch)
	}
	monCrown(t, g, them.ID)

	advanceToUpkeepOf(t, g, 1)
	if jaredBarred(g, me.ID) {
		t.Fatal("the bar outlived the turn")
	}
	g.WithWriteLock(func() {
		if err := g.SetMonarchForEffect(me.ID); err != nil {
			t.Fatal(err)
		}
	})
	if g.Monarch != me.ID {
		t.Fatalf("monarch = %v, want %v after the bar ended", g.Monarch, me.ID)
	}
}

// A card's own "you become the monarch" is refused under the bar.
func TestCantBecomeMonarchRefusesACardEffect(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.CantBecomeMonarchThisTurnForEffect(uuid.Nil, me.ID, "test")
		if err := g.SetMonarchForEffect(me.ID); err != nil {
			t.Fatal(err)
		}
	})
	if g.Monarch != uuid.Nil {
		t.Fatalf("a barred player was crowned by an effect: %v", g.Monarch)
	}
}

// While Jared's controller is the monarch, damage to it is prevented and
// becomes counters; otherwise it is dealt.
func TestJaredCarthalionPreventsDamageOnlyWhileYoureTheMonarch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := apaPush(g, me.ID, me.ID, game.Card{Name: "Jared Carthalion, True Heir", TypeLine: "Legendary Creature — Human Warrior", OracleID: jaredOracle, Power: 3, Toughness: 3})
	a, _, _ := pr8cThree(g, opp.ID)
	pr8Hit(t, g, id, map[uuid.UUID]int{a: 2})
	if pr6Marked(g, id) != 2 {
		t.Fatalf("not the monarch: damage %d, want 2", pr6Marked(g, id))
	}
	monCrown(t, g, me.ID)
	pr8Hit(t, g, id, map[uuid.UUID]int{a: 1})
	if pr6Marked(g, id) != 2 || pr8Counters(g, id, game.CounterPlusOne) != 1 {
		t.Fatalf("monarch: damage %d (want 2), counters %d (want 1)", pr6Marked(g, id), pr8Counters(g, id, game.CounterPlusOne))
	}
}
