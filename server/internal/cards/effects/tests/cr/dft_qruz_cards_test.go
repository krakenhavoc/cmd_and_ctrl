package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dft_qruz_cards_test.go — Aetherdrift (DFT), cards from Q, R and U to Z
// (#4464). Each test plays the real catalog entry and asserts what its
// text says. The fixtures are dft_ab_cards_test.go's and
// dft_st_cards_test.go's.

const (
	dftQRUZAetherhiveOracle   = "49d149d6-de94-48cf-974b-6bb4a37ce77c"
	dftQRUZImperiosaurOracle  = "83a76601-ef59-4888-af70-5088dd1f504c"
	dftQRUZRidesEndOracle     = "89e209f3-c474-453f-9659-ccebfc88bc0f"
	dftQRUZRipclawOracle      = "6b2ffdb8-3681-4d64-ac03-8c860082e60e"
	dftQRUZRoadRageOracle     = "da052082-b440-4abb-9c69-05d4d13e5fd2"
	dftQRUZBlowoutOracle      = "03008c7d-c738-473d-a968-1922b2899d1a"
	dftQRUZBoostbuggyOracle   = "2219e636-74d0-4f82-8704-e6d66c5f7f09"
	dftQRUZRunOverOracle      = "2b1bec38-5aaf-4ca0-9a04-fb4367c59c1e"
	dftQRUZVeloheartOracle    = "fa123b21-518a-4f1d-9629-c44576d45f21"
	dftQRUZBeastriderOracle   = "29f1484c-91b6-436a-a6be-3ab8b58c0fc1"
	dftQRUZQuickwelderOracle  = "87e8c044-5f83-437d-a89e-8a95559045f6"
	dftQRUZWillowrushOracle   = "c52eaa87-9251-4a47-83fd-04e582ade612"
	dftQRUZWickerfolkOracle   = "8ba5af76-2485-4fd6-bd2e-fe90d1f2691c"
	dftQRUZWreckRemoverOracle = "89c89499-23a8-412d-9406-1a7ffc88c64d"
	dftQRUZWretchedDollOracle = "628cf9ab-8f8b-49b6-b761-2450088ad03b"
)

func TestDftQRUZEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dftQRUZAetherhiveOracle:   "Rangers' Aetherhive",
		dftQRUZImperiosaurOracle:  "Regal Imperiosaur",
		dftQRUZRidesEndOracle:     "Ride's End",
		dftQRUZRipclawOracle:      "Ripclaw Wrangler",
		dftQRUZRoadRageOracle:     "Road Rage",
		dftQRUZBlowoutOracle:      "Roadside Blowout",
		dftQRUZBoostbuggyOracle:   "Rocketeer Boostbuggy",
		dftQRUZRunOverOracle:      "Run Over",
		dftQRUZVeloheartOracle:    "Veloheart Bike",
		dftQRUZBeastriderOracle:   "Veteran Beastrider",
		dftQRUZQuickwelderOracle:  "Voyager Quickwelder",
		dftQRUZWillowrushOracle:   "Willowrush Verge",
		dftQRUZWickerfolkOracle:   "Wreckage Wickerfolk",
		dftQRUZWreckRemoverOracle: "Wreck Remover",
		dftQRUZWretchedDollOracle: "Wretched Doll",
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

// --- fixtures ------------------------------------------------------------------

// dftQRUZDamageOn is the damage marked on a permanent.
func dftQRUZDamageOn(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.DamageMarked
}

// dftQRUZTapped reports whether a permanent is tapped.
func dftQRUZTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// --- Rangers' Aetherhive ------------------------------------------------------------

