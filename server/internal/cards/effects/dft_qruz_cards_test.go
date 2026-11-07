package effects

import (
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- fixtures ------------------------------------------------------------------

// dftQRUZDamageOn is the damage marked on a permanent.
func dftQRUZDamageOn(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.DamageMarked
}

// dftQRUZTapped reports whether a permanent is tapped.
func dftQRUZTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// --- Rangers' Aetherhive ------------------------------------------------------------

func TestDftQRUZRangersAetherhiveMakesAThopterOnYourExhaustActivationOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hive := dftABPush(g, me.ID, "Rangers' Aetherhive", dftABVehicle, dftQRUZAetherhiveOracle, 3, 5)
	buggy := dftABPush(g, me.ID, "Rocketeer Boostbuggy", dftABVehicle, dftQRUZBoostbuggyOracle, 3, 2)
	pilot := b12Creature(g, me.ID, "Pilot", "Creature — Goblin Pilot", 1, 1)
	advanceToMain(t, g)

	// Crewing is not an exhaust ability.
	dftABCrew(t, g, me, hive, 0, pilot)
	if n, _ := countTokensNamed(g, me.ID, "Thopter"); n != 0 {
		t.Fatalf("%d Thopters after crewing; only an exhaust activation makes one", n)
	}
	if !hasEffectiveKeyword(t, g, hive, "vigilance") {
		t.Error("the crewed Aetherhive has no vigilance")
	}

	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, buggy, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the Boostbuggy's exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Thopter"); n != 1 {
		t.Fatalf("%d Thopters after an exhaust activation, want 1", n)
	}
	if !hasEffectiveKeyword(t, g, dftQRUZTokenNamed(t, g, me.ID, "Thopter"), "flying") {
		t.Error("the Thopter has no flying")
	}

	// An opponent's exhaust activation is not "you activate".
	theirs := dftABPush(g, opp.ID, "Rocketeer Boostbuggy", dftABVehicle, dftQRUZBoostbuggyOracle, 3, 2)
	fillPool(opp, 3)
	if err := g.ActivateCatalogAbility(opp.ID, theirs, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the opponent's exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Thopter"); n != 1 {
		t.Errorf("%d Thopters after an opponent's exhaust activation, want still 1", n)
	}
}

// dftQRUZTokenNamed is the id of a token with that name under controller.
func dftQRUZTokenNamed(t *testing.T, g *game.Game, controller uuid.UUID, name string) uuid.UUID {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && IsToken(c) && c.Name == name {
			return c.InstanceID
		}
	}
	t.Fatalf("no %s token", name)
	return uuid.Nil
}

// --- Regal Imperiosaur --------------------------------------------------------------

func TestDftQRUZRegalImperiosaurPumpsYourOtherDinosaursOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	imperiosaur := dftABPush(g, me.ID, "Regal Imperiosaur", "Creature — Dinosaur", dftQRUZImperiosaurOracle, 5, 4)
	dino := b12Creature(g, me.ID, "Raptor", "Creature — Dinosaur", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Raptor", "Creature — Dinosaur", 2, 2)
	if p, tg := effectivePower(t, g, dino), effectiveToughness(t, g, dino); p != 3 || tg != 3 {
		t.Errorf("your other Dinosaur is %d/%d, want 3/3", p, tg)
	}
	if p, tg := effectivePower(t, g, imperiosaur), effectiveToughness(t, g, imperiosaur); p != 5 || tg != 4 {
		t.Errorf("the Imperiosaur is %d/%d, want its printed 5/4", p, tg)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Dinosaur has power %d, want 2", p)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's Dinosaur has power %d, want 2", p)
	}
}

// --- Ride's End -----------------------------------------------------------------------

