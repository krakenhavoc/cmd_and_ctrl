package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b421_cards_test.go — Return to Ravnica, removal (S82 OLD batch 421,
// #5566). Each test plays the real catalog entry.

const (
	oldB421SurveyOracle      = "35df2d01-2f9f-49a7-8fa1-cfaa52e5c7b6"
	oldB421StreetSweepOracle = "613d1a33-9fe5-48db-8277-955a055a787e"
	oldB421AerialOracle      = "6bac1495-16cf-4e1e-9126-482f85322c78"
	oldB421ArrowOracle       = "6c29add6-8524-4a33-9f2e-f792032dd6cd"
	oldB421EvidenceOracle    = "79aa86f5-bee4-498c-a0cd-719c05b136b2"
	oldB421BatterhornOracle  = "b9dd5841-7e4b-40bf-9969-e8fb5df46763"
)

func TestOldB421EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB421SurveyOracle:      "Survey the Wreckage",
		oldB421StreetSweepOracle: "Street Sweeper",
		oldB421AerialOracle:      "Aerial Predation",
		oldB421ArrowOracle:       "Avenging Arrow",
		oldB421EvidenceOracle:    "Destroy the Evidence",
		oldB421BatterhornOracle:  "Batterhorn",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB421SurveyTheWreckageDestroysALandAndMakesAGoblin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Survey the Wreckage", "Sorcery", oldB421SurveyOracle, effectstest.B16TargetCard(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Goblin"); got != 1 {
		t.Errorf("%d Goblin tokens, want 1", got)
	}
}

func TestOldB421SurveyTheWreckageRefusesANonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Survey the Wreckage", "Sorcery", oldB421SurveyOracle, effectstest.B16TargetCard(bear)); err == nil {
		t.Error("a creature was accepted as the land target")
	}
}

func TestOldB421AerialPredationDestroysAFlyerAndGainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	flyer := effectstest.ReqCreature(g, opp.ID, "Flyer", "", 2, 2, "flying")
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Aerial Predation", "Instant", oldB421AerialOracle, effectstest.B16TargetCard(flyer))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flyer) {
		t.Error("the flyer survived")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestOldB421AerialPredationRefusesAGroundCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Aerial Predation", "Instant", oldB421AerialOracle, effectstest.B16TargetCard(bear)); err == nil {
		t.Error("a creature without flying was accepted")
	}
}

func TestOldB421AvengingArrowDestroysACreatureThatDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	fighter := effectstest.ReqCreature(g, opp.ID, "Fighter", "", 2, 2)
	effectstest.DealCombatDamageToPlayer(g, fighter, g.Seats[0].ID, 2)
	effectstest.CastCatalogSpell(t, g, "Avenging Arrow", "Instant", oldB421ArrowOracle, effectstest.B16TargetCard(fighter))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fighter) {
		t.Error("the creature that dealt damage survived")
	}
}

func TestOldB421AvengingArrowRefusesACreatureThatDealtNoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	idle := effectstest.ReqCreature(g, g.Seats[1].ID, "Idle", "", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Avenging Arrow", "Instant", oldB421ArrowOracle, effectstest.B16TargetCard(idle)); err == nil {
		t.Error("a creature that dealt no damage was accepted")
	}
}

func TestOldB421DestroyTheEvidenceMillsTheLandsControllerUntilALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	effectstest.SetLibraryTopFirst(opp,
		effectstest.LibraryCreature("A"), effectstest.LibraryCreature("B"), effectstest.LibraryLand("Forest"),
		effectstest.LibraryCreature("C"))
	gy := opp.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Destroy the Evidence", "Sorcery", oldB421EvidenceOracle, effectstest.B16TargetCard(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
	// The destroyed land plus the run A, B, Forest.
	if got := opp.Graveyard.Size() - gy; got != 4 {
		t.Errorf("%d cards reached the graveyard, want 4 (the land and a run of three)", got)
	}
	if opp.Library.Size() != 1 {
		t.Errorf("library has %d cards, want 1", opp.Library.Size())
	}
}

func TestOldB421BatterhornMayDestroyAnArtifact(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
		effectstest.CastHoldCreature(t, g, "Batterhorn", "Creature — Beast", oldB421BatterhornOracle, 4, 3)
		effectstest.B04WaitForPick(t, g, me.ID)
		effectstest.AnswerPickTarget(t, g, rock)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(rock) == yes {
			t.Errorf("answered %v but the artifact's presence is %v", yes, g.Battlefield.Contains(rock))
		}
	}
}

func TestOldB421StreetSweeperDestroysAurasOnTargetLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sweeper := effectstest.B12Push(g, me.ID, "Street Sweeper", "Artifact Creature — Construct", oldB421StreetSweepOracle, 4, 6)
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	other := effectstest.B12Permanent(g, opp.ID, "Other Forest", "Basic Land — Forest")
	aura := effectstest.B12Push(g, opp.ID, "Wild Growth", "Enchantment — Aura", "706ae742-1807-44b7-a4fa-f2e26f61519a", 0, 0)
	otherAura := effectstest.B12Push(g, opp.ID, "Wild Growth", "Enchantment — Aura", "706ae742-1807-44b7-a4fa-f2e26f61519a", 0, 0)
	if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: land}); err != nil {
		t.Fatalf("AttachForEffect: %v", err)
	}
	if err := g.AttachForEffect(otherAura, game.TargetRef{Kind: game.TargetCard, ID: other}); err != nil {
		t.Fatalf("AttachForEffect: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if !g.Battlefield.Contains(aura) || !g.Battlefield.Contains(otherAura) {
		t.Fatalf("an Aura fell off before combat: %v %v", g.Battlefield.Contains(aura), g.Battlefield.Contains(otherAura))
	}
	if err := g.DeclareAttacker(sweeper, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.AnswerPickTarget(t, g, land)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura on the target land survived")
	}
	if !g.Battlefield.Contains(otherAura) {
		t.Error("an Aura on another land was destroyed")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land itself was destroyed")
	}
}

var _ = uuid.Nil
