package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b739_cards_test.go — Exodus, "Auras", batch 739 (#5626). Every test
// plays the real catalog entry through the engine.

const (
	o739Robe       = "093a20e5-ff14-41c7-b16c-1f745ddf6942"
	o739Bequeathal = "2a9054ce-7276-4030-901a-cf6aa18f2b62"
	o739Cunning    = "d5e48041-5f21-48e6-bd0c-000f02dfbe8a"
	o739Gaze       = "e04f46d9-43e7-45ef-b59c-1810f30bec72"
	o739Hunger     = "eed2606b-9a27-4af3-8120-c45a1d6b5c00"
	o739Paroxysm   = "f00dd41d-4b11-491a-9e7b-12245092ab77"
)

func TestS82OldB739EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{o739Robe, o739Bequeathal, o739Cunning, o739Gaze, o739Hunger, o739Paroxysm} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func o739Aura(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := effectstest.B12Push(g, owner, name, "Enchantment — Aura", oracle, 0, 0)
	effectstest.NcdAttach(t, g, aura, host)
	return aura
}

func TestS82OldB739RobeOfMirrorsGivesShroud(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	o739Aura(t, g, me.ID, "Robe of Mirrors", o739Robe, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "shroud") {
		t.Error("the enchanted creature has no shroud")
	}
}

func TestS82OldB739BequeathalDrawsTwoWhenTheEnchantedCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	o739Aura(t, g, me.ID, "Bequeathal", o739Bequeathal, bear)
	src := effectstest.B12Creature(g, me.ID, "Source", "Creature — Ogre", 5, 5)
	hand := me.Hand.Size()
	effectstest.Cmdr2B08Damage(g, src, bear, 5)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want two cards drawn by the Aura's controller", hand, me.Hand.Size())
	}
}

func TestS82OldB739PredatoryHungerGrowsOnlyForAnOpponentsCreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	o739Aura(t, g, me.ID, "Predatory Hunger", o739Hunger, bear)
	// My own creature spell: an opponent of the Aura's controller did not cast it.
	effectstest.CastCatalogSpell(t, g, "Grizzly", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, bear); n != 0 {
		t.Fatalf("the Aura's controller casting a creature gave %d counters", n)
	}
	// The opponent's creature spell.
	effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	effectstest.CastCatalogSpell(t, g, "Their Grizzly", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, bear); n != 1 {
		t.Errorf("counters = %d, want 1 after an opponent cast a creature spell", n)
	}
}

func TestS82OldB739DizzyingGazeShootsDownAFlyerOnlyAndNeedsYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 3,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	ground := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	gaze := o739Aura(t, g, me.ID, "Dizzying Gaze", o739Gaze, mine)
	effectstest.FloatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, gaze, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ground), Strict: true}); err == nil {
		t.Fatal("a creature without flying was a legal target")
	}
	effectstest.ApaActivate(t, g, me, gaze, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(flier)})
	if n := effectstest.DamageMarkedOn(g, flier); n != 1 {
		t.Errorf("the flier has %d damage, want 1", n)
	}
	// "Enchant creature you control": an Aura on an opponent's creature is refused.
	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g2)
	me2, opp2 := g2.Seats[g2.Turn.ActiveSeat], g2.Seats[1]
	theirs := effectstest.B12Creature(g2, opp2.ID, "Their Bear", "Creature — Bear", 2, 2)
	spec, _ := effects.Lookup(o739Gaze)
	if spec.Targets == nil || spec.Targets.CardOK == nil {
		t.Fatal("Dizzying Gaze has no enchant clause")
	}
	card, _ := effectstest.BattlefieldCard(g2, theirs)
	if spec.Targets.CardOK(g2, me2.ID, card, game.ZoneBattlefield) {
		t.Error("Dizzying Gaze may enchant a creature its controller does not control")
	}
}

func TestS82OldB739ParoxysmDestroysOnALandAndPumpsOtherwise(t *testing.T) {
	for _, land := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		o739Aura(t, g, me.ID, "Paroxysm", o739Paroxysm, bear)
		typeLine := "Sorcery"
		if land {
			typeLine = "Basic Land — Forest"
		}
		effectstest.PutOnTopOfLibrary(g, opp, "Top", typeLine, "")
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.PassPriorityAroundTable(t, g)
		g.RunStateChecksForTest()
		_, alive := effectstest.BattlefieldCard(g, bear)
		if land && alive {
			t.Error("the creature survived a revealed land")
		}
		if !land {
			if !alive {
				t.Fatal("the creature died to a revealed nonland")
			}
			if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
				t.Errorf("the creature is %d/%d, want 5/5", p, tough)
			}
		}
	}
}

func TestS82OldB739CunningPumpsThenIsSacrificedAtTheNextCleanupAfterAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for me.Hand.Size() > 5 {
		me.Hand.PopTop() // no hand-size discard to wait on in the cleanup step (CR 514.1)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := o739Aura(t, g, me.ID, "Cunning", o739Cunning, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Fatalf("the enchanted creature is %d/%d, want 5/5", p, tough)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, aura); !ok {
		t.Fatal("the Aura was sacrificed before the cleanup step")
	}
	effectstest.AdvanceToStepOnThisTurn(t, g, game.StepEnd)
	if _, ok := effectstest.BattlefieldCard(g, aura); !ok {
		t.Fatal("the Aura was sacrificed by the end step")
	}
	turn := g.Turn.Seq
	sawCleanup := false
	for i := 0; i < 40 && g.Turn.Seq == turn; i++ {
		if g.Turn.Step == game.StepCleanup {
			sawCleanup = true
		}
		if _, ok := effectstest.BattlefieldCard(g, aura); !ok && !sawCleanup {
			t.Fatal("the Aura was sacrificed before the cleanup step")
		}
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if !sawCleanup {
		t.Fatal("never reached the cleanup step")
	}
	if _, ok := effectstest.BattlefieldCard(g, aura); ok {
		t.Error("Cunning was not sacrificed at the cleanup step")
	}
}
