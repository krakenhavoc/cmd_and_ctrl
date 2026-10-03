package effects

import (
	"testing"

	"github.com/google/uuid"

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

// pr8Hit is one damage instruction (one instance): each source deals its
// amount to `to`, all at once, then the priority boundary runs.
func pr8Hit(t *testing.T, g *game.Game, to uuid.UUID, hits map[uuid.UUID]int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DamageInstanceForEffect(func() error {
			for src, n := range hits {
				if err := g.DealMarkedDamageForEffect(src, nil, to, n, game.DamageMarks{}); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
}

// pr8Unpreventable deals `n` damage that can't be prevented (CR 615.12).
func pr8Unpreventable(t *testing.T, g *game.Game, src, to uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealMarkedDamageForEffect(src, nil, to, n, game.DamageMarks{CantBePrevented: true}); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
}

// pr8Counters is the named counters on a battlefield permanent.
func pr8Counters(g *game.Game, id uuid.UUID, kind string) int {
	c := findBattlefieldCardForTest(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[kind]
}

// pr8Enter puts a catalogued permanent onto the battlefield through the
// entry pipeline, so "enters with N counters" applies.
func pr8Enter(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, owner.ID)
	c.TypeLine, c.OracleID = typeLine, oracle
	owner.Hand.PushTop(c)
	var id uuid.UUID
	g.WithWriteLock(func() {
		var err error
		id, err = g.PutFromHandOntoBattlefieldForEffect(c.InstanceID, game.HandEntryOptions{Controller: owner.ID})
		if err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	return id
}

// Phantom Centaur: blocked by three at once, all of it is prevented and
// one counter comes off; a black source meets protection first and costs
// no counter; unpreventable damage is dealt AND still costs a counter, so
// a 4/2 dealt 1 becomes a 3/1 with 1 damage and dies (the rulings).
func TestADR0108PR8PhantomCentaur(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	centaur := pr8Enter(t, g, me, "Phantom Centaur", "Creature — Centaur Spirit", pr8PhantomCentaur)
	if n := pr8Counters(g, centaur, game.CounterPlusOne); n != 3 {
		t.Fatalf("entered with %d counters, want 3", n)
	}
	a := pr7Creature(g, opp.ID, "A", 2, "R")
	b := pr7Creature(g, opp.ID, "B", 3, "G")
	c := pr7Creature(g, opp.ID, "C", 4, "W")
	pr8Hit(t, g, centaur, map[uuid.UUID]int{a: 2, b: 3, c: 4})
	if pr6Marked(g, centaur) != 0 || pr8Counters(g, centaur, game.CounterPlusOne) != 2 {
		t.Fatalf("after three at once: damage %d (want 0), counters %d (want 2)",
			pr6Marked(g, centaur), pr8Counters(g, centaur, game.CounterPlusOne))
	}
	black := pr7Creature(g, opp.ID, "Black", 2, "B")
	pr8Hit(t, g, centaur, map[uuid.UUID]int{black: 2})
	if pr6Marked(g, centaur) != 0 || pr8Counters(g, centaur, game.CounterPlusOne) != 2 {
		t.Fatalf("protection from black: damage %d (want 0), counters %d (want 2)",
			pr6Marked(g, centaur), pr8Counters(g, centaur, game.CounterPlusOne))
	}
	pr8Unpreventable(t, g, a, centaur, 1)
	if findBattlefieldCardForTest(g, centaur) != nil {
		t.Fatalf("unpreventable 1: still on the battlefield with %d damage and %d counters; want a 3/1 with 1 damage, dead",
			pr6Marked(g, centaur), pr8Counters(g, centaur, game.CounterPlusOne))
	}
}

// Nine Lives: three sources at once are three counters (the ruling, even
// past nine), damage that can't be prevented still adds one, nine exiles
// it, and its leaving loses you the game.
func TestADR0108PR8NineLives(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nine := pushCatalogPermanent(g, me.ID, "Nine Lives", "Enchantment", pr8NineLives, false)
	a := pr7Creature(g, opp.ID, "A", 2, "R")
	b := pr7Creature(g, opp.ID, "B", 3, "G")
	c := pr7Creature(g, opp.ID, "C", 4, "W")
	life := me.Life
	pr8Hit(t, g, me.ID, map[uuid.UUID]int{a: 2, b: 3, c: 4})
	if me.Life != life || pr8Counters(g, nine, nineLivesCounter) != 3 {
		t.Fatalf("life %d (want %d), counters %d (want 3: one per source)", me.Life, life, pr8Counters(g, nine, nineLivesCounter))
	}
	pr7Hit(t, g, a, me.ID, 1)
	pr7Hit(t, g, a, me.ID, 1)
	if pr8Counters(g, nine, nineLivesCounter) != 5 {
		t.Fatalf("two instances from one source: counters %d, want 5", pr8Counters(g, nine, nineLivesCounter))
	}
	pr8Unpreventable(t, g, a, me.ID, 2)
	if me.Life != life-2 || pr8Counters(g, nine, nineLivesCounter) != 6 {
		t.Fatalf("unpreventable: life %d (want %d), counters %d (want 6)", me.Life, life-2, pr8Counters(g, nine, nineLivesCounter))
	}
	pr8Hit(t, g, me.ID, map[uuid.UUID]int{a: 1, b: 1, c: 1})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, nine) != nil {
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	src := pr7Creature(g, opp.ID, "Src", 6, "R")
	castCatalogSpell(t, g, "Test of Faith", "Instant", pr8TestOfFaith, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	pr6Damage(t, g, src, bear, 6)
	if findBattlefieldCardForTest(g, bear) == nil {
		t.Fatal("the bear died: the counters go on before lethal damage is checked")
	}
	if pr6Marked(g, bear) != 3 || pr8Counters(g, bear, game.CounterPlusOne) != 3 {
		t.Fatalf("bear: damage %d (want 3), counters %d (want 3)", pr6Marked(g, bear), pr8Counters(g, bear, game.CounterPlusOne))
	}
}

// Test of Faith under damage that can't be prevented: no counters, and the
// shield keeps its charge for the next damage.
func TestADR0108PR8TestOfFaithUnderUnpreventableDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 4})
	src := pr7Creature(g, opp.ID, "Src", 6, "R")
	castCatalogSpell(t, g, "Test of Faith", "Instant", pr8TestOfFaith, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	pr8Unpreventable(t, g, src, bear, 2)
	if pr6Marked(g, bear) != 2 || pr8Counters(g, bear, game.CounterPlusOne) != 0 {
		t.Fatalf("unpreventable: damage %d (want 2), counters %d (want 0)", pr6Marked(g, bear), pr8Counters(g, bear, game.CounterPlusOne))
	}
	pr6Damage(t, g, src, bear, 1)
	if pr6Marked(g, bear) != 2 || pr8Counters(g, bear, game.CounterPlusOne) != 1 {
		t.Fatalf("then 1 more: damage %d (want 2), counters %d (want 1)", pr6Marked(g, bear), pr8Counters(g, bear, game.CounterPlusOne))
	}
}
