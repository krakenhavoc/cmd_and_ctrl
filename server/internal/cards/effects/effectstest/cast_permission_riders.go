package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cast_permission_riders_test.go — what a cast permission puts on the
// permanent its cast produces (#3768): Edgar, Master Machinist's "that
// artifact enters tapped" and Strago and Relm's "it gains haste and 'At
// the beginning of the end step, sacrifice this creature.'" (deck request
// #3655).

const (
	EdgarMasterMachinistOracle = "75348d55-99cd-4dab-97be-1f818c43a69d"
	stragoAndRelmOracle        = "316e7c28-2b7a-40ae-961b-909644109fe4"
)

// stragoGame is seat 0 in its precombat main with an untapped Strago and
// Relm and {2}{R} floating; seat 1's library has a land on top of a
// creature card. Returns the game and the two library cards.
func StragoGame(t *testing.T) (g *game.Game, me, opp *game.Player, strago, land, bear uuid.UUID) {
	t.Helper()
	g = NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	strago = PushCatalogPermanent(g, me.ID, "Strago and Relm", "Legendary Creature — Human Wizard", stragoAndRelmOracle, false)
	bear, land = uuid.New(), uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: bear, Name: "Stolen Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	opp.Library.PushTop(game.Card{InstanceID: land, Name: "Some Land", TypeLine: "Land", Owner: opp.ID, Controller: opp.ID})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "R"})
	if err := g.ActivateCatalogAbility(me.ID, strago, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("Sketch and Lore: %v", err)
	}
	PassPriorityAroundTable(t, g)
	return g, me, opp, strago, land, bear
}