func TestDftQRUZRidesEndCostsThreeLessAgainstATappedPermanentAndExiles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	untapped := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	vehicle := dftABPush(g, opp.ID, "Their Wrangler", dftABVehicle, dftQRUZRipclawOracle, 4, 3)
	g.WithWriteLock(func() { b16Tap(g, vehicle) })
	advanceToMain(t, g)
	card := game.Card{Name: "Ride's End", TypeLine: "Instant", ManaCost: "{4}{W}", OracleID: dftQRUZRidesEndOracle}

	if err := socB19CastWithPool(g, me, card, "W", 2, game.CastSpellParams{Targets: cardRefs(untapped)}); err == nil {
		t.Fatal("cast for two mana against an untapped creature")
	}
	if err := socB19CastWithPool(g, me, card, "W", 2, game.CastSpellParams{Targets: cardRefs(vehicle)}); err != nil {
		t.Fatalf("cast for {1}{W} against a tapped, uncrewed Vehicle: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(vehicle) {
		t.Error("the Vehicle was not exiled")
	}
}

// --- Ripclaw Wrangler -----------------------------------------------------------------

func TestDftQRUZRipclawWranglerMakesEachOpponentDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, p := range g.Seats {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	}
	advanceToMain(t, g)
	wrangler := dftABCast(t, g, me, game.Card{Name: "Ripclaw Wrangler", TypeLine: dftABVehicle,
		OracleID: dftQRUZRipclawOracle, Power: 4, Toughness: 3})
	if !g.Battlefield.Contains(wrangler) {
		t.Fatal("the Wrangler did not enter")
	}
	for _, p := range g.Seats {
		owed := discardOwed(g, p.ID)
		if p.ID == me.ID && owed != 0 {
			t.Errorf("you owe %d discards, want none", owed)
		}
		if p.ID != me.ID && owed != 1 {
			t.Errorf("an opponent owes %d discards, want 1", owed)
		}
	}
	if c, _ := battlefieldCard(g, wrangler); c.IsCreature() {
		t.Error("the uncrewed Wrangler is a creature")
	}
}

func TestDftQRUZRipclawWranglerCrewsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wrangler := dftABPush(g, me.ID, "Ripclaw Wrangler", dftABVehicle, dftQRUZRipclawOracle, 4, 3)
	one := b12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	two := b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, wrangler, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{one}}); err == nil {
		t.Fatal("crewed with total power 1; crew 2 wants 2")
	}
	dftABCrew(t, g, me, wrangler, 0, two)
}

// --- Road Rage --------------------------------------------------------------------------

func TestDftQRUZRoadRageDealsTwoPlusYourMountsAndVehicles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dftABPush(g, me.ID, "Ripclaw Wrangler", dftABVehicle, dftQRUZRipclawOracle, 4, 3)
	b12Creature(g, me.ID, "Steed", "Creature — Horse Mount", 2, 2)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dftABPush(g, opp.ID, "Their Wrangler", dftABVehicle, dftQRUZRipclawOracle, 4, 3)
	giant := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 7, 7)
	castCatalogSpell(t, g, "Road Rage", "Instant", dftQRUZRoadRageOracle, cardRefs(giant))
	passPriorityAroundTable(t, g)
	if got := dftQRUZDamageOn(t, g, giant); got != 4 {
		t.Errorf("Road Rage dealt %d, want 4 (2 plus a Vehicle and a Mount you control)", got)
	}
}

// --- Roadside Blowout -----------------------------------------------------------------

