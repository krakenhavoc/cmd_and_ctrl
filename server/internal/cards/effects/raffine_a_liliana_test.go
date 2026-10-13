package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// raffine_a_liliana_test.go — Liliana, Death's Majesty (#3212, slice raffine-a).

const lilianaDeathsMajestyOracle = "7fca595e-1902-40fd-a359-622caa612b6a"

func lilianaDMEffective(t *testing.T, g *game.Game, id uuid.UUID) game.Characteristic {
	t.Helper()
	var out game.Characteristic
	found := false
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				out = g.Battlefield.Cards[i].Effective()
				found = true
			}
		}
	})
	if !found {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return out
}

func TestLilianaDeathsMajestyIsRegisteredAndWhole(t *testing.T) {
	spec, ok := Lookup(lilianaDeathsMajestyOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 || len(spec.Activated) != 3 {
		t.Errorf("Completeness = %v, caveats %v, abilities %d", spec.Completeness, spec.Caveats, len(spec.Activated))
	}
}

func TestLilianaDeathsMajestyPlusOneMakesAZombieAndMills(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	lili := pushWalkerForTest(g, me.ID, "Liliana, Death's Majesty", lilianaDeathsMajestyOracle, 5)
	advanceToMainOf(t, g, seat)
	lib, gy := me.Library.Size(), me.Graveyard.Size()

	b16Activate(t, g, me.ID, lili, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	zombies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsCreature() && c.HasSubtype("Zombie") {
			zombies++
		}
	}
	if zombies != 1 {
		t.Errorf("Zombie creatures = %d, want 1", zombies)
	}
	if me.Library.Size() != lib-2 || me.Graveyard.Size() != gy+2 {
		t.Errorf("library %d→%d graveyard %d→%d, want two milled", lib, me.Library.Size(), gy, me.Graveyard.Size())
	}
	if got := counterCount(g, lili, game.CounterLoyalty); got != 6 {
		t.Errorf("loyalty = %d, want 6", got)
	}
}

// The −3 returns the card as a black Zombie in addition to what it was, and
// the addition is gone when it leaves and comes back.
func TestLilianaDeathsMajestyMinusThreeAddsBlackAndZombie(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	lili := pushWalkerForTest(g, me.ID, "Liliana, Death's Majesty", lilianaDeathsMajestyOracle, 5)
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	theirs := uuid.New()
	them.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Bears", TypeLine: "Creature — Bear",
		Owner: them.ID, Controller: them.ID})
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	advanceToMainOf(t, g, seat)

	// Only a creature card in MY graveyard is a legal target.
	if err := g.ActivateCatalogAbility(me.ID, lili, 1, game.ActivateAbilityParams{Targets: cardTarget(theirs)}); err == nil {
		t.Error("an opponent's graveyard card was accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, lili, 1, game.ActivateAbilityParams{Targets: cardTarget(land)}); err == nil {
		t.Error("a land card was accepted")
	}
	b16Activate(t, g, me.ID, lili, 1, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(bear) {
		t.Fatal("the Bear did not return")
	}
	ch := lilianaDMEffective(t, g, bear)
	if !typeListHasForTest(ch.Subtypes, "Zombie") || !typeListHasForTest(ch.Subtypes, "Bear") {
		t.Errorf("subtypes = %v, want Bear and Zombie", ch.Subtypes)
	}
	if !typeListHasForTest(ch.Colors, "B") || !typeListHasForTest(ch.Colors, "G") {
		t.Errorf("colors = %v, want green kept and black added", ch.Colors)
	}
	if got := counterCount(g, lili, game.CounterLoyalty); got != 2 {
		t.Errorf("loyalty = %d, want 2", got)
	}
}

// The −7 destroys every non-Zombie creature, yours included, and spares Zombies.
func TestLilianaDeathsMajestyMinusSevenSparesZombies(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	lili := pushWalkerForTest(g, me.ID, "Liliana, Death's Majesty", lilianaDeathsMajestyOracle, 7)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, them.ID, "Theirs", "Creature — Bear", 2, 2)
	zombie := b12Creature(g, them.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	advanceToMainOf(t, g, seat)

	b16Activate(t, g, me.ID, lili, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a non-Zombie creature survived the −7")
	}
	if !g.Battlefield.Contains(zombie) {
		t.Error("a Zombie was destroyed")
	}
}
