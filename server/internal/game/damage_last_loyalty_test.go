package game

import (
	"testing"

	"github.com/google/uuid"
)

// damage_last_loyalty_test.go — #5646: the gate that keeps a
// planeswalker's last loyalty counter when damage would remove them all
// (Deification), on a stubbed catalog static.

const lastLoyaltyOracle = "test-damage-last-loyalty"

var deificationStatic = DamageLastLoyaltyStatic{Label: "Deification", OfChosenType: true, WhileYouControlACreature: true}

// withDamageLastLoyalty stubs the catalog so a battlefield permanent with
// lastLoyaltyOracle has exactly these statics.
func withDamageLastLoyalty(t *testing.T, statics ...DamageLastLoyaltyStatic) {
	t.Helper()
	prev := CatalogDamageLastLoyalty
	CatalogDamageLastLoyalty = func(key string) []DamageLastLoyaltyStatic {
		if key == lastLoyaltyOracle {
			return statics
		}
		return nil
	}
	t.Cleanup(func() { CatalogDamageLastLoyalty = prev })
}

// pushLastLoyaltyEnchantment puts an enchantment carrying the stubbed
// statics onto the battlefield under `owner`, with `chosen` as the
// planeswalker type chosen as it entered.
func pushLastLoyaltyEnchantment(g *Game, owner *Player, chosen string) uuid.UUID {
	c := NewCard("Last Loyalty", owner.ID)
	c.TypeLine = "Enchantment"
	c.OracleID = lastLoyaltyOracle
	c.ChosenOption = chosen
	g.WithWriteLock(func() { g.Battlefield.PushTop(c) })
	return c.InstanceID
}

// pushDeifiedWalker puts a planeswalker with the given type line and loyalty
// onto the battlefield under `owner`.
func pushDeifiedWalker(g *Game, owner *Player, typeLine string, loyalty int) uuid.UUID {
	c := NewCard("Walker "+typeLine, owner.ID)
	c.TypeLine = typeLine
	c.Counters = map[string]int{CounterLoyalty: loyalty}
	g.WithWriteLock(func() { g.Battlefield.PushTop(c) })
	return c.InstanceID
}

func loyaltyOn(g *Game, id uuid.UUID) int {
	c, ok := battlefieldCardByID(g, id)
	if !ok {
		return -1
	}
	return c.Counters[CounterLoyalty]
}

// Damage that would remove every loyalty counter removes all but one;
// damage that would not is unchanged; a walker at one counter keeps it.
func TestDamageLeavesTheLastLoyaltyCounter(t *testing.T) {
	withDamageLastLoyalty(t, deificationStatic)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	pushColouredCreature(g, me, "My Creature", []string{"W"})
	pushLastLoyaltyEnchantment(g, me, "Ajani")
	walker := pushDeifiedWalker(g, me, "Legendary Planeswalker — Ajani", 5)

	dealToCreature(t, g, src, walker, 2)
	if got := loyaltyOn(g, walker); got != 3 {
		t.Fatalf("after 2 damage loyalty = %d, want 3: the damage would not remove them all", got)
	}
	dealToCreature(t, g, src, walker, 3)
	if got := loyaltyOn(g, walker); got != 1 {
		t.Fatalf("after exactly-lethal damage loyalty = %d, want 1", got)
	}
	dealToCreature(t, g, src, walker, 7)
	if got := loyaltyOn(g, walker); got != 1 {
		t.Fatalf("a walker at 1 loyalty dealt 7 has %d, want 1: all but one of one counter is none", got)
	}
	g.RunStateChecksForTest()
	if _, ok := battlefieldCardByID(g, walker); !ok {
		t.Fatal("the walker left the battlefield")
	}
	// The damage was still dealt in full: the event carries the amount.
	var dealt int
	for _, ev := range iwtEventsOfKind(g, EventDealDamage) {
		if ev.Target == walker {
			dealt += ev.Amount
		}
	}
	if dealt != 12 {
		t.Errorf("damage events to the walker total %d, want 12: the damage is not prevented", dealt)
	}
}

