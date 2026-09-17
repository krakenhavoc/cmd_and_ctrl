package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	sharpshooterOracle = "d81285b7-a718-411a-8be3-ecc0cfe0bcb0"
	traxosOracle       = "c1c78144-b335-4d22-a668-9173ab6a0d04"
	wallOfFrostOracle  = "741e4f32-0587-40fa-a73d-5bcf66b52348"
	kefnetMonument     = "b6294891-79e6-4f2a-a82d-6cffce968356"
	claustrophobiaOrcl = "62d8c8c8-bc24-42f2-9e2e-9efd08e47bb1"
)

func hasTriggerFor(g *game.Game, source uuid.UUID) bool {
	if triggerOnStack(g, source) != nil {
		return true
	}
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == source {
			return true
		}
	}
	return false
}

func TestGoblinSharpshooterUntapsOnlyAfterACreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shooter := pushCatalogPermanent(g, me.ID, "Goblin Sharpshooter", "Creature — Goblin", sharpshooterOracle, false)
	deadCreature := pushPermanentForTest(g, opp.ID, "Dead Bear", "", "Creature — Bear")
	if err := g.TapCard(shooter, true); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(deadCreature); err != nil {
			t.Errorf("sacrifice creature: %v", err)
		}
	})
	if !hasTriggerFor(g, shooter) {
		t.Fatal("creature death did not queue Goblin Sharpshooter's untap trigger")
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, shooter); c.Tapped {
		t.Error("Goblin Sharpshooter stayed tapped after a creature died")
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	shooter = pushCatalogPermanent(g, me.ID, "Goblin Sharpshooter", "Creature — Goblin", sharpshooterOracle, false)
	deadArtifact := pushPermanentForTest(g, opp.ID, "Dead Rock", "", "Artifact")
	if err := g.TapCard(shooter, true); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(deadArtifact); err != nil {
			t.Errorf("sacrifice artifact: %v", err)
		}
	})
	if hasTriggerFor(g, shooter) {
		t.Fatal("noncreature death queued Goblin Sharpshooter's creature-death trigger")
	}
	if c, _ := battlefieldCard(g, shooter); !c.Tapped {
		t.Error("Goblin Sharpshooter untapped for a noncreature death")
	}
}

func TestTraxosEntersTappedAndUntapsForHistoricCastOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := enterFromHand(t, g, me.ID, "Traxos, Scourge of Kroog", "Legendary Artifact Creature — Construct", traxosOracle)
	if c, _ := battlefieldCard(g, id); !c.Tapped {
		t.Fatal("Traxos did not enter tapped")
	}

	castCatalogSpell(t, g, "Historic Relic", "Artifact", "test-historic-relic", nil)
	if !hasTriggerFor(g, id) {
		t.Fatal("historic artifact cast did not queue Traxos's untap trigger")
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, id); c.Tapped {
		t.Error("Traxos stayed tapped after its historic-cast trigger resolved")
	}

	castCatalogSpell(t, g, "Ordinary Instant", "Instant", "test-ordinary-instant", nil)
	if hasTriggerFor(g, id) {
		t.Fatal("nonhistoric instant queued Traxos's historic-cast trigger")
	}
}

func TestWallOfFrostMarksTheAttackerThatItBlocks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := pushCatalogPermanent(g, opp.ID, "Wall of Frost", "Creature — Wall", wallOfFrostOracle, false)
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if !hasTriggerFor(g, wall) {
		t.Fatal("Wall of Frost did not queue a trigger when it blocked")
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, attacker)
	if len(c.NextUntapSkips) != 1 || c.NextUntapSkips[0].Player != uuid.Nil {
		t.Errorf("attacker marker = %#v, want one controller-keyed marker", c.NextUntapSkips)
	}
	wallCard, _ := battlefieldCard(g, wall)
	if len(wallCard.NextUntapSkips) != 0 {
		t.Errorf("blocking Wall of Frost received its own marker: %#v", wallCard.NextUntapSkips)
	}
}

func TestKefnetsMonumentDiscountsBlueCreaturesAndFreezesAnOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	monument := pushCatalogPermanent(g, me.ID, "Kefnet's Monument", "Legendary Artifact", kefnetMonument, false)
	if got := priceInHand(t, g, me, "Blue Bear", "Creature — Bear", "{1}{U}"); got != 1 {
		t.Errorf("blue creature cost under Kefnet's Monument = %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Green Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("nonblue creature cost under Kefnet's Monument = %d, want 2", got)
	}

	target := pushPermanentForTest(g, opp.ID, "Opposing Bear", "", "Creature — Bear")
	castAndResolveCreature(t, g, "Blue Creature", "Creature — Wizard", "test-blue-creature")
	pickCard(t, g, me.ID, target)
	if !hasTriggerFor(g, monument) {
		t.Fatal("Kefnet's Monument trigger did not reach the stack after choosing a target")
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, target)
	if len(c.NextUntapSkips) != 1 || c.NextUntapSkips[0].Player != uuid.Nil {
		t.Errorf("opposing target marker = %#v, want one controller-keyed marker", c.NextUntapSkips)
	}
}

func TestClaustrophobiaAttachesTapsAndHoldsItsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := pushPermanentForTest(g, me.ID, "Host", "", "Creature — Bear")
	aura := castCatalogSpell(t, g, "Claustrophobia", auraTypeLine, claustrophobiaOrcl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	if got := attachmentHostOf(t, g, aura); got.Kind != game.TargetCard || got.ID != host {
		t.Fatalf("Claustrophobia attached to %+v, want host %s", got, host)
	}
	if c, _ := battlefieldCard(g, host); !c.Tapped {
		t.Fatal("Claustrophobia's ETB trigger did not tap its enchanted creature")
	}
	advanceToUpkeepOf(t, g, 0)
	if c, _ := battlefieldCard(g, host); !c.Tapped {
		t.Error("enchanted creature untapped during its controller's untap step")
	}
}
