package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestDrivnodDoublesADeathTriggeredAbility — Blood Artist's dies
// trigger fires twice for one creature dying while Drivnod is out.
func TestDrivnodDoublesADeathTriggeredAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Drivnod, Carnage Dominus", "Legendary Creature — Phyrexian Horror", effectstest.DrivnodOracle, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", effectstest.BloodArtistOracle, false)
	victim := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	meBefore, oppBefore := me.Life, opp.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	for i := 0; i < 6 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	// A second Blood Artist trigger, doubled by Drivnod, needs its
	// own target too.
	for i := 0; i < 6 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drivnod := effectstest.PushCatalogPermanent(g, me.ID, "Drivnod, Carnage Dominus", "Legendary Creature — Phyrexian Horror", effectstest.DrivnodOracle, false)
	a := effectstest.PushGraveyardCardTyped(me, "Dead One", "Creature — Zombie")
	b := effectstest.PushGraveyardCardTyped(me, "Dead Two", "Creature — Zombie")
	c := effectstest.PushGraveyardCardTyped(me, "Dead Three", "Creature — Zombie")
	notACreature := effectstest.PushGraveyardCardTyped(me, "Old Land", "Land")

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
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - me.Life; got != 2*game.PhyrexianLifePerSymbol {
		t.Errorf("life paid = %d, want %d for two Phyrexian symbols", got, 2*game.PhyrexianLifePerSymbol)
	}
	if effectstest.CounterCount(g, drivnod, "indestructible") != 1 {
		t.Error("Drivnod should have an indestructible counter")
	}
	for _, id := range []uuid.UUID{a, b, c} {
		if !g.Exile.Contains(id) {
			t.Errorf("card %s should be exiled", id)
		}
	}
}
