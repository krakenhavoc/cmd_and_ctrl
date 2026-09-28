package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const cityscapeLevelerOracle = "d4d65797-2b92-4265-9169-133120c86c7f"

// The cast-trigger leg: fires while the Leveler is still a spell on
// the stack, above it, and its target is chosen by the CASTER.
func TestCityscapeLevelerCastTriggerDestroysAndMakesAPowerstone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", ManaCost: "{2}",
		Owner: opp.ID, Controller: opp.ID,
	})

	leveler := castCatalogSpell(t, g, "Cityscape Leveler", "Artifact Creature — Construct", cityscapeLevelerOracle, nil)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if p == nil || p.PickTargetMin != 0 {
		t.Fatalf("the cast trigger asks for up to one nonland permanent: %+v", p)
	}
	if !hasID(p.PickTargetCards, rock) {
		t.Fatal("the opponent's artifact should be a legal pick")
	}
	pickCard(t, g, me.ID, rock)
	if triggerOnStack(g, leveler) == nil {
		t.Fatal("the destroy trigger should sit above the Leveler on the stack")
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the targeted permanent should be destroyed")
	}
	if got := b43TokensNamed(g, opp.ID, "Powerstone"); got != 1 {
		t.Errorf("the DESTROYED PERMANENT'S CONTROLLER should get a Powerstone, got %d", got)
	}
	if !g.Battlefield.Contains(leveler) {
		t.Error("the Leveler resolves after its cast trigger")
	}
}

// The attack-trigger leg: the same printed ability, watched from the
// battlefield once the Leveler is a permanent.
func TestCityscapeLevelerAttackTriggerDestroysAndMakesAPowerstone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leveler := pushCatalogPermanent(g, me.ID, "Cityscape Leveler", "Artifact Creature — Construct", cityscapeLevelerOracle, false)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Rock", TypeLine: "Artifact", ManaCost: "{2}",
		Owner: opp.ID, Controller: opp.ID,
	})

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(leveler, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the targeted permanent should be destroyed")
	}
	if got := b43TokensNamed(g, opp.ID, "Powerstone"); got != 1 {
		t.Errorf("the destroyed permanent's controller should get a Powerstone, got %d", got)
	}
}

// CR 603.3d: with nothing else on the board when the Leveler is cast,
// "up to one" asks nothing and the spell just resolves.
func TestCityscapeLevelerCastWithNoOtherPermanentsAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	leveler := castCatalogSpell(t, g, "Cityscape Leveler", "Artifact Creature — Construct", cityscapeLevelerOracle, nil)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(leveler) {
		t.Fatal("the Leveler should resolve normally")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("no prompt should be queued with nothing to destroy")
	}
}

// Unearth returns the Leveler with haste; the printed trample keyword
// travels with it either way.
func TestCityscapeLevelerUnearth(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := seedGraveyardCard(t, g, "Cityscape Leveler", "Artifact Creature — Construct", cityscapeLevelerOracle)

	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Unearth: %v", err)
	}
	passPriorityAroundTable(t, g)

	c, ok := g.LookupCardForEffect(id)
	if !ok || !g.Battlefield.Contains(id) {
		t.Fatalf("the unearthed Leveler should be on the battlefield")
	}
	eff := c.Effective()
	if !hasKeywordForTest(eff.Abilities, "haste") {
		t.Errorf("the unearthed Leveler should have haste: %v", eff.Abilities)
	}
	if !hasKeywordForTest(eff.Abilities, "trample") {
		t.Errorf("the Leveler should keep trample: %v", eff.Abilities)
	}
}

func hasKeywordForTest(abilities []string, kw string) bool {
	for _, a := range abilities {
		if a == kw {
			return true
		}
	}
	return false
}