func TestDftQRUZRoadsideBlowoutCostsTwoLessForManaValueOneAndBouncesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	one := dftSTCostly(g, opp.ID, "Elf", "Creature — Elf", "{G}")
	two := dftSTCostly(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	mine := dftSTCostly(g, me.ID, "My Elf", "Creature — Elf", "{G}")
	drawn := seedLibrary(me, "Drawn")[0]
	advanceToMain(t, g)
	card := game.Card{Name: "Roadside Blowout", TypeLine: "Sorcery", ManaCost: "{2}{U}", OracleID: dftQRUZBlowoutOracle}

	if err := socB19CastWithPool(g, me, card, "U", 1, game.CastSpellParams{Targets: cardRefs(two)}); err == nil {
		t.Fatal("cast for {U} against a mana value 2 creature")
	}
	if err := socB19CastWithPool(g, me, card, "U", 3, game.CastSpellParams{Targets: cardRefs(mine)}); err == nil {
		t.Fatal("targeted a creature you control")
	}
	hand := me.Hand.Size()
	if err := socB19CastWithPool(g, me, card, "U", 1, game.CastSpellParams{Targets: cardRefs(one)}); err != nil {
		t.Fatalf("cast for {U} against a mana value 1 creature: %v", err)
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	buggy := dftABPush(g, me.ID, "Rocketeer Boostbuggy", dftABVehicle, dftQRUZBoostbuggyOracle, 3, 2)
	pilot := b12Creature(g, me.ID, "Pilot", "Creature — Goblin Pilot", 1, 1)
	advanceToMain(t, g)
	dftABCrew(t, g, me, buggy, 1, pilot)
	declareAttack(t, g, opp.ID, buggy)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures after the Boostbuggy attacked, want 1", n)
	}
}

func TestDftQRUZRocketeerBoostbuggyExhaustsIntoAPermanentCreatureWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	buggy := dftABPush(g, me.ID, "Rocketeer Boostbuggy", dftABVehicle, dftQRUZBoostbuggyOracle, 3, 2)
	advanceToMain(t, g)
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, buggy, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, buggy)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Fatalf("the Boostbuggy is not an artifact creature: %q", c.TypeLine)
	}
	if got := countersOn(g, buggy, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if got := effectivePowerWithCounters(t, g, buggy); got != 4 {
		t.Errorf("power with its counter %d, want 4", got)
	}
	dftMPPassTurn(t, g)
	if c, _ := battlefieldCard(g, buggy); !c.IsCreature() {
		t.Error("the exhausted Boostbuggy stopped being a creature when the turn ended")
	}
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, buggy, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the exhaust ability was activated twice")
	}
}

// --- Run Over -----------------------------------------------------------------------------

func TestDftQRUZRunOverCostsOneLessWithYourMountAndBites(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	steed := b12Creature(g, me.ID, "Steed", "Creature — Horse Mount", 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 5, 5)
	advanceToMain(t, g)
	card := game.Card{Name: "Run Over", TypeLine: "Instant", ManaCost: "{1}{G}", OracleID: dftQRUZRunOverOracle}

	if err := socB19CastWithPool(g, me, card, "G", 1, game.CastSpellParams{Targets: cardRefs(bear, victim)}); err == nil {
		t.Fatal("cast for {G} with a creature that is not a Mount or Vehicle")
	}
	if err := socB19CastWithPool(g, me, card, "G", 1, game.CastSpellParams{Targets: cardRefs(steed, victim)}); err != nil {
		t.Fatalf("cast for {G} with your Mount: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := dftQRUZDamageOn(t, g, victim); got != 3 {
		t.Errorf("the Ox took %d damage, want the Mount's power 3", got)
	}
	if got := dftQRUZDamageOn(t, g, steed); got != 0 {
		t.Errorf("the Mount took %d damage; Run Over is one-sided", got)
	}
}

func TestDftQRUZRunOverNeedsAnOpponentsCreatureForTheSecondTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Elk", "Creature — Elk", 2, 2)
	if err := castCatalogSpellErr(t, g, "Run Over", "Instant", dftQRUZRunOverOracle, cardRefs(bear, other)); err == nil {
		t.Error("Run Over targeted your own creature as the one taking the damage")
	}
}

// --- Veloheart Bike -----------------------------------------------------------------------