// The gate's conditions, each read as the damage lands.
func TestDamageLastLoyaltyConditions(t *testing.T) {
	for _, tc := range []struct {
		name     string
		chosen   string
		typeLine string
		creature bool
		mine     bool
		want     int
	}{
		{"covered", "Ajani", "Legendary Planeswalker — Ajani", true, true, 1},
		{"no creature", "Ajani", "Legendary Planeswalker — Ajani", false, true, 0},
		{"another type", "Jace", "Legendary Planeswalker — Ajani", true, true, 0},
		{"choice unanswered", "", "Legendary Planeswalker — Ajani", true, true, 0},
		{"an opponent's walker", "Ajani", "Legendary Planeswalker — Ajani", true, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withDamageLastLoyalty(t, deificationStatic)
			g := newActiveGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			src := pushColouredCreature(g, opp, "Source", []string{"R"})
			if tc.creature {
				pushColouredCreature(g, me, "My Creature", []string{"W"})
			}
			pushLastLoyaltyEnchantment(g, me, tc.chosen)
			owner := me
			if !tc.mine {
				owner = opp
			}
			walker := pushDeifiedWalker(g, owner, tc.typeLine, 3)
			dealToCreature(t, g, src, walker, 4)
			if got := loyaltyOn(g, walker); got != tc.want {
				t.Errorf("loyalty = %d, want %d", got, tc.want)
			}
		})
	}
}

// The static is read off the battlefield: once its permanent has left,
// damage takes the last counter. And only damage is capped: a counter
// removal that is not damage takes the last counter while it is there.
func TestDamageLastLoyaltyIsAStaticAndOnlyAboutDamage(t *testing.T) {
	withDamageLastLoyalty(t, deificationStatic)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	pushColouredCreature(g, me, "My Creature", []string{"W"})
	ench := pushLastLoyaltyEnchantment(g, me, "Ajani")
	a := pushDeifiedWalker(g, me, "Legendary Planeswalker — Ajani", 2)
	b := pushDeifiedWalker(g, me, "Legendary Planeswalker — Ajani", 2)

	g.WithWriteLock(func() { _ = g.AddCounterForEffect(a, CounterLoyalty, -2) })
	if got := loyaltyOn(g, a); got != 0 {
		t.Errorf("removing 2 loyalty counters with an effect left %d, want 0: that is not damage", got)
	}
	g.WithWriteLock(func() { _, _ = g.Battlefield.Remove(ench) })
	dealToCreature(t, g, src, b, 2)
	if got := loyaltyOn(g, b); got != 0 {
		t.Errorf("loyalty = %d after the static's permanent left, want 0", got)
	}
}

// A planeswalker that is also a creature has the whole damage marked
// (CR 120.3e) while the loyalty removal is capped, and it is itself the
// creature its controller controls.
func TestDamageLastLoyaltyOnACreaturePlaneswalker(t *testing.T) {
	withDamageLastLoyalty(t, deificationStatic)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	pushLastLoyaltyEnchantment(g, me, "Gideon")
	walker := pushDeifiedWalker(g, me, "Legendary Planeswalker Creature — Gideon Soldier", 3)
	g.WithWriteLock(func() { findBattlefieldCard(g, walker).Toughness = 20 })

	dealToCreature(t, g, src, walker, 5)
	if got := loyaltyOn(g, walker); got != 1 {
		t.Errorf("loyalty = %d, want 1", got)
	}
	if got := damageOn(g, walker); got != 5 {
		t.Errorf("damage marked = %d, want all 5", got)
	}
}

// The chosen type is on the snapshot, so the cap holds in a restored
// game.
func TestDamageLastLoyaltySurvivesARestore(t *testing.T) {
	withDamageLastLoyalty(t, deificationStatic)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushColouredCreature(g, opp, "Source", []string{"R"})
	pushColouredCreature(g, me, "My Creature", []string{"W"})
	pushLastLoyaltyEnchantment(g, me, "Ajani")
	walker := pushDeifiedWalker(g, me, "Legendary Planeswalker — Ajani", 4)

	r := strictRestore(t, g)
	dealToCreature(t, r, src, walker, 9)
	if got := loyaltyOn(r, walker); got != 1 {
		t.Errorf("loyalty in the restored game = %d, want 1", got)
	}
}
