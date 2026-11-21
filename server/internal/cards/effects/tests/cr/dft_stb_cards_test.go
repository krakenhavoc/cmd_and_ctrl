package testscr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dft_stb_cards_test.go — Aetherdrift (DFT), the second batch of cards
// from S to T (#4464). Each test plays the real catalog entry and asserts
// what its text says. The fixtures (dftABPush, dftABCast, dftABCrew,
// dftABCardRef, dftABInGraveyard) are dft_ab_cards_test.go's, and
// dftSZCycles is dft_sz_cards_test.go's.

const (
	dftSTBSabSunenOracle    = "07e15387-f346-4fa0-b556-d1c52d4acc07"
	dftSTBSalvationOracle   = "7903666e-7803-4cfc-9ed9-482a6b577c96"
	dftSTBCompactorOracle   = "7f113875-a411-463b-8b4d-91d1dee95c2a"
	dftSTBShefetOracle      = "06355d3f-a82d-4f12-a399-016a12824795"
	dftSTBSilkenOracle      = "53a59ba9-f4e9-460b-928a-ce6d7ae883f7"
	dftSTBSkyserpentOracle  = "5da8b86a-a7be-4fb5-9a45-a47d5cb4b372"
	dftSTBPileupOracle      = "8f7479e6-b2bd-46b1-96ad-34deb43565c1"
	dftSTBMechcycleOracle   = "7d347d59-0cd5-4207-86cb-5d6244c7c70a"
	dftSTBLastRideOracle    = "3ef03540-b538-466d-aa3c-da6750117ba4"
	dftSTBVelocipedeOracle  = "899007b5-6a07-4491-9104-bb537106b07d"
	dftSTBSabSunenType      = "Legendary Creature — God"
	dftSTBLastRideType      = "Legendary Artifact — Vehicle"
	dftSTBArtifactCreatureT = "Artifact Creature — Construct"
)

func TestDftSTBEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dftSTBSabSunenOracle:   "Sab-Sunen, Luxa Embodied",
		dftSTBSalvationOracle:  "Salvation Engine",
		dftSTBCompactorOracle:  "Scrap Compactor",
		dftSTBShefetOracle:     "Shefet Archfiend",
		dftSTBSilkenOracle:     "Silken Strength",
		dftSTBSkyserpentOracle: "Skyserpent Seeker",
		dftSTBPileupOracle:     "Spectacular Pileup",
		dftSTBMechcycleOracle:  "Spire Mechcycle",
		dftSTBLastRideOracle:   "The Last Ride",
		dftSTBVelocipedeOracle: "Thunderous Velocipede",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Sab-Sunen, Luxa Embodied -----------------------------------------------------

func TestDftSTBSabSunenGrowsToAnOddCountDrawsTwoAndCantAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sab := effectstest.B12Push(g, me.ID, "Sab-Sunen, Luxa Embodied", dftSTBSabSunenType, dftSTBSabSunenOracle, 6, 6)
	effectstest.SeedLibrary(me, "First Draw", "Second Draw", "Third Draw", "Fourth Draw")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	before := effectstest.HandSize(me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, sab, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d +1/+1 counters after the first main phase began, want 1", got)
	}
	if got := effectstest.HandSize(me); got != before+2 {
		t.Errorf("hand %d, want %d: one counter is odd, so it draws two", got, before+2)
	}
	if r := effectstest.EffectiveOf(t, g, sab).Restrictions; r&game.CantAttack == 0 || r&game.CantBlock == 0 {
		t.Errorf("restrictions %v: an odd count must stop it attacking and blocking", r)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sab, opp.ID); err == nil {
		t.Error("Sab-Sunen attacked with one counter on it")
	}
}

func TestDftSTBSabSunenAtAnEvenCountDrawsNothingAndAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sab := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sab-Sunen, Luxa Embodied", TypeLine: dftSTBSabSunenType,
		OracleID: dftSTBSabSunenOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1},
	})
	effectstest.SeedLibrary(me, "First Draw", "Second Draw", "Third Draw", "Fourth Draw")
	if r := effectstest.EffectiveOf(t, g, sab).Restrictions; r&game.CantAttack == 0 {
		t.Fatalf("restrictions %v with one counter, want CantAttack", r)
	}
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	before := effectstest.HandSize(me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, sab, game.CounterPlusOne); got != 2 {
		t.Fatalf("%d +1/+1 counters, want 2", got)
	}
	if got := effectstest.HandSize(me); got != before {
		t.Errorf("hand %d, want %d: two counters is even, so it draws nothing", got, before)
	}
	if r := effectstest.EffectiveOf(t, g, sab).Restrictions; r&(game.CantAttack|game.CantBlock) != 0 {
		t.Errorf("restrictions %v at an even count, want none", r)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sab, opp.ID); err != nil {
		t.Errorf("Sab-Sunen could not attack with two counters: %v", err)
	}
}

