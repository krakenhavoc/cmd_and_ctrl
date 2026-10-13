package effects

import (
	"testing"

	"github.com/google/uuid"

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
	b18Kill(t, g, id)
	return latestPickTarget(g, me.ID) != nil
}

func TestRakdosJoinsUpFiresForTheRingBearerNazgulOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Rakdos Joins Up", "Legendary Enchantment", b33RakdosJoinsUpOracle, nil)
	passPriorityAroundTable(t, g)

	first := castRingCreature(t, g, "Nazgûl", "Creature — Wraith Knight", nazgulOracle, 1, 2)
	ringSettle(t, g)
	if ringBearerOf(g, me.ID) != first {
		t.Fatalf("setup: the first Nazgûl is not the Ring-bearer")
	}
	second := castRingCreature(t, g, "Nazgûl", "Creature — Wraith Knight", nazgulOracle, 1, 2)
	ringSettle(t, g)
	if ringPrompt(g, me.ID) != nil {
		answerRing(t, g, me.ID, second)
		ringSettle(t, g)
	}
	if ringBearerOf(g, me.ID) != second {
		t.Fatalf("setup: the second Nazgûl is not the Ring-bearer")
	}
	if ringBFCard(t, g, first).IsLegendary() {
		t.Fatal("setup: the first Nazgûl is still legendary")
	}
	// The first one is no longer the bearer: not a legendary creature.
	if rakdosFiredOnDeathOf(t, g, me, first) {
		t.Error("Rakdos Joins Up fired for a Nazgûl that is not the Ring-bearer (#2878)")
	}
}

func TestRakdosJoinsUpFiresForTheRingBearerNazgul(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Rakdos Joins Up", "Legendary Enchantment", b33RakdosJoinsUpOracle, nil)
	passPriorityAroundTable(t, g)
	nazgul := castRingCreature(t, g, "Nazgûl", "Creature — Wraith Knight", nazgulOracle, 1, 2)
	ringSettle(t, g)
	if ringBearerOf(g, me.ID) != nazgul {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Rakdos Joins Up", "Legendary Enchantment", b33RakdosJoinsUpOracle, nil)
	passPriorityAroundTable(t, g)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	nazgul := castRingCreature(t, g, "Nazgûl", "Creature — Wraith Knight", nazgulOracle, 1, 2)
	ringSettle(t, g)
	answerRing(t, g, me.ID, bear)
	ringSettle(t, g)
	if ringBearerOf(g, me.ID) != bear {
		t.Fatal("setup: the Bears are not the Ring-bearer")
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(nazgul) })
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("Rakdos Joins Up fired for a sacrificed Nazgûl that was not the Ring-bearer (#2878)")
	}

	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) == nil {
		t.Error("Rakdos Joins Up did not fire for the sacrificed Ring-bearer, who was legendary")
	}
}
