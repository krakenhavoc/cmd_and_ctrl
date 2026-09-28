package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const armyOfTheDamnedOracle = "75d667ec-86f4-4850-a3b6-e7a9fc7053b0"

func TestArmyOfTheDamnedCreatesThirteenTappedZombies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	castCatalogSpell(t, g, "Army of the Damned", "Sorcery", armyOfTheDamnedOracle, nil)
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := seedGraveyardCard(t, g, "Army of the Damned", "Sorcery", armyOfTheDamnedOracle)

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := b43TokensNamed(g, me.ID, "Zombie"); got != 13 {
		t.Fatalf("Zombie tokens = %d, want 13", got)
	}
	if !g.Exile.Contains(id) {
		t.Errorf("the flashed-back sorcery should be exiled")
	}
}