// --- Salvation Engine ------------------------------------------------------------

func TestDftSTBSalvationEnginePumpsOtherArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	engine := effectstest.DftABPush(g, me.ID, "Salvation Engine", effectstest.DftABVehicle, dftSTBSalvationOracle, 6, 10)
	construct := effectstest.B12Creature(g, me.ID, "Construct", dftSTBArtifactCreatureT, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, construct); p != 4 || tough != 4 {
		t.Errorf("the artifact creature is %d/%d, want 4/4", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the nonartifact creature is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, engine); p != 6 || tough != 10 {
		t.Errorf("the Engine is %d/%d, want its printed 6/10: the anthem is for others", p, tough)
	}
}

func TestDftSTBSalvationEngineReturnsAnArtifactCardWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	engine := effectstest.DftABPush(g, me.ID, "Salvation Engine", effectstest.DftABVehicle, dftSTBSalvationOracle, 6, 10)
	five := effectstest.B12Creature(g, me.ID, "Five", "Creature — Giant", 5, 5)
	six := effectstest.B12Creature(g, me.ID, "Six", "Creature — Giant", 6, 6)
	stone := effectstest.DftABInGraveyard(me, "Mind Stone", "Artifact")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, engine, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{five}}); err == nil {
		t.Fatal("power 5 crewed a Crew 6 Vehicle")
	}
	effectstest.DftABCrew(t, g, me, engine, 0, six)
	effectstest.DeclareAttack(t, g, opp.ID, engine)
	effectstest.PickCard(t, g, me.ID, stone)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, stone)
	if !ok {
		t.Fatal("the artifact card did not return to the battlefield")
	}
	if c.Tapped {
		t.Error("the artifact returned tapped")
	}
}

// --- Scrap Compactor -------------------------------------------------------------

func TestDftSTBScrapCompactorDealsThreeToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	compactor := effectstest.B12Push(g, me.ID, "Scrap Compactor", "Artifact", dftSTBCompactorOracle, 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, compactor, 0, game.ActivateAbilityParams{Targets: effectstest.DftABCardRef(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(compactor) {
		t.Error("the Compactor was not sacrificed")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the 3/3 survived 3 damage")
	}
}

func TestDftSTBScrapCompactorDestroysAnUncrewedVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	compactor := effectstest.B12Push(g, me.ID, "Scrap Compactor", "Artifact", dftSTBCompactorOracle, 0, 0)
	vehicle := effectstest.DftABPush(g, opp.ID, "Their Vehicle", effectstest.DftABVehicle, "", 3, 3)
	rock := effectstest.B12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 6)
	if err := g.ActivateCatalogAbility(me.ID, compactor, 1, game.ActivateAbilityParams{Targets: effectstest.DftABCardRef(rock)}); err == nil {
		t.Fatal("the destroy row targeted a noncreature, non-Vehicle artifact")
	}
	effectstest.FillPool(me, 6)
	if err := g.ActivateCatalogAbility(me.ID, compactor, 1, game.ActivateAbilityParams{Targets: effectstest.DftABCardRef(vehicle)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vehicle) {
		t.Error("the uncrewed Vehicle was not destroyed")
	}
	if g.Battlefield.Contains(compactor) {
		t.Error("the Compactor was not sacrificed")
	}
}

// --- Shefet Archfiend ------------------------------------------------------------

func TestDftSTBShefetArchfiendShrinksEveryOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ox", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Elk", 2, 2)
	effectstest.AdvanceToMain(t, g)
	fiend := effectstest.DftABCast(t, g, me, game.Card{Name: "Shefet Archfiend", TypeLine: "Creature — Demon",
		OracleID: dftSTBShefetOracle, Power: 5, Toughness: 5})
	if g.Battlefield.Contains(small) || g.Battlefield.Contains(mine) {
		t.Error("a 2/2 survived -2/-2")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, big); p != 1 || tough != 1 {
		t.Errorf("the 3/3 is %d/%d, want 1/1", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, fiend); p != 5 || tough != 5 {
		t.Errorf("the Archfiend is %d/%d, want 5/5: it spares itself", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, fiend, "flying") {
		t.Error("the Archfiend has no flying")
	}
}

func TestDftSTBShefetArchfiendCyclesForTwo(t *testing.T) {
	dftSZCycles(t, "Shefet Archfiend", "Creature — Demon", dftSTBShefetOracle, "{C}{C}", 0)
}

// --- Silken Strength -------------------------------------------------------------

func TestDftSTBSilkenStrengthUntapsAndPumpsAVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vehicle := effectstest.DftABPush(g, me.ID, "My Vehicle", effectstest.DftABVehicle, "", 3, 3)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, vehicle).Tapped = true })
	effectstest.CastCatalogSpell(t, g, "Silken Strength", "Enchantment — Aura", dftSTBSilkenOracle, effectstest.DftABCardRef(vehicle))
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, vehicle)
	if c.Tapped {
		t.Error("the enchanted Vehicle was not untapped")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, vehicle); p != 4 || tough != 5 {
		t.Errorf("the enchanted Vehicle is %d/%d, want 4/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, vehicle, "reach") {
		t.Error("the enchanted Vehicle has no reach")
	}
}