func TestDftQRUZRangersAetherhiveMakesAThopterOnYourExhaustActivationOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hive := effectstest.DftABPush(g, me.ID, "Rangers' Aetherhive", effectstest.DftABVehicle, dftQRUZAetherhiveOracle, 3, 5)
	buggy := effectstest.DftABPush(g, me.ID, "Rocketeer Boostbuggy", effectstest.DftABVehicle, dftQRUZBoostbuggyOracle, 3, 2)
	pilot := effectstest.B12Creature(g, me.ID, "Pilot", "Creature — Goblin Pilot", 1, 1)
	effectstest.AdvanceToMain(t, g)

	// Crewing is not an exhaust ability.
	effectstest.DftABCrew(t, g, me, hive, 0, pilot)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Thopter"); n != 0 {
		t.Fatalf("%d Thopters after crewing; only an exhaust activation makes one", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, hive, "vigilance") {
		t.Error("the crewed Aetherhive has no vigilance")
	}

	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, buggy, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the Boostbuggy's exhaust: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Thopter"); n != 1 {
		t.Fatalf("%d Thopters after an exhaust activation, want 1", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, dftQRUZTokenNamed(t, g, me.ID, "Thopter"), "flying") {
		t.Error("the Thopter has no flying")
	}

	// An opponent's exhaust activation is not "you activate".
	theirs := effectstest.DftABPush(g, opp.ID, "Rocketeer Boostbuggy", effectstest.DftABVehicle, dftQRUZBoostbuggyOracle, 3, 2)
	effectstest.FillPool(opp, 3)
	if err := g.ActivateCatalogAbility(opp.ID, theirs, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the opponent's exhaust: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Thopter"); n != 1 {
		t.Errorf("%d Thopters after an opponent's exhaust activation, want still 1", n)
	}
}

// dftQRUZTokenNamed is the id of a token with that name under controller.
func dftQRUZTokenNamed(t *testing.T, g *game.Game, controller uuid.UUID, name string) uuid.UUID {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && effects.IsToken(c) && c.Name == name {
			return c.InstanceID
		}
	}
	t.Fatalf("no %s token", name)
	return uuid.Nil
}

// --- Regal Imperiosaur --------------------------------------------------------------

func TestDftQRUZRegalImperiosaurPumpsYourOtherDinosaursOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	imperiosaur := effectstest.DftABPush(g, me.ID, "Regal Imperiosaur", "Creature — Dinosaur", dftQRUZImperiosaurOracle, 5, 4)
	dino := effectstest.B12Creature(g, me.ID, "Raptor", "Creature — Dinosaur", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Raptor", "Creature — Dinosaur", 2, 2)
	if p, tg := effectstest.EffectivePower(t, g, dino), effectstest.EffectiveToughness(t, g, dino); p != 3 || tg != 3 {
		t.Errorf("your other Dinosaur is %d/%d, want 3/3", p, tg)
	}
	if p, tg := effectstest.EffectivePower(t, g, imperiosaur), effectstest.EffectiveToughness(t, g, imperiosaur); p != 5 || tg != 4 {
		t.Errorf("the Imperiosaur is %d/%d, want its printed 5/4", p, tg)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Dinosaur has power %d, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's Dinosaur has power %d, want 2", p)
	}
}

// --- Ride's End -----------------------------------------------------------------------

func TestDftQRUZRidesEndCostsThreeLessAgainstATappedPermanentAndExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	untapped := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	vehicle := effectstest.DftABPush(g, opp.ID, "Their Wrangler", effectstest.DftABVehicle, dftQRUZRipclawOracle, 4, 3)
	g.WithWriteLock(func() { effectstest.B16Tap(g, vehicle) })
	effectstest.AdvanceToMain(t, g)
	card := game.Card{Name: "Ride's End", TypeLine: "Instant", ManaCost: "{4}{W}", OracleID: dftQRUZRidesEndOracle}

	if err := effectstest.SocB19CastWithPool(g, me, card, "W", 2, game.CastSpellParams{Targets: effectstest.CardRefs(untapped)}); err == nil {
		t.Fatal("cast for two mana against an untapped creature")
	}
	if err := effectstest.SocB19CastWithPool(g, me, card, "W", 2, game.CastSpellParams{Targets: effectstest.CardRefs(vehicle)}); err != nil {
		t.Fatalf("cast for {1}{W} against a tapped, uncrewed Vehicle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(vehicle) {
		t.Error("the Vehicle was not exiled")
	}
}