func TestDftQRUZVeloheartBikeGainsTwoAndTapsForAnyColourTheTurnItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	life := lifeOf(g, me.ID)
	bike := dftABCast(t, g, me, game.Card{Name: "Veloheart Bike", TypeLine: dftABVehicle,
		OracleID: dftQRUZVeloheartOracle, Power: 4, Toughness: 2})
	if got := lifeOf(g, me.ID); got != life+2 {
		t.Errorf("life %d, want %d", got, life+2)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	b28TapForMana(t, g, me.ID, bike, "U")
	if got := poolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("the uncrewed Bike made %v, want {U}", got)
	}
}

func TestDftQRUZVeloheartBikeCrewsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bike := dftABPush(g, me.ID, "Veloheart Bike", dftABVehicle, dftQRUZVeloheartOracle, 4, 2)
	one := b12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	two := b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, bike, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{one}}); err == nil {
		t.Fatal("crewed with total power 1; crew 2 wants 2")
	}
	dftABCrew(t, g, me, bike, 0, two)
}

// --- Veteran Beastrider -------------------------------------------------------------------

func TestDftQRUZVeteranBeastriderUntapsYourCreaturesAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dftABPush(g, me.ID, "Veteran Beastrider", "Creature — Human Knight", dftQRUZBeastriderOracle, 3, 4)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		b16Tap(g, mine)
		b16Tap(g, theirs)
	})
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if dftQRUZTapped(t, g, mine) {
		t.Error("your creature is still tapped after your end step trigger")
	}
	if !dftQRUZTapped(t, g, theirs) {
		t.Error("an opponent's creature was untapped")
	}
}

func TestDftQRUZVeteranBeastriderPumpsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rider := dftABPush(g, me.ID, "Veteran Beastrider", "Creature — Human Knight", dftQRUZBeastriderOracle, 3, 4)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	fillPool(me, 2)
	fillPoolColored(me, "G", 1)
	fillPoolColored(me, "W", 1)
	b16Activate(t, g, me.ID, rider, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 3 || tg != 3 {
		t.Errorf("your Bear is %d/%d, want 3/3", p, tg)
	}
	if p := effectivePower(t, g, rider); p != 4 {
		t.Errorf("the Beastrider has power %d, want 4", p)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
	dftMPPassTurn(t, g)
	if p := effectivePower(t, g, mine); p != 2 {
		t.Errorf("the pump outlived the turn: power %d", p)
	}
}

// --- Voyager Quickwelder ------------------------------------------------------------------

func TestDftQRUZVoyagerQuickwelderDiscountsYourArtifactSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Voyager Quickwelder", "Artifact Creature — Robot Artificer", dftQRUZQuickwelderOracle, false)
	if got := priceInHand(t, g, me, "My Chalice", "Artifact", "{2}"); got != 1 {
		t.Errorf("your artifact spell costs %d, want 1", got)
	}
	if got := priceInHand(t, g, them, "Their Chalice", "Artifact", "{2}"); got != 2 {
		t.Errorf("an opponent's artifact spell costs %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "My Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("your creature spell costs %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "My Thopter", "Artifact Creature — Thopter", "{W}"); got != 1 {
		t.Errorf("a {W} artifact creature costs %d, want 1 (coloured mana is never reduced)", got)
	}
}

// --- Willowrush Verge ---------------------------------------------------------------------

func TestDftQRUZWillowrushVergeNeedsAForestOrAnIslandForGreen(t *testing.T) {
	for _, typeLine := range []string{"Basic Land — Forest", "Basic Land — Island", "Land — Forest Island"} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		land := seedLand(g, me.ID, "Willowrush Verge", "Land", dftQRUZWillowrushOracle)
		seedLand(g, me.ID, "Plains", "Basic Land — Plains", "")
		if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err == nil {
			t.Fatalf("%s: {G} with only a Plains", typeLine)
		}
		b28TapForMana(t, g, me.ID, land, "")
		if got := poolColors(me); len(got) != 1 || got[0] != "U" {
			t.Fatalf("%s: the blue ability gave %v, want {U}", typeLine, got)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		me.ManaPool = nil

		seedLand(g, me.ID, "Other", typeLine, "")
		b40TapForMana(t, g, me.ID, land, 1, "")
		if got := poolColors(me); len(got) != 1 || got[0] != "G" {
			t.Errorf("%s: pool %v, want {G}", typeLine, got)
		}
	}
}

// --- Wreckage Wickerfolk ------------------------------------------------------------------

func TestDftQRUZWreckageWickerfolkFliesAndSurveilsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Bin Me", "Keep Me", "Under")
	advanceToMain(t, g)
	gy := me.Graveyard.Size()
	folk := dftABCast(t, g, me, game.Card{Name: "Wreckage Wickerfolk", TypeLine: "Artifact Creature — Scarecrow",
		OracleID: dftQRUZWickerfolkOracle, Power: 1, Toughness: 3})
	if !hasEffectiveKeyword(t, g, folk, "flying") {
		t.Error("the Wickerfolk has no flying")
	}
	c := surveilChoiceFor(g, me.ID)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := dftABInGraveyard(opp, "Their Bear", "Creature — Bear")
	advanceToMain(t, g)
	life := lifeOf(g, me.ID)
	dftABCast(t, g, me, game.Card{Name: "Wreck Remover", TypeLine: "Artifact Creature — Construct",
		OracleID: dftQRUZWreckRemoverOracle, Power: 3, Toughness: 4})
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) {
		t.Error("the graveyard card was not exiled")
	}
	if got := lifeOf(g, me.ID); got != life+1 {
		t.Errorf("life %d, want %d", got, life+1)
	}
}

