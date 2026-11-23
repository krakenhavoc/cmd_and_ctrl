package testsa

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestArmyOfTheDamnedCreatesThirteenTappedZombies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]

	effectstest.CastCatalogSpell(t, g, "Army of the Damned", "Sorcery", effectstest.ArmyOfTheDamnedOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	n, tapped := 0, 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.Name == "Zombie" {
			n++
			if c.Tapped {
				tapped++
			}
		}
	}
	if n != 13 {
		t.Fatalf("Zombie tokens = %d, want 13", n)
	}
	if tapped != 13 {
		t.Errorf("tapped Zombie tokens = %d, want 13", tapped)
	}
}

func TestArmyOfTheDamnedFlashback(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.SeedGraveyardCard(t, g, "Army of the Damned", "Sorcery", effectstest.ArmyOfTheDamnedOracle)

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.B43TokensNamed(g, me.ID, "Zombie"); got != 13 {
		t.Fatalf("Zombie tokens = %d, want 13", got)
	}
	if !g.Exile.Contains(id) {
		t.Errorf("the flashed-back sorcery should be exiled")
	}
}