func TestDftSTBSilkenStrengthRefusesANoncreatureArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Permanent(g, me.ID, "Mind Stone", "Artifact")
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Silken Strength", "Enchantment — Aura", dftSTBSilkenOracle, effectstest.DftABCardRef(rock)); err == nil {
		t.Error("Silken Strength enchanted a noncreature, non-Vehicle artifact")
	}
}

// --- Skyserpent Seeker -----------------------------------------------------------

func TestDftSTBSkyserpentSeekerPutsTwoLandsTappedAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seeker := effectstest.B12Push(g, me.ID, "Skyserpent Seeker", "Creature — Snake", dftSTBSkyserpentOracle, 1, 1)
	extra := effectstest.LibTop(me, game.Card{Name: "Extra Island", TypeLine: "Basic Land — Island"})
	island := effectstest.LibTop(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	opt := effectstest.LibTop(me, game.Card{Name: "Opt", TypeLine: "Instant"})
	forest := effectstest.LibTop(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	bear := effectstest.LibTop(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{forest, island} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok || !c.Tapped {
			t.Errorf("land %s on the battlefield %v, tapped %v, want tapped", id, ok, c.Tapped)
		}
	}
	if g.Battlefield.Contains(extra) {
		t.Error("the reveal went past the second land")
	}
	if bottom := effectstest.PlBottomIDs(me, 2); !bottom[bear] || !bottom[opt] {
		t.Error("the revealed nonlands are not on the bottom of the library")
	}
	if got := effectstest.CountersOn(g, seeker, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	effectstest.FillPool(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, seeker, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrAbilityExhausted) {
		t.Fatalf("second activation: %v, want ErrAbilityExhausted", err)
	}
}

// --- Spectacular Pileup ----------------------------------------------------------

func TestDftSTBSpectacularPileupDestroysIndestructibleCreaturesAndVehicles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	god := effectstest.B12Push(g, opp.ID, "Sab-Sunen, Luxa Embodied", dftSTBSabSunenType, dftSTBSabSunenOracle, 6, 6)
	vehicle := effectstest.DftABPush(g, opp.ID, "Their Vehicle", effectstest.DftABVehicle, "", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, god, "indestructible") {
		t.Fatal("the test God is not indestructible")
	}
	effectstest.CastCatalogSpell(t, g, "Spectacular Pileup", "Sorcery", dftSTBPileupOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"the indestructible God": god, "the uncrewed Vehicle": vehicle, "the Bear": bear} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", name)
		}
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("a noncreature, non-Vehicle artifact was destroyed")
	}
}

func TestDftSTBSpectacularPileupCyclesForTwo(t *testing.T) {
	dftSZCycles(t, "Spectacular Pileup", "Sorcery", dftSTBPileupOracle, "{C}{C}", 0)
}

// --- Spire Mechcycle -------------------------------------------------------------

func TestDftSTBSpireMechcycleExhaustsByTappingAVehicleAndCountsMountsAndVehicles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cycle := effectstest.DftABPush(g, me.ID, "Spire Mechcycle", effectstest.DftABVehicle, dftSTBMechcycleOracle, 5, 4)
	other := effectstest.DftABPush(g, me.ID, "My Vehicle", effectstest.DftABVehicle, "", 2, 2)
	steed := effectstest.B12Creature(g, me.ID, "Steed", "Creature — Horse Mount", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DftABPush(g, opp.ID, "Their Vehicle", effectstest.DftABVehicle, "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, cycle, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a creature that is neither a Mount nor a Vehicle paid the exhaust cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, cycle, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{other}}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, other); !c.Tapped {
		t.Error("the Vehicle tapped for the cost is untapped")
	}
	c, _ := effectstest.BattlefieldCard(g, cycle)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Fatalf("the Mechcycle is not an artifact creature: %q", c.TypeLine)
	}
	if got := effectstest.CountersOn(g, cycle, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2 (your other Vehicle and your Mount)", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, cycle, "haste") {
		t.Error("the Mechcycle has no haste")
	}
	if err := g.ActivateCatalogAbility(me.ID, cycle, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{steed}}); err == nil {
		t.Error("the exhaust ability was activated twice")
	}
}