func TestDftQRUZWreckRemoverGainsALifeWithNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := dftABInGraveyard(opp, "Their Bear", "Creature — Bear")
	advanceToMain(t, g)
	life := lifeOf(g, me.ID)
	dftABCast(t, g, me, game.Card{Name: "Wreck Remover", TypeLine: "Artifact Creature — Construct",
		OracleID: dftQRUZWreckRemoverOracle, Power: 3, Toughness: 4})
	b26DeclinePickTarget(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(theirs) {
		t.Error("a card was exiled with no target chosen")
	}
	if got := lifeOf(g, me.ID); got != life+1 {
		t.Errorf("life %d, want %d", got, life+1)
	}
}

func TestDftQRUZWreckRemoverExilesWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	remover := dftABPush(g, me.ID, "Wreck Remover", "Artifact Creature — Construct", dftQRUZWreckRemoverOracle, 3, 4)
	theirs := dftABInGraveyard(opp, "Their Bear", "Creature — Bear")
	life := lifeOf(g, me.ID)
	declareAttack(t, g, opp.ID, remover)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) {
		t.Error("the attack trigger did not exile the graveyard card")
	}
	if got := lifeOf(g, me.ID); got != life+1 {
		t.Errorf("life %d, want %d", got, life+1)
	}
}

func TestDftQRUZWreckRemoverCycles(t *testing.T) {
	dftSTCycles(t, "Wreck Remover", "Artifact Creature — Construct", dftQRUZWreckRemoverOracle, 0)
}

// --- Wretched Doll ------------------------------------------------------------------------

func TestDftQRUZWretchedDollSurveilsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top", "Next")
	doll := dftABPush(g, me.ID, "Wretched Doll", "Artifact Creature — Toy", dftQRUZWretchedDollOracle, 3, 1)
	advanceToMain(t, g)
	fillPoolColored(me, "B", 1)
	b16Activate(t, g, me.ID, doll, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top")
	doll := pushCatalogPermanent(g, me.ID, "Wretched Doll", "Artifact Creature — Toy", dftQRUZWretchedDollOracle, true)
	advanceToMain(t, g)
	fillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, doll, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a summoning-sick Doll tapped to surveil")
	}
}
