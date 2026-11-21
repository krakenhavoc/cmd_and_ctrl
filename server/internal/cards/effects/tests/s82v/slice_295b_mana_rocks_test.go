package testss82v

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsch "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ch"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestB295PrismaticLensTapsForColorlessAndPaidAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lens := effectstest.PushCatalogPermanent(g, me.ID, "Prismatic Lens", "Artifact", effectstest.PrismaticLensOracle, false)

	if err := g.ActivateManaAbility(me.ID, lens, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility 0: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"C"}) {
		t.Errorf("pool after ability 0 = %v, want [C]", got)
	}
	me.ManaPool.EmptyPool()
	effectstest.Untap(g, lens)

	// The second ability needs a {1} to pay before it taps.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, lens, 1, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("ActivateManaAbility 1: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"G"}) {
		t.Errorf("pool after ability 1 = %v, want [G] (the {1} was spent, not produced)", got)
	}
}

func TestB295AstralCornucopiaEntersWithXCountersAndTapsForThatManyOfOneColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	rock := effectstest.B12PlayFromHand(t, g, "Astral Cornucopia", "Artifact", effectstest.AstralCornucopiaOracle, game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, rock, game.CounterCharge); got != 3 {
		t.Fatalf("charge counters = %d, want 3", got)
	}
	me := g.Seats[0]
	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"U", "U", "U"}) {
		t.Errorf("pool = %v, want three blue mana (one per charge counter)", got)
	}
}

func TestB295CursedMirrorTapsForRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mirror := effectstest.PushCatalogPermanent(g, me.ID, "Cursed Mirror", "Artifact", effectstest.CursedMirrorOracle, false)
	if err := g.ActivateManaAbility(me.ID, mirror, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"R"}) {
		t.Errorf("pool = %v, want [R]", got)
	}
	spec, ok := effects.Lookup(effectstest.CursedMirrorOracle)
	if !ok {
		t.Fatal("Cursed Mirror is registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Error("Cursed Mirror is full since #1593 built its copy-until-end-of-turn ETB")
	}
}

func TestB295LiquimetalTorqueMakesATargetAnArtifactUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	torque := effectstest.PushCatalogPermanent(g, me, "Liquimetal Torque", "Artifact", effectstest.LiquimetalTorqueOracle, false)
	bear := effectstest.PushVanillaCreature(g, them, "Grizzly Bears", 2, 2)

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
	effectstest.PassPriorityAroundTable(t, g)

	types := effectstest.EffectiveTypes(t, g, bear)
	if !effectstest.HasAbility(types, "Artifact") {
		t.Errorf("types = %v, want Artifact added", types)
	}
	if !effectstest.HasAbility(types, "Creature") {
		t.Errorf("types = %v, the printed type should still be there too", types)
	}
}

func TestB295MidnightClockTapsForBlueAndAccumulatesHourCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	clock := effectstest.PushCatalogPermanent(g, me.ID, "Midnight Clock", "Artifact", effectstest.MidnightClockOracle, false)

	if err := g.ActivateManaAbility(me.ID, clock, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"U"}) {
		t.Errorf("pool = %v, want [U]", got)
	}

	// The upkeep trigger fires for ANY player's upkeep.
	before := effectstest.CounterOn(g, clock, "hour")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, clock, "hour"); got != before+1 {
		t.Errorf("hour counters = %d, want %d after an upkeep passes", got, before+1)
	}

	spec, ok := effects.Lookup(effectstest.MidnightClockOracle)
	if !ok {
		t.Fatal("Midnight Clock is registered")
	}
	if spec.Completeness != effects.CompletenessFull {
		t.Errorf("Midnight Clock's twelfth-counter payoff is built (#1727): Completeness = %v", spec.Completeness)
	}
}

func TestB295CagedSunPumpsCreaturesOfTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	redMine := effectstest.B31Push(g, me, "Goblin Guide", "Creature — Goblin Scout", "", "{R}", 2, 2, "R")
	greenMine := effectstest.B31Push(g, me, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	redTheirs := effectstest.B31Push(g, them, "Raging Goblin", "Creature — Goblin Berserker", "", "{R}", 1, 1, "R")

	effectstest.PushChosenColorPermanent(t, g, me, "Caged Sun", "Artifact", effectstest.CagedSunOracle, "R")

	if got := effectstest.EffectivePower(t, g, redMine); got != 3 {
		t.Errorf("redMine power = %d, want 3 (+1/+1)", got)
	}
	if got := effectstest.EffectivePower(t, g, greenMine); got != 2 {
		t.Errorf("greenMine power = %d, want 2 (unaffected)", got)
	}
	if got := effectstest.EffectivePower(t, g, redTheirs); got != 1 {
		t.Errorf("redTheirs power = %d, want 1 — the anthem has no controller clause, only colour", got)
	}

	spec, ok := effects.Lookup(effectstest.CagedSunOracle)
	if !ok {
		t.Fatal("Caged Sun is registered")
	}
	if spec.Completeness != effects.CompletenessFull {
		t.Errorf("Completeness = %v, want CompletenessFull — the mana-doubling clause is implemented too (#1600)", spec.Completeness)
	}
}

