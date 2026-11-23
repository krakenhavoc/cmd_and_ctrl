package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const cityscapeLevelerOracle = "d4d65797-2b92-4265-9169-133120c86c7f"

// The cast-trigger leg: fires while the Leveler is still a spell on
// the stack, above it, and its target is chosen by the CASTER.
func TestCityscapeLevelerCastTriggerDestroysAndMakesAPowerstone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", ManaCost: "{2}",
		Owner: opp.ID, Controller: opp.ID,
	})

	leveler := effectstest.CastCatalogSpell(t, g, "Cityscape Leveler", "Artifact Creature — Construct", cityscapeLevelerOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || p.PickTargetMin != 0 {
		t.Fatalf("the cast trigger asks for up to one nonland permanent: %+v", p)
	}
	if !effectstest.HasID(p.PickTargetCards, rock) {
		t.Fatal("the opponent's artifact should be a legal pick")
	}
	effectstest.PickCard(t, g, me.ID, rock)
	if effectstest.TriggerOnStack(g, leveler) == nil {
		t.Fatal("the destroy trigger should sit above the Leveler on the stack")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the targeted permanent should be destroyed")
	}
	if got := effectstest.B43TokensNamed(g, opp.ID, "Powerstone"); got != 1 {
		t.Errorf("the DESTROYED PERMANENT'S CONTROLLER should get a Powerstone, got %d", got)
	}
	if !g.Battlefield.Contains(leveler) {
		t.Error("the Leveler resolves after its cast trigger")
	}
}

// The attack-trigger leg: the same printed ability, watched from the
// battlefield once the Leveler is a permanent.
func TestCityscapeLevelerAttackTriggerDestroysAndMakesAPowerstone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leveler := effectstest.PushCatalogPermanent(g, me.ID, "Cityscape Leveler", "Artifact Creature — Construct", cityscapeLevelerOracle, false)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", ManaCost: "{2}",
		Owner: opp.ID, Controller: opp.ID,
	})

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(leveler, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the targeted permanent should be destroyed")
	}
	if got := effectstest.B43TokensNamed(g, opp.ID, "Powerstone"); got != 1 {
		t.Errorf("the destroyed permanent's controller should get a Powerstone, got %d", got)
	}
}

// CR 603.3d: with nothing else on the board when the Leveler is cast,
// "up to one" asks nothing and the spell just resolves.
func TestCityscapeLevelerCastWithNoOtherPermanentsAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]

	leveler := effectstest.CastCatalogSpell(t, g, "Cityscape Leveler", "Artifact Creature — Construct", cityscapeLevelerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(leveler) {
		t.Fatal("the Leveler should resolve normally")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("no prompt should be queued with nothing to destroy")
	}
}

// Unearth returns the Leveler with haste; the printed trample keyword
// travels with it either way.
func TestCityscapeLevelerUnearth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.SeedGraveyardCard(t, g, "Cityscape Leveler", "Artifact Creature — Construct", cityscapeLevelerOracle)

	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Unearth: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	c, ok := g.LookupCardForEffect(id)
	if !ok || !g.Battlefield.Contains(id) {
		t.Fatalf("the unearthed Leveler should be on the battlefield")
	}
	eff := c.Effective()
	if !effectstest.HasKeywordForTest(eff.Abilities, "haste") {
		t.Errorf("the unearthed Leveler should have haste: %v", eff.Abilities)
	}
	if !effectstest.HasKeywordForTest(eff.Abilities, "trample") {
		t.Errorf("the Leveler should keep trample: %v", eff.Abilities)
	}
}
