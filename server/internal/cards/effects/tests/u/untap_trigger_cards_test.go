package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func hasTriggerFor(g *game.Game, source uuid.UUID) bool {
	if effectstest.TriggerOnStack(g, source) != nil {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shooter := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Sharpshooter", "Creature — Goblin", effectstest.SharpshooterOracle, false)
	deadCreature := effectstest.PushPermanentForTest(g, opp.ID, "Dead Bear", "", "Creature — Bear")
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
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, shooter); c.Tapped {
		t.Error("Goblin Sharpshooter stayed tapped after a creature died")
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	shooter = effectstest.PushCatalogPermanent(g, me.ID, "Goblin Sharpshooter", "Creature — Goblin", effectstest.SharpshooterOracle, false)
	deadArtifact := effectstest.PushPermanentForTest(g, opp.ID, "Dead Rock", "", "Artifact")
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
	if c, _ := effectstest.BattlefieldCard(g, shooter); !c.Tapped {
		t.Error("Goblin Sharpshooter untapped for a noncreature death")
	}
}

// Goblin Sharpshooter's own death triggers its untap ("whenever a
// creature dies"), and by the time the trigger resolves the
// Sharpshooter is in the graveyard. Untapping it does nothing; it is
// not an effect error (ADR 0126 PR 5's catalog soak, seed 2026100601).
func TestGoblinSharpshooterDyingUntapsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shooter := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Sharpshooter", "Creature — Goblin", effectstest.SharpshooterOracle, false)
	if err := g.TapCard(shooter, true); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(shooter); err != nil {
			t.Errorf("sacrifice: %v", err)
		}
	})
	if !hasTriggerFor(g, shooter) {
		t.Fatal("Goblin Sharpshooter's own death did not trigger its untap (CR 603.10a looks back)")
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventEffectError {
				t.Errorf("effect error resolving the untap: %s", ev.ErrorMsg)
			}
		}
	})
}

func TestTraxosEntersTappedAndUntapsForHistoricCastOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.EnterFromHand(t, g, me.ID, "Traxos, Scourge of Kroog", "Legendary Artifact Creature — Construct", effectstest.TraxosOracle)
	if c, _ := effectstest.BattlefieldCard(g, id); !c.Tapped {
		t.Fatal("Traxos did not enter tapped")
	}

	effectstest.CastCatalogSpell(t, g, "Historic Relic", "Artifact", "test-historic-relic", nil)
	if !hasTriggerFor(g, id) {
		t.Fatal("historic artifact cast did not queue Traxos's untap trigger")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, id); c.Tapped {
		t.Error("Traxos stayed tapped after its historic-cast trigger resolved")
	}

	effectstest.CastCatalogSpell(t, g, "Ordinary Instant", "Instant", "test-ordinary-instant", nil)
	if hasTriggerFor(g, id) {
		t.Fatal("nonhistoric instant queued Traxos's historic-cast trigger")
	}
}

func TestWallOfFrostMarksTheAttackerThatItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.PushCatalogPermanent(g, opp.ID, "Wall of Frost", "Creature — Wall", effectstest.WallOfFrostOracle, false)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	// Block assignments are staged until the declaration locks in
	// (#830) — since #1501, when the defender finishes, priority being
	// parked until then. Stop when the trigger is queued, before
	// another round of passes would resolve it.
	effectstest.FinishBlockDeclarations(t, g)
	for i := 0; i < len(g.Seats) && !hasTriggerFor(g, wall); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("lock in blocks: %v", err)
		}
	}
	if !hasTriggerFor(g, wall) {
		t.Fatal("Wall of Frost did not queue a trigger when it blocked")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, attacker)
	if len(c.NextUntapSkips) != 1 || c.NextUntapSkips[0].Player != uuid.Nil {
		t.Errorf("attacker marker = %#v, want one controller-keyed marker", c.NextUntapSkips)
	}
	wallCard, _ := effectstest.BattlefieldCard(g, wall)
	if len(wallCard.NextUntapSkips) != 0 {
		t.Errorf("blocking Wall of Frost received its own marker: %#v", wallCard.NextUntapSkips)
	}
}

func TestKefnetsMonumentDiscountsBlueCreaturesAndFreezesAnOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	monument := effectstest.PushCatalogPermanent(g, me.ID, "Kefnet's Monument", "Legendary Artifact", effectstest.KefnetMonument, false)
	if got := effectstest.PriceInHand(t, g, me, "Blue Bear", "Creature — Bear", "{1}{U}"); got != 1 {
		t.Errorf("blue creature cost under Kefnet's Monument = %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Green Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("nonblue creature cost under Kefnet's Monument = %d, want 2", got)
	}

	target := effectstest.PushPermanentForTest(g, opp.ID, "Opposing Bear", "", "Creature — Bear")
	effectstest.CastAndResolveCreature(t, g, "Blue Creature", "Creature — Wizard", "test-blue-creature")
	effectstest.PickCard(t, g, me.ID, target)
	if !hasTriggerFor(g, monument) {
		t.Fatal("Kefnet's Monument trigger did not reach the stack after choosing a target")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, target)
	if len(c.NextUntapSkips) != 1 || c.NextUntapSkips[0].Player != uuid.Nil {
		t.Errorf("opposing target marker = %#v, want one controller-keyed marker", c.NextUntapSkips)
	}
}

func TestClaustrophobiaAttachesTapsAndHoldsItsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.PushPermanentForTest(g, me.ID, "Host", "", "Creature — Bear")
	aura := effectstest.CastCatalogSpell(t, g, "Claustrophobia", effectstest.AuraTypeLine, effectstest.ClaustrophobiaOrcl,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AttachmentHostOf(t, g, aura); got.Kind != game.TargetCard || got.ID != host {
		t.Fatalf("Claustrophobia attached to %+v, want host %s", got, host)
	}
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Fatal("Claustrophobia's ETB trigger did not tap its enchanted creature")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Error("enchanted creature untapped during its controller's untap step")
	}
}