// --- Ripclaw Wrangler -----------------------------------------------------------------

func TestDftQRUZRipclawWranglerMakesEachOpponentDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, p := range g.Seats {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	}
	effectstest.AdvanceToMain(t, g)
	wrangler := effectstest.DftABCast(t, g, me, game.Card{Name: "Ripclaw Wrangler", TypeLine: effectstest.DftABVehicle,
		OracleID: dftQRUZRipclawOracle, Power: 4, Toughness: 3})
	if !g.Battlefield.Contains(wrangler) {
		t.Fatal("the Wrangler did not enter")
	}
	for _, p := range g.Seats {
		owed := effectstest.DiscardOwed(g, p.ID)
		if p.ID == me.ID && owed != 0 {
			t.Errorf("you owe %d discards, want none", owed)
		}
		if p.ID != me.ID && owed != 1 {
			t.Errorf("an opponent owes %d discards, want 1", owed)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, wrangler); c.IsCreature() {
		t.Error("the uncrewed Wrangler is a creature")
	}
}

func TestDftQRUZRipclawWranglerCrewsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wrangler := effectstest.DftABPush(g, me.ID, "Ripclaw Wrangler", effectstest.DftABVehicle, dftQRUZRipclawOracle, 4, 3)
	one := effectstest.B12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	two := effectstest.B12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, wrangler, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{one}}); err == nil {
		t.Fatal("crewed with total power 1; crew 2 wants 2")
	}
	effectstest.DftABCrew(t, g, me, wrangler, 0, two)
}

// --- Road Rage --------------------------------------------------------------------------

func TestDftQRUZRoadRageDealsTwoPlusYourMountsAndVehicles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.DftABPush(g, me.ID, "Ripclaw Wrangler", effectstest.DftABVehicle, dftQRUZRipclawOracle, 4, 3)
	effectstest.B12Creature(g, me.ID, "Steed", "Creature — Horse Mount", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DftABPush(g, opp.ID, "Their Wrangler", effectstest.DftABVehicle, dftQRUZRipclawOracle, 4, 3)
	giant := effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 7, 7)
	effectstest.CastCatalogSpell(t, g, "Road Rage", "Instant", dftQRUZRoadRageOracle, effectstest.CardRefs(giant))
	effectstest.PassPriorityAroundTable(t, g)
	if got := dftQRUZDamageOn(t, g, giant); got != 4 {
		t.Errorf("Road Rage dealt %d, want 4 (2 plus a Vehicle and a Mount you control)", got)
	}
}

// --- Roadside Blowout -----------------------------------------------------------------

func TestDftQRUZRoadsideBlowoutCostsTwoLessForManaValueOneAndBouncesAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	one := dftSTCostly(g, opp.ID, "Elf", "Creature — Elf", "{G}")
	two := dftSTCostly(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	mine := dftSTCostly(g, me.ID, "My Elf", "Creature — Elf", "{G}")
	drawn := effectstest.SeedLibrary(me, "Drawn")[0]
	effectstest.AdvanceToMain(t, g)
	card := game.Card{Name: "Roadside Blowout", TypeLine: "Sorcery", ManaCost: "{2}{U}", OracleID: dftQRUZBlowoutOracle}

	if err := effectstest.SocB19CastWithPool(g, me, card, "U", 1, game.CastSpellParams{Targets: effectstest.CardRefs(two)}); err == nil {
		t.Fatal("cast for {U} against a mana value 2 creature")
	}
	if err := effectstest.SocB19CastWithPool(g, me, card, "U", 3, game.CastSpellParams{Targets: effectstest.CardRefs(mine)}); err == nil {
		t.Fatal("targeted a creature you control")
	}
	hand := me.Hand.Size()
	if err := effectstest.SocB19CastWithPool(g, me, card, "U", 1, game.CastSpellParams{Targets: effectstest.CardRefs(one)}); err != nil {
		t.Fatalf("cast for {U} against a mana value 1 creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(one) {
		t.Error("the creature did not return to its owner's hand")
	}
	// The Blowout goes into the hand and out again as it is cast; the draw adds one.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d (one card drawn)", got, hand+1)
	}
	if !me.Hand.Contains(drawn) {
		t.Error("the top card of the library was not drawn")
	}
}

