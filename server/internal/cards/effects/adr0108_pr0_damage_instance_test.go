package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr0_damage_instance_test.go — ADR 0108 PR 0: the ADR 0107 §6
// shield cards re-tested on the damage instance.

// ADR 0107 §6's known limit, on a printed card. Reverse Damage: "The next
// time a source of your choice would deal damage to you this turn,
// prevent that damage. You gain life equal to the damage prevented this
// way." A source that deals 4 damage to you and then, in a second
// instruction of the same resolution, 2 more is two instances (CR 615.8,
// 608.2c): the 4 is prevented and gained, the 2 is dealt. Before the
// instance both were one event batch, both were prevented and 6 was
// gained.
func TestPR0ReverseDamageMeetsOnlyTheFirstInstance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pridemate := pr7Pridemate(g, me.ID)
	src := pr7Creature(g, opp.ID, "Src", 6, "R")
	castCatalogSpell(t, g, "Reverse Damage", "Instant", pr7ReverseDamage, nil)
	passPriorityAroundTable(t, g)
	pr7Choose(t, g, me.ID, src)
	life := me.Life
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 4)
		_ = g.DealDamageToPlayerForEffect(src, me.ID, 2)
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if me.Life != life+4-2 {
		t.Errorf("life %d → %d, want +4 gained for the first instance and 2 dealt by the second", life, me.Life)
	}
	if got := pr7Counters(g, pridemate); got != 1 {
		t.Errorf("Pridemate has %d counters, want 1", got)
	}
}

// One printed instruction the catalog deals as a loop is ONE instance. A
// black source's "deals 2 damage to each creature" (damageEachMatching,
// Pyroclasm's body) and "deals 2 damage to each creature and each player"
// (Pestilence's body, b23DamageEachCreatureAndEachPlayer) against
// Shadowbane ("you and/or creatures you control"): all of it is
// prevented, and the life is gained once, with the total (CR 615.5,
// 615.8).
func TestPR0ShadowbaneCoversAWholeEachInstruction(t *testing.T) {
	for _, tc := range []struct {
		name   string
		deal   func(ctx *Context) error
		gained int
	}{
		{"each creature", func(ctx *Context) error { return damageEachMatching(ctx, Creature(), 2) }, 4},
		{"each creature and each player", func(ctx *Context) error { return b23DamageEachCreatureAndEachPlayer(ctx, 2) }, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			a := pr7Creature(g, me.ID, "Mine A", 1, "W")
			b := pr7Creature(g, me.ID, "Mine B", 1, "W")
			src := pr7Creature(g, opp.ID, "Src", 3, "B")
			castCatalogSpell(t, g, "Shadowbane", "Instant", pr7Shadowbane, nil)
			passPriorityAroundTable(t, g)
			pr7Choose(t, g, me.ID, src)
			life, theirs := me.Life, opp.Life
			g.WithWriteLock(func() {
				item := &game.StackItem{SourceCardID: src, Controller: opp.ID, Owner: opp.ID}
				if err := tc.deal(NewContext(g, item)); err != nil {
					t.Fatal(err)
				}
			})
			g.RunStateChecksForTest()
			if damageMarkedOn(g, a) != 0 || damageMarkedOn(g, b) != 0 {
				t.Fatalf("creature damage %d and %d, want 0: one instance, all of it prevented",
					damageMarkedOn(g, a), damageMarkedOn(g, b))
			}
			if me.Life != life+tc.gained {
				t.Errorf("life %d → %d, want +%d gained once for the whole instance", life, me.Life, tc.gained)
			}
			if damageMarkedOn(g, src) == 0 {
				t.Errorf("the source took no damage: only you and your creatures are protected")
			}
			if tc.name == "each creature and each player" && opp.Life != theirs-2 {
				t.Errorf("opponent %d → %d: their damage is not protected", theirs, opp.Life)
			}
		})
	}
}
