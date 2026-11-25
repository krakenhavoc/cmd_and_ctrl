package game

import (
	"testing"

	"github.com/google/uuid"
)

// cant_get_counters_test.go — #6241 (CR 122.1, CR 101.2): Solemnity's
// "Players can't get counters" and Suncleanser's "That player can't get
// counters for as long as this creature remains on the battlefield".

const solemnityTestOracle = "test-6241-solemnity"

func playerCounter(g *Game, p *Player, name string) int {
	n := 0
	g.ReadSnapshot(func() { n = p.Counters[name] })
	return n
}

func giveCounter(t *testing.T, g *Game, p *Player, name string, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddPlayerCounterForEffect(p.ID, name, n); err != nil {
			t.Fatalf("add %s: %v", name, err)
		}
	})
}

// Solemnity's acceptance: with it out, an energy gain, a poison counter
// and a proliferate give no player a counter; counters already there
// stay, and once it leaves players get counters again.
func TestPlayersCantGetCountersWorldRule(t *testing.T) {
	withStaticAbilities(t, func(key string) []StaticAbility {
		if key != solemnityTestOracle {
			return nil
		}
		return []StaticAbility{{Layer: Layer6Ability,
			AppliesTo: func(target *Card, _ *Game, source *Card) bool { return target.InstanceID == source.InstanceID },
			Apply:     func(c *Characteristic, _ *Card, _ *Game, _ *Card) { c.ExtraRestrictions |= PlayersCantGetCounters }}}
	})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giveCounter(t, g, opp, CounterPoison, 2)
	sol := pushTypedTestCard(g, Card{Name: "Solemnity", TypeLine: "Enchantment", OracleID: solemnityTestOracle, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })

	giveCounter(t, g, me, CounterEnergy, 3)
	giveCounter(t, g, opp, CounterPoison, 1)
	giveCounter(t, g, me, CounterExperience, 1)
	if n := playerCounter(g, me, CounterEnergy); n != 0 {
		t.Errorf("energy under Solemnity = %d, want 0", n)
	}
	if n := playerCounter(g, me, CounterExperience); n != 0 {
		t.Errorf("experience under Solemnity = %d, want 0", n)
	}
	if n := playerCounter(g, opp, CounterPoison); n != 2 {
		t.Errorf("poison under Solemnity = %d, want the 2 already there", n)
	}
	var players []uuid.UUID
	g.ReadSnapshot(func() { _, players = g.ProliferateCandidatesForEffect() })
	if len(players) != 0 {
		t.Errorf("proliferate offers players %v under Solemnity", players)
	}

	g.WithWriteLock(func() {
		if _, err := g.Battlefield.Remove(sol); err != nil {
			t.Fatal(err)
		}
		g.layerVersion.Add(1)
		g.RecomputeLayersIfStaleLocked()
	})
	giveCounter(t, g, me, CounterEnergy, 1)
	if n := playerCounter(g, me, CounterEnergy); n != 1 {
		t.Errorf("energy after Solemnity left = %d, want 1", n)
	}
}

// Suncleanser: the named player can't get counters while the source is
// on the battlefield; another player can, and the bar ends when the
// source leaves.
func TestPlayerCantGetCountersWhileSourceRemains(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushTypedTestCard(g, Card{Name: "Suncleanser", TypeLine: "Creature — Human Cleric", Owner: me.ID, Controller: me.ID, Power: 1, Toughness: 4})
	g.WithWriteLock(func() {
		d, ok := g.ForAsLongAsOnBattlefieldDuration(src)
		if !ok {
			t.Fatal("no duration")
		}
		if !g.PlayerCantGetCountersForDurationForEffect(src, opp.ID, d, "Suncleanser") {
			t.Fatal("not registered")
		}
	})
	giveCounter(t, g, opp, CounterEnergy, 2)
	giveCounter(t, g, me, CounterEnergy, 2)
	if n := playerCounter(g, opp, CounterEnergy); n != 0 {
		t.Errorf("the target got %d energy, want 0", n)
	}
	if n := playerCounter(g, me, CounterEnergy); n != 2 {
		t.Errorf("the other player got %d energy, want 2", n)
	}
	g.WithWriteLock(func() {
		if _, err := g.Battlefield.Remove(src); err != nil {
			t.Fatal(err)
		}
	})
	giveCounter(t, g, opp, CounterEnergy, 1)
	if n := playerCounter(g, opp, CounterEnergy); n != 1 {
		t.Errorf("after the source left the target got %d energy, want 1", n)
	}
}
