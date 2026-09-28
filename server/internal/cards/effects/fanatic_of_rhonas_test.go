package effects

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const fanaticOfRhonasOracle = "7973820b-fdaf-46ec-9e3e-d4c0e77b5067"

func TestFanaticOfRhonasPlainManaAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fanatic := seedPermanentWithOracle(g, me.ID, "Fanatic of Rhonas", "Creature — Snake Druid", fanaticOfRhonasOracle)

	if err := g.ActivateManaAbility(me.ID, fanatic, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
}

// Ferocious is gated: no creature with power 4+ means the second
// ability fails with the cost unpaid, not a partial {G}{G}{G}{G}.
func TestFanaticOfRhonasFerociousNeedsABigCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fanatic := seedPermanentWithOracle(g, me.ID, "Fanatic of Rhonas", "Creature — Snake Druid", fanaticOfRhonasOracle)

	if err := g.ActivateManaAbility(me.ID, fanatic, 1, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("err = %v, want ErrConditionNotMet", err)
	}
	if len(me.ManaPool) != 0 {
		t.Fatalf("pool = %v, want empty — a failed gate pays nothing", me.ManaPool)
	}
	if c, ok := g.LookupCardForEffect(fanatic); !ok || c.Tapped {
		t.Fatalf("a failed gate must not tap the source")
	}

	pushVanillaCreature(g, me.ID, "Big Bear", 4, 4)
	if err := g.ActivateManaAbility(me.ID, fanatic, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility with a power-4 creature out: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 4 {
		t.Errorf("pool %v, want four green mana", got)
	}
}

// Eternalize makes a 4/4 black Zombie Snake Druid with no mana cost —
// the printed subtypes are kept "in addition to its other types", and
// the token is exiled from the graveyard to make it.
func TestFanaticOfRhonasEternalize(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := seedGraveyardCard(t, g, "Fanatic of Rhonas", "Creature — Snake Druid", fanaticOfRhonasOracle)

	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Eternalize: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Errorf("the eternalized card should be exiled")
	}
	var token *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == "Fanatic of Rhonas" && c.Controller == me.ID && c.InstanceID != id {
			token = c
		}
	}
	if token == nil {
		t.Fatalf("no eternalized token on the battlefield")
	}
	if token.Power != 4 || token.Toughness != 4 {
		t.Errorf("token P/T = %d/%d, want 4/4", token.Power, token.Toughness)
	}
	if token.ManaCost != "" {
		t.Errorf("token mana cost = %q, want empty", token.ManaCost)
	}
	if len(token.Colors) != 1 || token.Colors[0] != "B" {
		t.Errorf("token colors = %v, want [B]", token.Colors)
	}
}
