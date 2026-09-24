package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// slice_295b_mana_rocks_test.go — slice 295-b (mana rocks), relates
// to #295.

const (
	prismaticLensOracle       = "e3056f28-868b-401a-a528-7528d639bdeb"
	astralCornucopiaOracle    = "1bc42024-52da-4d93-8b47-544f0a4a72a1"
	cursedMirrorOracle        = "4d67e2a7-4aa7-44cc-853b-500d7aac046d"
	liquimetalTorqueOracle    = "b7d4b7dd-fbb1-4ca3-875f-ef13a95e66ad"
	midnightClockOracle       = "c68faebc-b2cd-461b-b93e-e1fcd4816810"
	cagedSunOracle            = "09b895ff-e729-48d1-bfc1-ea5fd7adda6a"
	herdHeirloomOracle        = "78b6dd40-c182-4037-a4d3-9fd012b2c584"
	chromaticOrreryOracle     = "95c3976c-33f3-490b-bfd3-7f1af2fe0416"
	covetedJewelOracle        = "98492d7d-3b9e-4ae1-ac45-1b508d6d2670"
	theEternityElevatorOracle = "11323af4-b8b8-4ca9-932f-377c7fd77dea"
)

func TestB295PrismaticLensTapsForColorlessAndPaidAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lens := pushCatalogPermanent(g, me.ID, "Prismatic Lens", "Artifact", prismaticLensOracle, false)

	if err := g.ActivateManaAbility(me.ID, lens, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility 0: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"C"}) {
		t.Errorf("pool after ability 0 = %v, want [C]", got)
	}
	me.ManaPool.EmptyPool()
	untap(g, lens)

	// The second ability needs a {1} to pay before it taps.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, lens, 1, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("ActivateManaAbility 1: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"G"}) {
		t.Errorf("pool after ability 1 = %v, want [G] (the {1} was spent, not produced)", got)
	}
}

func TestB295AstralCornucopiaEntersWithXCountersAndTapsForThatManyOfOneColor(t *testing.T) {
	g := newCatalogGame(t)
	rock := b12PlayFromHand(t, g, "Astral Cornucopia", "Artifact", astralCornucopiaOracle, game.CastSpellParams{XValue: 3})
	passPriorityAroundTable(t, g)
	if got := counterOn(g, rock, game.CounterCharge); got != 3 {
		t.Fatalf("charge counters = %d, want 3", got)
	}
	me := g.Seats[0]
	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"U", "U", "U"}) {
		t.Errorf("pool = %v, want three blue mana (one per charge counter)", got)
	}
}

func TestB295CursedMirrorTapsForRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mirror := pushCatalogPermanent(g, me.ID, "Cursed Mirror", "Artifact", cursedMirrorOracle, false)
	if err := g.ActivateManaAbility(me.ID, mirror, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool = %v, want [R]", got)
	}
	spec, ok := Lookup(cursedMirrorOracle)
	if !ok {
		t.Fatal("Cursed Mirror is registered")
	}
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) == 0 {
		t.Error("Cursed Mirror should declare a caveat for the unbuilt copy-until-end-of-turn ETB")
	}
}

func TestB295LiquimetalTorqueMakesATargetAnArtifactUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	torque := pushCatalogPermanent(g, me, "Liquimetal Torque", "Artifact", liquimetalTorqueOracle, false)
	bear := pushVanillaCreature(g, them, "Grizzly Bears", 2, 2)

	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(me, torque, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)

	types := effectiveTypes(t, g, bear)
	if !hasAbility(types, "Artifact") {
		t.Errorf("types = %v, want Artifact added", types)
	}
	if !hasAbility(types, "Creature") {
		t.Errorf("types = %v, the printed type should still be there too", types)
	}
}

func TestB295MidnightClockTapsForBlueAndAccumulatesHourCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	clock := pushCatalogPermanent(g, me.ID, "Midnight Clock", "Artifact", midnightClockOracle, false)

	if err := g.ActivateManaAbility(me.ID, clock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("pool = %v, want [U]", got)
	}

	// The upkeep trigger fires for ANY player's upkeep.
	before := counterOn(g, clock, "hour")
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, clock, "hour"); got != before+1 {
		t.Errorf("hour counters = %d, want %d after an upkeep passes", got, before+1)
	}

	spec, ok := Lookup(midnightClockOracle)
	if !ok {
		t.Fatal("Midnight Clock is registered")
	}
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) == 0 {
		t.Error("Midnight Clock should declare a caveat for the unbuilt twelfth-counter payoff")
	}
}