// --- Rocketeer Boostbuggy -------------------------------------------------------------

func TestDftQRUZRocketeerBoostbuggyMakesATreasureWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	buggy := effectstest.DftABPush(g, me.ID, "Rocketeer Boostbuggy", effectstest.DftABVehicle, dftQRUZBoostbuggyOracle, 3, 2)
	pilot := effectstest.B12Creature(g, me.ID, "Pilot", "Creature — Goblin Pilot", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.DftABCrew(t, g, me, buggy, 1, pilot)
	effectstest.DeclareAttack(t, g, opp.ID, buggy)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures after the Boostbuggy attacked, want 1", n)
	}
}

func TestDftQRUZRocketeerBoostbuggyExhaustsIntoAPermanentCreatureWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	buggy := effectstest.DftABPush(g, me.ID, "Rocketeer Boostbuggy", effectstest.DftABVehicle, dftQRUZBoostbuggyOracle, 3, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, buggy, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, buggy)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Fatalf("the Boostbuggy is not an artifact creature: %q", c.TypeLine)
	}
	if got := effectstest.CountersOn(g, buggy, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, buggy); got != 4 {
		t.Errorf("power with its counter %d, want 4", got)
	}
	effectstest.DftMPPassTurn(t, g)
	if c, _ := effectstest.BattlefieldCard(g, buggy); !c.IsCreature() {
		t.Error("the exhausted Boostbuggy stopped being a creature when the turn ended")
	}
	effectstest.FillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, buggy, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the exhaust ability was activated twice")
	}
}

// --- Run Over -----------------------------------------------------------------------------

func TestDftQRUZRunOverCostsOneLessWithYourMountAndBites(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	steed := effectstest.B12Creature(g, me.ID, "Steed", "Creature — Horse Mount", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 5, 5)
	effectstest.AdvanceToMain(t, g)
	card := game.Card{Name: "Run Over", TypeLine: "Instant", ManaCost: "{1}{G}", OracleID: dftQRUZRunOverOracle}

	if err := effectstest.SocB19CastWithPool(g, me, card, "G", 1, game.CastSpellParams{Targets: effectstest.CardRefs(bear, victim)}); err == nil {
		t.Fatal("cast for {G} with a creature that is not a Mount or Vehicle")
	}
	if err := effectstest.SocB19CastWithPool(g, me, card, "G", 1, game.CastSpellParams{Targets: effectstest.CardRefs(steed, victim)}); err != nil {
		t.Fatalf("cast for {G} with your Mount: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := dftQRUZDamageOn(t, g, victim); got != 3 {
		t.Errorf("the Ox took %d damage, want the Mount's power 3", got)
	}
	if got := dftQRUZDamageOn(t, g, steed); got != 0 {
		t.Errorf("the Mount took %d damage; Run Over is one-sided", got)
	}
}

func TestDftQRUZRunOverNeedsAnOpponentsCreatureForTheSecondTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Elk", "Creature — Elk", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Run Over", "Instant", dftQRUZRunOverOracle, effectstest.CardRefs(bear, other)); err == nil {
		t.Error("Run Over targeted your own creature as the one taking the damage")
	}
}

// --- Veloheart Bike -----------------------------------------------------------------------

func TestDftQRUZVeloheartBikeGainsTwoAndTapsForAnyColourTheTurnItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	life := effectstest.LifeOf(g, me.ID)
	bike := effectstest.DftABCast(t, g, me, game.Card{Name: "Veloheart Bike", TypeLine: effectstest.DftABVehicle,
		OracleID: dftQRUZVeloheartOracle, Power: 4, Toughness: 2})
	if got := effectstest.LifeOf(g, me.ID); got != life+2 {
		t.Errorf("life %d, want %d", got, life+2)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.B28TapForMana(t, g, me.ID, bike, "U")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("the uncrewed Bike made %v, want {U}", got)
	}
}

