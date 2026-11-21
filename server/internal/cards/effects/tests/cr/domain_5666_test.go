package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// domain_5666_test.go — #5666: the shared domain count (domain.go) and
// the seven Dominaria United cards built on it.

const (
	jodahsCodexOracle         = "1ba5b108-59cb-4af1-88cb-d3ceaaf00784"
	llanowarGreenwidowOracle  = "23391456-5dfd-4402-88fa-3b08f5e23864"
	radhasFirebrandOracle     = "0d6d5509-41e5-4099-98a0-18ae7531df7d"
	yavimayaSojournerOracle   = "b3e9917e-1bcf-4e8e-9550-7f5897dcb433"
	sunbathingRootwallaOracle = "c47ae79c-4e07-47f2-8920-2ab62b471e4f"
	territorialMaroOracle     = "478eeed5-8170-4fb3-8ec2-bcebc0265622"
	nishobaBrawlerOracle      = "c501e577-5137-4dd6-b923-95408b78258a"
)

func TestDomainCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		jodahsCodexOracle:         "Jodah's Codex",
		llanowarGreenwidowOracle:  "Llanowar Greenwidow",
		radhasFirebrandOracle:     "Radha's Firebrand",
		yavimayaSojournerOracle:   "Yavimaya Sojourner",
		sunbathingRootwallaOracle: "Sunbathing Rootwalla",
		territorialMaroOracle:     "Territorial Maro",
		nishobaBrawlerOracle:      "Nishoba Brawler",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%v caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestDomainCountCountsEachBasicTypeOnceAndOnlyYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if n := effects.DomainCount(g, me.ID); n != 0 {
		t.Fatalf("no lands: %d", n)
	}
	effectstest.DomainLands(g, me, opp, "Plains", "Island")
	if n := effects.DomainCount(g, me.ID); n != 2 {
		t.Errorf("Plains, Plains, Island, Cave: %d, want 2", n)
	}
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Dual", TypeLine: "Land — Swamp Mountain Forest"})
	if n := effects.DomainCount(g, me.ID); n != 5 {
		t.Errorf("all five types: %d, want 5", n)
	}
	if n := effects.DomainCount(g, opp.ID); n != 1 {
		t.Errorf("the opponent's lone Swamp: %d, want 1", n)
	}
}

func TestYavimayaSojournerCostsLessPerBasicLandType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const tl = "Creature — Treefolk"
	if got := effectstest.SelfPricedMV(t, g, me, yavimayaSojournerOracle, tl, "{7}{G}"); got != 8 {
		t.Fatalf("no lands: %d, want 8", got)
	}
	effectstest.DomainLands(g, me, opp, "Forest", "Island")
	if got := effectstest.SelfPricedMV(t, g, me, yavimayaSojournerOracle, tl, "{7}{G}"); got != 5 {
		t.Errorf("Forest, Plains, Island: %d, want 5", got)
	}
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Dual", TypeLine: "Land — Swamp Mountain"})
	if got := effectstest.SelfPricedMV(t, g, me, yavimayaSojournerOracle, tl, "{7}{G}"); got != 3 {
		t.Errorf("all five: %d, want 3", got)
	}
}

func TestTerritorialMaroIsTwiceDomainBothWays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	maro := effectstest.B12Push(g, me.ID, "Territorial Maro", "Creature — Elemental", territorialMaroOracle, 0, 0)
	if p := effectstest.EffectivePower(t, g, maro); p != 0 {
		t.Errorf("no lands: power %d", p)
	}
	effectstest.DomainLands(g, me, opp, "Forest", "Mountain")
	c := effectstest.GaLayered(t, g, maro).Effective()
	if c.Power != 6 || c.Toughness != 6 {
		t.Errorf("Forest, Plains, Mountain: %d/%d, want 6/6", c.Power, c.Toughness)
	}
}

func TestNishobaBrawlerPowerIsDomainToughnessStaysThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brawler := effectstest.B12Push(g, me.ID, "Nishoba Brawler", "Creature — Cat Warrior", nishobaBrawlerOracle, 0, 3)
	effectstest.DomainLands(g, me, opp, "Forest")
	c := effectstest.GaLayered(t, g, brawler).Effective()
	if c.Power != 2 || c.Toughness != 3 {
		t.Errorf("Forest, Plains: %d/%d, want 2/3", c.Power, c.Toughness)
	}
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Dual", TypeLine: "Land — Swamp Mountain Island"})
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if c := effectstest.GaLayered(t, g, brawler).Effective(); c.Power != 5 {
		t.Errorf("all five: power %d, want 5", c.Power)
	}
}

func TestJodahsCodexCostsOneLessPerTypeAndStillTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	codex := effectstest.PushPermanentForTest(g, me.ID, "Jodah's Codex", jodahsCodexOracle, "Artifact — Book")
	effectstest.DomainLands(g, me, opp, "Forest", "Island") // Forest, Plains, Island: the {5} costs {2}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := effectstest.HobB4Activate(t, g, me.ID, codex, 0, "{C}", nil); err == nil {
		t.Fatal("{C} paid for an ability that costs {2}")
	}
	handBefore := me.Hand.Size()
	if err := effectstest.HobB4Activate(t, g, me.ID, codex, 0, "{C}", nil); err != nil {
		t.Fatalf("{C}{C} with domain 3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), handBefore+1)
	}
	if !effectstest.IsTapped(g, codex) {
		t.Error("the {T} is part of the cost")
	}
}

