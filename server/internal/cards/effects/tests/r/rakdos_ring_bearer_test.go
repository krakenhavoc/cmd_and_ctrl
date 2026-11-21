package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rakdos_ring_bearer_test.go — #2878. A Nazgûl is not legendary, but
// the Ring makes its controller's Ring-bearer legendary (CR 701.54c)
// and Nazgûl tempts the Ring when it enters. "Whenever a legendary
// creature you control dies" is therefore right to fire for the
// Ring-bearer Nazgûl and wrong for any other Nazgûl.

// rakdosFiredOnDeathOf kills `id` and reports whether Rakdos Joins Up
// put its opponent-picker on the table.
func rakdosFiredOnDeathOf(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) bool {
	t.Helper()
	effectstest.B18Kill(t, g, id)
	return effectstest.LatestPickTarget(g, me.ID) != nil
}

func TestRakdosJoinsUpFiresForTheRingBearerNazgulOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Rakdos Joins Up", "Legendary Enchantment", effectstest.B33RakdosJoinsUpOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	first := effectstest.CastRingCreature(t, g, "Nazgûl", "Creature — Wraith Knight", effectstest.NazgulOracle, 1, 2)
	effectstest.RingSettle(t, g)
	if effectstest.RingBearerOf(g, me.ID) != first {
		t.Fatalf("setup: the first Nazgûl is not the Ring-bearer")
	}
	second := effectstest.CastRingCreature(t, g, "Nazgûl", "Creature — Wraith Knight", effectstest.NazgulOracle, 1, 2)
	effectstest.RingSettle(t, g)
	if effectstest.RingPrompt(g, me.ID) != nil {
		effectstest.AnswerRing(t, g, me.ID, second)
		effectstest.RingSettle(t, g)
	}
	if effectstest.RingBearerOf(g, me.ID) != second {
		t.Fatalf("setup: the second Nazgûl is not the Ring-bearer")
	}
	if effectstest.RingBFCard(t, g, first).IsLegendary() {
		t.Fatal("setup: the first Nazgûl is still legendary")
	}
	// The first one is no longer the bearer: not a legendary creature.
	if rakdosFiredOnDeathOf(t, g, me, first) {
		t.Error("Rakdos Joins Up fired for a Nazgûl that is not the Ring-bearer (#2878)")
	}
}

func TestRakdosJoinsUpFiresForTheRingBearerNazgul(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Rakdos Joins Up", "Legendary Enchantment", effectstest.B33RakdosJoinsUpOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	nazgul := effectstest.CastRingCreature(t, g, "Nazgûl", "Creature — Wraith Knight", effectstest.NazgulOracle, 1, 2)
	effectstest.RingSettle(t, g)
	if effectstest.RingBearerOf(g, me.ID) != nazgul {
		t.Fatal("setup: not the Ring-bearer")
	}
	if !rakdosFiredOnDeathOf(t, g, me, nazgul) {
		t.Error("the Ring-bearer is legendary (CR 701.54c); Rakdos Joins Up should fire")
	}
}

// The sacrifice exit (Fleshbag Marauder's, the report's table) stamps
// last-known supertypes the same way the destroy exit does: the bearer
// sacrificed was legendary, a Nazgûl sacrificed while a different
// creature is the bearer was not.
func TestRakdosJoinsUpSacrificeExitReadsTheRingBearerToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Rakdos Joins Up", "Legendary Enchantment", effectstest.B33RakdosJoinsUpOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	nazgul := effectstest.CastRingCreature(t, g, "Nazgûl", "Creature — Wraith Knight", effectstest.NazgulOracle, 1, 2)
	effectstest.RingSettle(t, g)
	effectstest.AnswerRing(t, g, me.ID, bear)
	effectstest.RingSettle(t, g)
	if effectstest.RingBearerOf(g, me.ID) != bear {
		t.Fatal("setup: the Bears are not the Ring-bearer")
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(nazgul) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("Rakdos Joins Up fired for a sacrificed Nazgûl that was not the Ring-bearer (#2878)")
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Error("Rakdos Joins Up did not fire for the sacrificed Ring-bearer, who was legendary")
	}
}