func TestDftQRUZVeloheartBikeCrewsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bike := effectstest.DftABPush(g, me.ID, "Veloheart Bike", effectstest.DftABVehicle, dftQRUZVeloheartOracle, 4, 2)
	one := effectstest.B12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	two := effectstest.B12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, bike, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{one}}); err == nil {
		t.Fatal("crewed with total power 1; crew 2 wants 2")
	}
	effectstest.DftABCrew(t, g, me, bike, 0, two)
}

// --- Veteran Beastrider -------------------------------------------------------------------

func TestDftQRUZVeteranBeastriderUntapsYourCreaturesAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.DftABPush(g, me.ID, "Veteran Beastrider", "Creature — Human Knight", dftQRUZBeastriderOracle, 3, 4)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		effectstest.B16Tap(g, mine)
		effectstest.B16Tap(g, theirs)
	})
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if dftQRUZTapped(t, g, mine) {
		t.Error("your creature is still tapped after your end step trigger")
	}
	if !dftQRUZTapped(t, g, theirs) {
		t.Error("an opponent's creature was untapped")
	}
}

func TestDftQRUZVeteranBeastriderPumpsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rider := effectstest.DftABPush(g, me.ID, "Veteran Beastrider", "Creature — Human Knight", dftQRUZBeastriderOracle, 3, 4)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPool(me, 2)
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.B16Activate(t, g, me.ID, rider, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 3 || tg != 3 {
		t.Errorf("your Bear is %d/%d, want 3/3", p, tg)
	}
	if p := effectstest.EffectivePower(t, g, rider); p != 4 {
		t.Errorf("the Beastrider has power %d, want 4", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
	effectstest.DftMPPassTurn(t, g)
	if p := effectstest.EffectivePower(t, g, mine); p != 2 {
		t.Errorf("the pump outlived the turn: power %d", p)
	}
}

// --- Voyager Quickwelder ------------------------------------------------------------------

func TestDftQRUZVoyagerQuickwelderDiscountsYourArtifactSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Voyager Quickwelder", "Artifact Creature — Robot Artificer", dftQRUZQuickwelderOracle, false)
	if got := effectstest.PriceInHand(t, g, me, "My Chalice", "Artifact", "{2}"); got != 1 {
		t.Errorf("your artifact spell costs %d, want 1", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Their Chalice", "Artifact", "{2}"); got != 2 {
		t.Errorf("an opponent's artifact spell costs %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "My Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("your creature spell costs %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "My Thopter", "Artifact Creature — Thopter", "{W}"); got != 1 {
		t.Errorf("a {W} artifact creature costs %d, want 1 (coloured mana is never reduced)", got)
	}
}

// --- Willowrush Verge ---------------------------------------------------------------------

func TestDftQRUZWillowrushVergeNeedsAForestOrAnIslandForGreen(t *testing.T) {
	for _, typeLine := range []string{"Basic Land — Forest", "Basic Land — Island", "Land — Forest Island"} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		land := effectstest.SeedLand(g, me.ID, "Willowrush Verge", "Land", dftQRUZWillowrushOracle)
		effectstest.SeedLand(g, me.ID, "Plains", "Basic Land — Plains", "")
		if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err == nil {
			t.Fatalf("%s: {G} with only a Plains", typeLine)
		}
		effectstest.B28TapForMana(t, g, me.ID, land, "")
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "U" {
			t.Fatalf("%s: the blue ability gave %v, want {U}", typeLine, got)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		me.ManaPool = nil

		effectstest.SeedLand(g, me.ID, "Other", typeLine, "")
		effectstest.B40TapForMana(t, g, me.ID, land, 1, "")
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
			t.Errorf("%s: pool %v, want {G}", typeLine, got)
		}
	}
}

// --- Wreckage Wickerfolk ------------------------------------------------------------------

func TestDftQRUZWreckageWickerfolkFliesAndSurveilsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Bin Me", "Keep Me", "Under")
	effectstest.AdvanceToMain(t, g)
	gy := me.Graveyard.Size()
	folk := effectstest.DftABCast(t, g, me, game.Card{Name: "Wreckage Wickerfolk", TypeLine: "Artifact Creature — Scarecrow",
		OracleID: dftQRUZWickerfolkOracle, Power: 1, Toughness: 3})
	if !effectstest.HasEffectiveKeyword(t, g, folk, "flying") {
		t.Error("the Wickerfolk has no flying")
	}
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil when the Wickerfolk entered")
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("surveil looked at %d cards, want 2", len(c.ScryCards))
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards[1:], c.ScryCards[:1]); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if got := me.Graveyard.Size(); got != gy+1 {
		t.Errorf("graveyard %d, want %d (one card surveilled away)", got, gy+1)
	}
}

// --- Wreck Remover ------------------------------------------------------------------------

func TestDftQRUZWreckRemoverExilesAGraveyardCardAndGainsALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.DftABInGraveyard(opp, "Their Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	life := effectstest.LifeOf(g, me.ID)
	effectstest.DftABCast(t, g, me, game.Card{Name: "Wreck Remover", TypeLine: "Artifact Creature — Construct",
		OracleID: dftQRUZWreckRemoverOracle, Power: 3, Toughness: 4})
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) {
		t.Error("the graveyard card was not exiled")
	}
	if got := effectstest.LifeOf(g, me.ID); got != life+1 {
		t.Errorf("life %d, want %d", got, life+1)
	}
}

