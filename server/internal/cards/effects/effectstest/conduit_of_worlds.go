package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conduit_of_worlds_test.go — #2173: a cast permission with an "if you
// do" follow-up. Conduit of Worlds.

const conduitOfWorldsOracle = "ed14be15-8f8d-4fe3-a147-f5da8ed873bf"

// conduitTable sets up Conduit on seat 0's battlefield at main phase,
// a nonland permanent card and a land card in their graveyard, and a
// {R} instant in hand.
func ConduitTable(t *testing.T) (g *game.Game, me *game.Player, conduit, bear, land, bolt uuid.UUID) {
	t.Helper()
	g = NewCatalogGame(t)
	me = g.Seats[0]
	conduit = PushCatalogPermanent(g, me.ID, "Conduit of Worlds", "Artifact", conduitOfWorldsOracle, false)
	bear = PushGraveyardCardWithTypeLineID(me, "Grizzly Bears", "Creature — Bear")
	land = PushGraveyardCardWithTypeLineID(me, "Forest", "Basic Land — Forest")
	bolt = uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	// The bear needs a cost to be castable at all.
	for i := range me.Graveyard.Cards {
		if me.Graveyard.Cards[i].InstanceID == bear {
			me.Graveyard.Cards[i].ManaCost = "{1}{G}"
			me.Graveyard.Cards[i].Power, me.Graveyard.Cards[i].Toughness = 2, 2
		}
	}
	AdvanceTo(t, g, game.StepPrecombatMain)
	return
}

func ConduitActivate(t *testing.T, g *game.Game, me *game.Player, conduit, target uuid.UUID) error {
	t.Helper()
	err := g.ActivateCatalogAbility(me.ID, conduit, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	if err == nil {
		PassPriorityAroundTable(t, g)
	}
	return err
}

func ConduitMana(t *testing.T, g *game.Game, me *game.Player, cost string) {
	t.Helper()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, cost); err != nil {
		t.Fatal(err)
	}
}