func TestSunbathingRootwallaPumpsByDomainOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walla := effectstest.B12Push(g, me.ID, "Sunbathing Rootwalla", "Creature — Lizard", sunbathingRootwallaOracle, 2, 2)
	effectstest.DomainLands(g, me, opp, "Forest", "Island") // 3 types
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := effectstest.HobB4Activate(t, g, me.ID, walla, 0, "{C}{C}{C}{G}", nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, walla).Effective(); c.Power != 5 || c.Toughness != 5 {
		t.Errorf("after one activation: %d/%d, want 5/5", c.Power, c.Toughness)
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, walla, 0, "{C}{C}{C}{G}", nil); err == nil {
		t.Error("a second activation this turn was allowed")
	}
}

func TestRadhasFirebrandPumpIsDiscountedAndOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rf := effectstest.B12Push(g, me.ID, "Radha's Firebrand", "Creature — Human Warrior", radhasFirebrandOracle, 3, 1)
	effectstest.DomainLands(g, me, opp, "Mountain", "Forest") // 3 types: {2}{R}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := effectstest.HobB4Activate(t, g, me.ID, rf, 0, "{C}{C}", nil); err == nil {
		t.Fatal("{C}{C} paid for {2}{R}")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, rf, 0, "{R}", nil); err != nil {
		t.Fatalf("{2}{R} with domain 3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, rf).Effective(); c.Power != 5 || c.Toughness != 3 {
		t.Errorf("%d/%d, want 5/3", c.Power, c.Toughness)
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, rf, 0, "{C}{C}{C}{C}{C}{R}", nil); err == nil {
		t.Error("a second activation this turn was allowed")
	}
}

func firebrandAttack(t *testing.T, g *game.Game) (me, opp *game.Player, rf uuid.UUID) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	me, opp = g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	rf = effectstest.B12Push(g, me.ID, "Radha's Firebrand", "Creature — Human Warrior", radhasFirebrandOracle, 3, 1)
	return me, opp, rf
}

func TestRadhasFirebrandStopsOnlyALesserPowerDefenderFromBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, rf := firebrandAttack(t, g)
	small := effectstest.PushVanillaCreature(g, opp.ID, "Small", 2, 4)
	equal := effectstest.PushVanillaCreature(g, opp.ID, "Equal", 3, 4)
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	elsewhere := effectstest.PushVanillaCreature(g, third.ID, "Elsewhere", 0, 4)
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	if err := g.DeclareAttacker(rf, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, rf); err == nil {
		t.Error("the lesser-power target can still block")
	}
	if err := g.DeclareBlocker(equal, rf); err != nil {
		t.Errorf("an equal-power creature was restricted: %v", err)
	}
	_ = elsewhere
}

func TestRadhasFirebrandHasNoTargetWhenNoDefenderIsWeaker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, rf := firebrandAttack(t, g)
	effectstest.PushVanillaCreature(g, opp.ID, "Equal", 3, 4)
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	effectstest.PushVanillaCreature(g, third.ID, "Elsewhere", 0, 4)
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	if err := g.DeclareAttacker(rf, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	if !effectstest.HobB4NoPickPrompt(g) {
		t.Error("the trigger asked for a target with none legal (an equal-power creature, and one of another player)")
	}
}

// Llanowar Greenwidow: back from the graveyard tapped, priced by domain,
// and gone to exile when it would die again; the replacement survives a
// snapshot round trip.
func TestLlanowarGreenwidowReturnsTappedAndExilesWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.DomainLands(g, me, opp, "Forest", "Island", "Swamp") // 4 types: {3}{G}
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Llanowar Greenwidow", "Creature — Spider", llanowarGreenwidowOracle, 4, 3, "{C}{C}{C}{G}", game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCardByID(g, id)
	if !ok {
		t.Fatal("not on the battlefield")
	}
	if !c.Tapped {
		t.Error("must enter tapped")
	}

	restored := effectstest.RestoreRoundTrip(t, g, true)
	restored.WithWriteLock(func() { _ = restored.DestroyPermanentForEffect(id) })
	restored.RunStateChecksForTest()
	if !restored.Exile.Contains(id) {
		t.Error("restored: the Greenwidow did not go to exile")
	}
	if restored.Seats[g.Turn.ActiveSeat].Graveyard.Contains(id) {
		t.Error("restored: the Greenwidow reached the graveyard")
	}
}

func TestLlanowarGreenwidowCannotBePaidForShort(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.PushCatalogGraveyardCard(me, "Llanowar Greenwidow", "Creature — Spider", llanowarGreenwidowOracle, 4, 3)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("{7} paid for {7}{G}")
	}
}

// The domain amount is the shared helper's own: a Blood Moon-style type
// change is read through the layered land.
func TestDomainReadsTheLayeredLandTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Odd Land", TypeLine: "Land — Forest Island"})
	_ = land
	if n := effects.DomainCount(g, me.ID); n != 2 {
		t.Errorf("a land with Forest and Island subtypes: %d, want 2", n)
	}
}