func TestB295HerdHeirloomManaIsRestrictedToCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.SeedPermanentWithOracle(g, me.ID, "Herd Heirloom", "Artifact", effectstest.HerdHeirloomOracle)

	if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 1 {
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

	spec, ok := effects.Lookup(effectstest.HerdHeirloomOracle)
	if !ok {
		t.Fatal("Herd Heirloom is registered")
	}
	if spec.Completeness != effects.CompletenessFull {
		t.Errorf("Herd Heirloom's granted-trigger ability is built (#1600): Completeness = %v", spec.Completeness)
	}
}

func TestB295ChromaticOrreryTapsForFiveColorlessAndDrawsPerColorControlled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	orrery := effectstest.PushCatalogPermanent(g, me.ID, "Chromatic Orrery", "Legendary Artifact", effectstest.ChromaticOrreryOracle, false)

	if err := g.ActivateManaAbility(me.ID, orrery, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 5 {
		t.Fatalf("pool = %v, want five colorless mana", got)
	}
	me.ManaPool.EmptyPool()
	effectstest.Untap(g, orrery)

	effectstest.PushVanillaCreature(g, me.ID, "Colorless Guy", 1, 1)
	redMine := effectstest.B31Push(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", "", "{R}", 2, 2, "R")
	greenMine := effectstest.B31Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	_ = redMine
	_ = greenMine

	handBefore := me.Hand.Size()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	ctx := effectstest.CtxFor(g, &game.StackItem{Controller: me.ID, SourceCardID: orrery})
	if err := cardsch.ChromaticOrreryDraw(g, ctx.Item); err != nil {
		t.Fatalf("chromaticOrreryDraw: %v", err)
	}
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand = %d, want %d (two colours: red and green)", got, handBefore+2)
	}

	spec, ok := effects.Lookup(effectstest.ChromaticOrreryOracle)
	if !ok {
		t.Fatal("Chromatic Orrery is registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 || len(spec.AnyColorSpend) == 0 {
		t.Errorf("Chromatic Orrery's any-colour spend is built (#1600): Completeness = %v, caveats %v", spec.Completeness, spec.Caveats)
	}
}

func TestB295CovetedJewelDrawsThreeOnEntryAndTapsForThreeOfOneColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	handBefore := me.Hand.Size()
	jewel := effectstest.CastCatalogSpell(t, g, "Coveted Jewel", "Artifact", effectstest.CovetedJewelOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+3 {
		t.Fatalf("hand = %d, want %d (cast it, then draw three)", got, handBefore+3)
	}
	effectstest.Untap(g, jewel)
	if err := g.ActivateManaAbility(me.ID, jewel, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); !reflect.DeepEqual(got, []string{"B", "B", "B"}) {
		t.Errorf("pool = %v, want three black mana", got)
	}

	spec, ok := effects.Lookup(effectstest.CovetedJewelOracle)
	if !ok {
		t.Fatal("Coveted Jewel is registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Coveted Jewel's steal trigger is built (#1600, coveted_jewel_test.go): Completeness = %v, caveats %v", spec.Completeness, spec.Caveats)
	}
}

func TestB295TheEternityElevatorStationsAndUnlocksTheTwentyPlusAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elevator := effectstest.PushCatalogPermanent(g, me.ID, "The Eternity Elevator", "Legendary Artifact — Spacecraft", effectstest.TheEternityElevatorOracle, false)

	if err := g.ActivateManaAbility(me.ID, elevator, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility 0: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 3 {
		t.Fatalf("pool = %v, want three colorless", got)
	}
	me.ManaPool.EmptyPool()
	effectstest.Untap(g, elevator)

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
	if got := effectstest.PoolColors(me); len(got) != 20 || got[0] != "W" {
		t.Errorf("pool = %v (len %d), want 20 white mana", got, len(got))
	}
}
