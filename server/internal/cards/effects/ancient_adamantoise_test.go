package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const adamantoiseOracle = "e6873252-653a-47a8-99b7-b2ef70aa1f7f"

// adamantoiseWith seats an Adamantoise (toughness 3 so a test can kill
// it) and returns it with an opposing damage source.
func adamantoiseWith(g *game.Game) (toad, src uuid.UUID) {
	me, opp := g.Seats[0], g.Seats[1]
	toad = apaPush(g, me.ID, me.ID, apaCreature("Ancient Adamantoise", adamantoiseOracle, 8, 3))
	src = pr7Creature(g, opp.ID, "Pinger", 1, "R")
	return toad, src
}

func TestAdamantoiseDamageSurvivesCleanup(t *testing.T) {
	g := newCatalogGame(t)
	toad, src := adamantoiseWith(g)
	pr6Damage(t, g, src, toad, 2)
	endTurn(t, g)
	if got := pr6Marked(g, toad); got != 2 {
		t.Fatalf("marked damage after cleanup = %d, want 2", got)
	}
}

func TestAdamantoiseLethalDamageAccumulatesOverTurns(t *testing.T) {
	g := newCatalogGame(t)
	toad, src := adamantoiseWith(g)
	pr6Damage(t, g, src, toad, 2)
	endTurn(t, g)
	endTurn(t, g)
	pr6Damage(t, g, src, toad, 1)
	if got := pr6Marked(g, toad); got != -1 {
		t.Fatalf("Adamantoise with 3 damage on toughness 3 should have died; marked = %d", got)
	}
}

func TestAdamantoiseThatLostAbilitiesIsCleaned(t *testing.T) {
	g := newCatalogGame(t)
	toad, src := adamantoiseWith(g)
	pr6Damage(t, g, src, toad, 2)
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(toad),
			[]game.Mod{game.LoseAllAbilitiesMod()}, game.IndefiniteDuration(), "test — loses all abilities")
	})
	endTurn(t, g)
	if got := pr6Marked(g, toad); got != 0 {
		t.Fatalf("marked damage after cleanup = %d, want 0 (no abilities, no exemption)", got)
	}
}

func TestPhasedOutAdamantoiseIsCleaned(t *testing.T) {
	g := newCatalogGame(t)
	toad, src := adamantoiseWith(g)
	pr6Damage(t, g, src, toad, 2)
	g.WithWriteLock(func() {
		if err := g.PhaseOutForEffect(uuid.Nil, toad); err != nil {
			t.Fatal(err)
		}
	})
	endTurn(t, g)
	for _, c := range g.PhasedOut.Cards {
		if c.InstanceID == toad {
			if c.DamageMarked != 0 {
				t.Fatalf("phased-out Adamantoise kept %d damage, want 0", c.DamageMarked)
			}
			return
		}
	}
	t.Fatal("Adamantoise is not in the phased-out zone")
}

func TestRegenerationStillClearsAdamantoiseDamage(t *testing.T) {
	g := newCatalogGame(t)
	toad, src := adamantoiseWith(g)
	pr6Damage(t, g, src, toad, 2)
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(toad); err != nil {
			t.Fatal(err)
		}
	})
	pr6Damage(t, g, src, toad, 1) // lethal: the shield regenerates it
	if got := pr6Marked(g, toad); got != 0 {
		t.Fatalf("marked damage after regeneration = %d, want 0", got)
	}
}

func TestAdamantoiseDiesExilesItAndMakesTenTappedTreasures(t *testing.T) {
	g := newCatalogGame(t)
	toad, src := adamantoiseWith(g)
	pr6Damage(t, g, src, toad, 3)
	passPriorityAroundTable(t, g)
	exiled := false
	for _, c := range g.Exile.Cards {
		if c.InstanceID == toad {
			exiled = true
		}
	}
	if !exiled {
		t.Fatal("Adamantoise was not exiled")
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && c.Controller == g.Seats[0].ID {
			if !c.Tapped {
				t.Fatal("Treasure entered untapped")
			}
			n++
		}
	}
	if n != 10 {
		t.Fatalf("Treasures = %d, want 10", n)
	}
}
