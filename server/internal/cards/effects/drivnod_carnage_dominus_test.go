package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const drivnodOracle = "51780f71-bf60-4208-94ea-76fa84790fb6"

// TestDrivnodDoublesADeathTriggeredAbility — Blood Artist's dies
// trigger fires twice for one creature dying while Drivnod is out.
func TestDrivnodDoublesADeathTriggeredAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Drivnod, Carnage Dominus", "Legendary Creature — Phyrexian Horror", drivnodOracle, false)
	pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)
	victim := pushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	meBefore, oppBefore := me.Life, opp.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	for i := 0; i < 6 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	pickPlayer(t, g, me.ID, opp.ID)
	// A second Blood Artist trigger, doubled by Drivnod, needs its
	// own target too.
	for i := 0; i < 6 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		pickPlayer(t, g, me.ID, opp.ID)
	}
	passPriorityAroundTable(t, g)

	if got := opp.Life; oppBefore-got != 2 {
		t.Errorf("opponent life lost %d, want 2 (Blood Artist doubled by Drivnod)", oppBefore-got)
	}
	if got := me.Life; got-meBefore != 2 {
		t.Errorf("controller life gained %d, want 2", got-meBefore)
	}
}

// TestDrivnodIndestructibleCounterAbility — exile three creature
// cards from the graveyard plus two Phyrexian mana symbols paid in
// life puts an indestructible counter on Drivnod.
func TestDrivnodIndestructibleCounterAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	drivnod := pushCatalogPermanent(g, me.ID, "Drivnod, Carnage Dominus", "Legendary Creature — Phyrexian Horror", drivnodOracle, false)
	a := pushGraveyardCardTyped(me, "Dead One", "Creature — Zombie")
	b := pushGraveyardCardTyped(me, "Dead Two", "Creature — Zombie")
	c := pushGraveyardCardTyped(me, "Dead Three", "Creature — Zombie")
	notACreature := pushGraveyardCardTyped(me, "Old Land", "Land")

	if err := g.ActivateCatalogAbility(me.ID, drivnod, 0, game.ActivateAbilityParams{
		ExileIDs:      []uuid.UUID{a, b, notACreature},
		PhyrexianLife: 2,
	}); err == nil {
		t.Fatal("a land among the exiled cards should be refused for \"three creature cards\"")
	}

	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, drivnod, 0, game.ActivateAbilityParams{
		ExileIDs:      []uuid.UUID{a, b, c},
		PhyrexianLife: 2,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := life - me.Life; got != 2*game.PhyrexianLifePerSymbol {
		t.Errorf("life paid = %d, want %d for two Phyrexian symbols", got, 2*game.PhyrexianLifePerSymbol)
	}
	if counterCount(g, drivnod, "indestructible") != 1 {
		t.Error("Drivnod should have an indestructible counter")
	}
	for _, id := range []uuid.UUID{a, b, c} {
		if !g.Exile.Contains(id) {
			t.Errorf("card %s should be exiled", id)
		}
	}
}