// --- The Last Ride ---------------------------------------------------------------

func TestDftSTBTheLastRideShrinksByYourLifeTotal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ride := effectstest.DftABPush(g, me.ID, "The Last Ride", dftSTBLastRideType, dftSTBLastRideOracle, 13, 13)
	pilot := effectstest.B12Creature(g, me.ID, "Pilot", "Creature — Human Pilot", 2, 2)
	effectstest.AdvanceToMain(t, g)
	life := effectstest.LifeOf(g, me.ID)
	if p, tough := effectstest.EffectivePTOf(t, g, ride); p != 13-life || tough != 13-life {
		t.Errorf("at %d life the Ride is %d/%d, want %d/%d", life, p, tough, 13-life, 13-life)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 10-life) })
	if p, tough := effectstest.EffectivePTOf(t, g, ride); p != 3 || tough != 3 {
		t.Errorf("at 10 life the Ride is %d/%d, want 3/3", p, tough)
	}
	effectstest.DftABCrew(t, g, me, ride, 1, pilot)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(uuid.Nil, me.ID, 6) })
	if p, tough := effectstest.EffectivePTOf(t, g, ride); p != 9 || tough != 9 {
		t.Errorf("at 4 life the Ride is %d/%d, want 9/9", p, tough)
	}
}

func TestDftSTBTheLastRideDrawsForManaAndTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ride := effectstest.DftABPush(g, me.ID, "The Last Ride", dftSTBLastRideType, dftSTBLastRideOracle, 13, 13)
	effectstest.SeedLibrary(me, "Drawn")
	effectstest.AdvanceToMain(t, g)
	life := effectstest.LifeOf(g, me.ID)
	effectstest.FillPool(me, 2)
	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, ride, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life-2 {
		t.Errorf("life %d, want %d", got, life-2)
	}
	if !effectstest.B02bHandHasNamed(me, "Drawn") {
		t.Error("the ability drew nothing")
	}
}

// --- Thunderous Velocipede -------------------------------------------------------

func TestDftSTBThunderousVelocipedeGivesOneCounterOrThreeByManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.DftABPush(g, me.ID, "Thunderous Velocipede", effectstest.DftABVehicle, dftSTBVelocipedeOracle, 5, 5)
	effectstest.AdvanceToMain(t, g)
	small := effectstest.DftABCast(t, g, me, game.Card{Name: "Small", TypeLine: "Creature — Bear", ManaCost: "{3}{G}", Power: 2, Toughness: 2})
	if got := effectstest.CountersOn(g, small, game.CounterPlusOne); got != 1 {
		t.Errorf("a mana value 4 creature entered with %d counters, want 1", got)
	}
	big := effectstest.DftABCast(t, g, me, game.Card{Name: "Big", TypeLine: "Creature — Giant", ManaCost: "{4}{G}", Power: 5, Toughness: 5})
	if got := effectstest.CountersOn(g, big, game.CounterPlusOne); got != 3 {
		t.Errorf("a mana value 5 creature entered with %d counters, want 3", got)
	}
	vehicle := effectstest.DftABCast(t, g, me, game.Card{Name: "Cart", TypeLine: effectstest.DftABVehicle, ManaCost: "{2}", Power: 3, Toughness: 3})
	if got := effectstest.CountersOn(g, vehicle, game.CounterPlusOne); got != 1 {
		t.Errorf("a mana value 2 Vehicle entered with %d counters, want 1", got)
	}
	rock := effectstest.DftABCast(t, g, me, game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"})
	if got := effectstest.CountersOn(g, rock, game.CounterPlusOne); got != 0 {
		t.Errorf("a noncreature, non-Vehicle artifact entered with %d counters, want 0", got)
	}
}

func TestDftSTBThunderousVelocipedeLeavesOpponentsCreaturesAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.DftABPush(g, opp.ID, "Thunderous Velocipede", effectstest.DftABVehicle, dftSTBVelocipedeOracle, 5, 5)
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.DftABCast(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	if got := effectstest.CountersOn(g, mine, game.CounterPlusOne); got != 0 {
		t.Errorf("a creature entering under the Velocipede controller's opponent got %d counters, want 0", got)
	}
}
