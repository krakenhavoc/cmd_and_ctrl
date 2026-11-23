package testsa

import (
	"testing"

	"github.com/google/uuid"

	cardsn "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/n"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr8_cards_test.go — ADR 0108 PR 8 (#1906): the cards a
// prevention static's additional effect (ReplacementEffect.Then /
// ThenPer) and the scoped shields' Then (owner decision 2) unblock. The
// shared helpers here drive damage the way the rulings talk about it: from
// several sources AT ONCE (one damage instance), and as damage that can't
// be prevented (CR 615.12).

const (
	pr8PhantomCentaur = "9bb8c54b-1228-4b7c-8651-52cb5b0f6e72"
	pr8NineLives      = "236e1f57-7ef5-455a-82a6-8ff6b85d8849"
	pr8TestOfFaith    = "3397aa3d-bf73-4ca3-a806-059361603079"
)

// Phantom Centaur: blocked by three at once, all of it is prevented and
// one counter comes off; a black source meets protection first and costs
// no counter; unpreventable damage is dealt AND still costs a counter, so
// a 4/2 dealt 1 becomes a 3/1 with 1 damage and dies (the rulings).
func TestADR0108PR8PhantomCentaur(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	centaur := effectstest.Pr8Enter(t, g, me, "Phantom Centaur", "Creature — Centaur Spirit", pr8PhantomCentaur)
	if n := effectstest.Pr8Counters(g, centaur, game.CounterPlusOne); n != 3 {
		t.Fatalf("entered with %d counters, want 3", n)
	}
	a := effectstest.Pr7Creature(g, opp.ID, "A", 2, "R")
	b := effectstest.Pr7Creature(g, opp.ID, "B", 3, "G")
	c := effectstest.Pr7Creature(g, opp.ID, "C", 4, "W")
	effectstest.Pr8Hit(t, g, centaur, map[uuid.UUID]int{a: 2, b: 3, c: 4})
	if effectstest.Pr6Marked(g, centaur) != 0 || effectstest.Pr8Counters(g, centaur, game.CounterPlusOne) != 2 {
		t.Fatalf("after three at once: damage %d (want 0), counters %d (want 2)",
			effectstest.Pr6Marked(g, centaur), effectstest.Pr8Counters(g, centaur, game.CounterPlusOne))
	}
	black := effectstest.Pr7Creature(g, opp.ID, "Black", 2, "B")
	effectstest.Pr8Hit(t, g, centaur, map[uuid.UUID]int{black: 2})
	if effectstest.Pr6Marked(g, centaur) != 0 || effectstest.Pr8Counters(g, centaur, game.CounterPlusOne) != 2 {
		t.Fatalf("protection from black: damage %d (want 0), counters %d (want 2)",
			effectstest.Pr6Marked(g, centaur), effectstest.Pr8Counters(g, centaur, game.CounterPlusOne))
	}
	effectstest.Pr8Unpreventable(t, g, a, centaur, 1)
	if effectstest.FindBattlefieldCardForTest(g, centaur) != nil {
		t.Fatalf("unpreventable 1: still on the battlefield with %d damage and %d counters; want a 3/1 with 1 damage, dead",
			effectstest.Pr6Marked(g, centaur), effectstest.Pr8Counters(g, centaur, game.CounterPlusOne))
	}
}

// Nine Lives: three sources at once are three counters (the ruling, even
// past nine), damage that can't be prevented still adds one, nine exiles
// it, and its leaving loses you the game.
func TestADR0108PR8NineLives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nine := effectstest.PushCatalogPermanent(g, me.ID, "Nine Lives", "Enchantment", pr8NineLives, false)
	a := effectstest.Pr7Creature(g, opp.ID, "A", 2, "R")
	b := effectstest.Pr7Creature(g, opp.ID, "B", 3, "G")
	c := effectstest.Pr7Creature(g, opp.ID, "C", 4, "W")
	life := me.Life
	effectstest.Pr8Hit(t, g, me.ID, map[uuid.UUID]int{a: 2, b: 3, c: 4})
	if me.Life != life || effectstest.Pr8Counters(g, nine, cardsn.NineLivesCounter) != 3 {
		t.Fatalf("life %d (want %d), counters %d (want 3: one per source)", me.Life, life, effectstest.Pr8Counters(g, nine, cardsn.NineLivesCounter))
	}
	effectstest.Pr7Hit(t, g, a, me.ID, 1)
	effectstest.Pr7Hit(t, g, a, me.ID, 1)
	if effectstest.Pr8Counters(g, nine, cardsn.NineLivesCounter) != 5 {
		t.Fatalf("two instances from one source: counters %d, want 5", effectstest.Pr8Counters(g, nine, cardsn.NineLivesCounter))
	}
	effectstest.Pr8Unpreventable(t, g, a, me.ID, 2)
	if me.Life != life-2 || effectstest.Pr8Counters(g, nine, cardsn.NineLivesCounter) != 6 {
		t.Fatalf("unpreventable: life %d (want %d), counters %d (want 6)", me.Life, life-2, effectstest.Pr8Counters(g, nine, cardsn.NineLivesCounter))
	}
	effectstest.Pr8Hit(t, g, me.ID, map[uuid.UUID]int{a: 1, b: 1, c: 1})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, nine) != nil {
		t.Fatal("nine incarnation counters: Nine Lives is still on the battlefield")
	}
	if !me.Eliminated {
		t.Fatal("Nine Lives left the battlefield and its controller is still in the game")
	}
}

// Test of Faith: 3 of 6 is prevented, three counters go on before lethal
// damage is checked (the ruling: a 1/1 ends up a 4/4 with 3 damage), and
// the follow-up under unpreventable damage puts on none.
func TestADR0108PR8TestOfFaith(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 6, "R")
	effectstest.CastCatalogSpell(t, g, "Test of Faith", "Instant", pr8TestOfFaith, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr6Damage(t, g, src, bear, 6)
	if effectstest.FindBattlefieldCardForTest(g, bear) == nil {
		t.Fatal("the bear died: the counters go on before lethal damage is checked")
	}
	if effectstest.Pr6Marked(g, bear) != 3 || effectstest.Pr8Counters(g, bear, game.CounterPlusOne) != 3 {
		t.Fatalf("bear: damage %d (want 3), counters %d (want 3)", effectstest.Pr6Marked(g, bear), effectstest.Pr8Counters(g, bear, game.CounterPlusOne))
	}
}

// Test of Faith under damage that can't be prevented: no counters, and the
// shield keeps its charge for the next damage.
func TestADR0108PR8TestOfFaithUnderUnpreventableDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 4})
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 6, "R")
	effectstest.CastCatalogSpell(t, g, "Test of Faith", "Instant", pr8TestOfFaith, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr8Unpreventable(t, g, src, bear, 2)
	if effectstest.Pr6Marked(g, bear) != 2 || effectstest.Pr8Counters(g, bear, game.CounterPlusOne) != 0 {
		t.Fatalf("unpreventable: damage %d (want 2), counters %d (want 0)", effectstest.Pr6Marked(g, bear), effectstest.Pr8Counters(g, bear, game.CounterPlusOne))
	}
	effectstest.Pr6Damage(t, g, src, bear, 1)
	if effectstest.Pr6Marked(g, bear) != 2 || effectstest.Pr8Counters(g, bear, game.CounterPlusOne) != 1 {
		t.Fatalf("then 1 more: damage %d (want 2), counters %d (want 1)", effectstest.Pr6Marked(g, bear), effectstest.Pr8Counters(g, bear, game.CounterPlusOne))
	}
}