func TestDftQRUZWreckRemoverGainsALifeWithNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.DftABInGraveyard(opp, "Their Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	life := effectstest.LifeOf(g, me.ID)
	effectstest.DftABCast(t, g, me, game.Card{Name: "Wreck Remover", TypeLine: "Artifact Creature — Construct",
		OracleID: dftQRUZWreckRemoverOracle, Power: 3, Toughness: 4})
	effectstest.B26DeclinePickTarget(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(theirs) {
		t.Error("a card was exiled with no target chosen")
	}
	if got := effectstest.LifeOf(g, me.ID); got != life+1 {
		t.Errorf("life %d, want %d", got, life+1)
	}
}

func TestDftQRUZWreckRemoverExilesWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	remover := effectstest.DftABPush(g, me.ID, "Wreck Remover", "Artifact Creature — Construct", dftQRUZWreckRemoverOracle, 3, 4)
	theirs := effectstest.DftABInGraveyard(opp, "Their Bear", "Creature — Bear")
	life := effectstest.LifeOf(g, me.ID)
	effectstest.DeclareAttack(t, g, opp.ID, remover)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) {
		t.Error("the attack trigger did not exile the graveyard card")
	}
	if got := effectstest.LifeOf(g, me.ID); got != life+1 {
		t.Errorf("life %d, want %d", got, life+1)
	}
}

func TestDftQRUZWreckRemoverCycles(t *testing.T) {
	effectstest.DftSTCycles(t, "Wreck Remover", "Artifact Creature — Construct", dftQRUZWreckRemoverOracle, 0)
}

// --- Wretched Doll ------------------------------------------------------------------------

func TestDftQRUZWretchedDollSurveilsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Top", "Next")
	doll := effectstest.DftABPush(g, me.ID, "Wretched Doll", "Artifact Creature — Toy", dftQRUZWretchedDollOracle, 3, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "B", 1)
	effectstest.B16Activate(t, g, me.ID, doll, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the Doll queued no surveil")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("surveil looked at %d cards, want 1", len(c.ScryCards))
	}
	if !dftQRUZTapped(t, g, doll) {
		t.Error("the Doll did not tap")
	}
}

func TestDftQRUZWretchedDollCannotTapTheTurnItArrives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Top")
	doll := effectstest.PushCatalogPermanent(g, me.ID, "Wretched Doll", "Artifact Creature — Toy", dftQRUZWretchedDollOracle, true)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, doll, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a summoning-sick Doll tapped to surveil")
	}
}