func TestB295CagedSunPumpsCreaturesOfTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	redMine := b31Push(g, me, "Goblin Guide", "Creature — Goblin Scout", "", "{R}", 2, 2, "R")
	greenMine := b31Push(g, me, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	redTheirs := b31Push(g, them, "Raging Goblin", "Creature — Goblin Berserker", "", "{R}", 1, 1, "R")

	pushChosenColorPermanent(t, g, me, "Caged Sun", "Artifact", cagedSunOracle, "R")

	if got := effectivePower(t, g, redMine); got != 3 {
		t.Errorf("redMine power = %d, want 3 (+1/+1)", got)
	}
	if got := effectivePower(t, g, greenMine); got != 2 {
		t.Errorf("greenMine power = %d, want 2 (unaffected)", got)
	}
	if got := effectivePower(t, g, redTheirs); got != 1 {
		t.Errorf("redTheirs power = %d, want 1 — the anthem has no controller clause, only colour", got)
	}

	spec, ok := Lookup(cagedSunOracle)
	if !ok {
		t.Fatal("Caged Sun is registered")
	}
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) == 0 {
		t.Error("Caged Sun should declare a caveat for the unbuilt land-mana-doubling clause")
	}
}

func TestB295HerdHeirloomManaIsRestrictedToCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := seedPermanentWithOracle(g, me.ID, "Herd Heirloom", "Artifact", herdHeirloomOracle)

	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 1 {
		t.Fatalf("pool = %v, want one mana", me.ManaPool)
	}

	cost, _ := game.ParseCost("{G}")
	creatureSpell := game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"}
	noncreatureSpell := game.Card{Name: "Giant Growth", TypeLine: "Instant", ManaCost: "{G}"}

	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(creatureSpell)) {
		t.Error("Herd Heirloom mana refused a creature spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(noncreatureSpell)) {
		t.Error("Herd Heirloom mana paid for a noncreature spell — stronger than printed")
	}

	spec, ok := Lookup(herdHeirloomOracle)
	if !ok {
		t.Fatal("Herd Heirloom is registered")
	}
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) == 0 {
		t.Error("Herd Heirloom should declare a caveat for the unbuilt granted-trigger ability")
	}
}

func TestB295ChromaticOrreryTapsForFiveColorlessAndDrawsPerColorControlled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	orrery := pushCatalogPermanent(g, me.ID, "Chromatic Orrery", "Legendary Artifact", chromaticOrreryOracle, false)

	if err := g.ActivateManaAbility(me.ID, orrery, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); len(got) != 5 {
		t.Fatalf("pool = %v, want five colorless mana", got)
	}
	me.ManaPool.EmptyPool()
	untap(g, orrery)

	pushVanillaCreature(g, me.ID, "Colorless Guy", 1, 1)
	redMine := b31Push(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", "", "{R}", 2, 2, "R")
	greenMine := b31Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	_ = redMine
	_ = greenMine

	handBefore := me.Hand.Size()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	ctx := ctxFor(g, &game.StackItem{Controller: me.ID, SourceCardID: orrery})
	if err := chromaticOrreryDraw(g, ctx.Item); err != nil {
		t.Fatalf("chromaticOrreryDraw: %v", err)
	}
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand = %d, want %d (two colours: red and green)", got, handBefore+2)
	}

	spec, ok := Lookup(chromaticOrreryOracle)
	if !ok {
		t.Fatal("Chromatic Orrery is registered")
	}
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) == 0 {
		t.Error("Chromatic Orrery should declare a caveat for the unbuilt any-colour-spend static")
	}
}

func TestB295CovetedJewelDrawsThreeOnEntryAndTapsForThreeOfOneColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	handBefore := me.Hand.Size()
	jewel := castCatalogSpell(t, g, "Coveted Jewel", "Artifact", covetedJewelOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+3 {
		t.Fatalf("hand = %d, want %d (cast it, then draw three)", got, handBefore+3)
	}
	untap(g, jewel)
	if err := g.ActivateManaAbility(me.ID, jewel, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"B", "B", "B"}) {
		t.Errorf("pool = %v, want three black mana", got)
	}

	spec, ok := Lookup(covetedJewelOracle)
	if !ok {
		t.Fatal("Coveted Jewel is registered")
	}
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) == 0 {
		t.Error("Coveted Jewel should declare a caveat for the unbuilt unblocked-attack steal trigger")
	}
}

func TestB295TheEternityElevatorStationsAndUnlocksTheTwentyPlusAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elevator := pushCatalogPermanent(g, me.ID, "The Eternity Elevator", "Legendary Artifact — Spacecraft", theEternityElevatorOracle, false)

	if err := g.ActivateManaAbility(me.ID, elevator, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility 0: %v", err)
	}
	if got := poolColors(me); len(got) != 3 {
		t.Fatalf("pool = %v, want three colorless", got)
	}
	me.ManaPool.EmptyPool()
	untap(g, elevator)

	// Below 20 charge counters, the 20+ ability isn't there to activate.
	if err := g.ActivateManaAbility(me.ID, elevator, 1, game.ManaAbilityParams{Colors: []string{"W"}}); err == nil {
		t.Error("the 20+ ability activated below the threshold")
	}

	if err := g.AddCounterForEffect(elevator, game.CounterCharge, 20); err != nil {
		t.Fatalf("AddCounterForEffect: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, elevator, 1, game.ManaAbilityParams{Colors: []string{"W"}}); err != nil {
		t.Fatalf("ActivateManaAbility 1 at 20 counters: %v", err)
	}
	if got := poolColors(me); len(got) != 20 || got[0] != "W" {
		t.Errorf("pool = %v (len %d), want 20 white mana", got, len(got))
	}
}
