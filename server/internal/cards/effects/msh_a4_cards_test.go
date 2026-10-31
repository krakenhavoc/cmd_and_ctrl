package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_a4_cards_test.go — Marvel Super Heroes, half A, fourth batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshA4BruceOracle       = "b6c0ee14-2028-4e8b-9acd-ff78821bca9b"
	mshA4DocSamsonOracle   = "699610c9-c991-4faf-b712-d015ef0799ad"
	mshA4HerculesOracle    = "c3a8cc79-3771-484d-84dd-1a97f8bd12c2"
	mshA4HeroicFeastOracle = "31cae00a-3a74-4a59-80b0-a656f524e878"
	mshA4HiddenLairOracle  = "7069d241-4e66-40bf-afd1-551a4a5457f0"
	mshA4HourOracle        = "2774fcb7-8988-4a18-b5ec-7534dd8b2e05"
	mshA4HulklingOracle    = "03c934a6-be57-480d-8c84-4e46d9ccb406"
	mshA4HelperOracle      = "503928cd-1abd-44cd-92c3-6dc19bb07f78"
	mshA4IronLadOracle     = "3e715032-0d21-441b-b2d4-c91084728718"
	mshA4IronManOracle     = "94bc796d-8c04-4835-8a99-7e5dce9540ee"
)

func mshA4HasKeyword(g *game.Game, id uuid.UUID, kw string) bool {
	var has bool
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			has = game.HasKeyword(&c, kw)
		}
	})
	return has
}

func mshA4Ref(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

func TestMshA4EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mshA4BruceOracle:        "Bruce Banner",
		mshA4BruceOracle + "#1": "The Incredible Hulk",
		mshA4DocSamsonOracle:    "Doc Samson, Super Psychiatrist",
		mshA4HerculesOracle:     "Hercules, Prince of Power",
		mshA4HeroicFeastOracle:  "Heroic Feast",
		mshA4HiddenLairOracle:   "Hidden Lair",
		mshA4HourOracle:         "Hour of Defeat",
		mshA4HulklingOracle:     "Hulkling, Burgeoning Bruiser",
		mshA4HelperOracle:       "Hydraulic Helper",
		mshA4IronLadOracle:      "Iron Lad, Diverging Destiny",
		mshA4IronManOracle:      "Iron Man, Master of Machines",
	}
	for oracle, name := range want {
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

// --- Hour of Defeat ---------------------------------------------------------------

func TestMshA4HourOfDefeatDestroysTheCreatureThenSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Bin Me", "Keep Me")
	victim := pushDiesCreatureForTest(g, opp.ID, "Victim", "", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Hour of Defeat", "Instant", mshA4HourOracle, []game.TargetRef{mshA4Ref(victim)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the target creature survived")
	}
	c := surveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("want a surveil 1 prompt, got %+v", c)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if top, ok := g.LookupCardForEffect(c.ScryCards[0]); !ok || top.Name != "Bin Me" {
		t.Errorf("surveilled %v, want the top card", c.ScryCards)
	}
	if !me.Graveyard.Contains(c.ScryCards[0]) {
		t.Error("the surveilled card did not go to the graveyard")
	}
}

func TestMshA4HourOfDefeatSurveilsEvenWhenTheCreatureIsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Top Card", "Next Card")
	tough := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tough", TypeLine: "Creature — Golem", Power: 2, Toughness: 2,
		Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Hour of Defeat", "Instant", mshA4HourOracle, []game.TargetRef{mshA4Ref(tough)})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tough) {
		t.Error("an indestructible creature was destroyed")
	}
	if surveilChoiceFor(g, me.ID) == nil {
		t.Error("no surveil prompt after the destroy did nothing")
	}
}

// --- Hulkling ----------------------------------------------------------------------

func mshA4Hulkling(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "Hulkling, Burgeoning Bruiser", "Legendary Creature — Kree Skrull Hero", mshA4HulklingOracle, "{2}{G}", 2, 3, "G")
}

func mshA4CastCreature(t *testing.T, g *game.Game, name string, power, toughness int) uuid.UUID {
	t.Helper()
	return castFromHand(t, g, game.Card{Name: name, TypeLine: "Creature — Test", Power: power, Toughness: toughness}, game.CastSpellParams{})
}

func TestMshA4HulklingGrowsOnlyWhenTheEnteringCreatureHasGreaterPowerOrToughness(t *testing.T) {
	cases := []struct {
		name string
		p, t int
		want int
	}{
		{"greater power", 3, 1, 1},
		{"greater toughness", 1, 4, 1},
		{"equal both", 2, 3, 0},
		{"smaller both", 1, 1, 0},
	}
	for _, tc := range cases {
		g := newCatalogGame(t)
		me := g.Seats[0]
		hulk := mshA4Hulkling(g, me.ID)
		mshA4CastCreature(t, g, "Entrant", tc.p, tc.t)
		passPriorityAroundTable(t, g)
		if got := plusOnes(g, hulk); got != tc.want {
			t.Errorf("%s: Hulkling has %d counters, want %d", tc.name, got, tc.want)
		}
	}
}

func TestMshA4HulklingIgnoresAnOpponentsCreatureAndHasVigilance(t *testing.T) {
	g := newCatalogGame(t)
	hulk := mshA4Hulkling(g, g.Seats[0].ID)
	castFlashCreatureFrom(t, g, g.Seats[1], game.Card{Name: "Big Theirs", TypeLine: "Creature — Test", Power: 9, Toughness: 9})
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, hulk); got != 0 {
		t.Errorf("an opponent's creature grew Hulkling: %d counters", got)
	}
	if !mshA4HasKeyword(g, hulk, "vigilance") {
		t.Error("Hulkling has no vigilance")
	}
}

func TestMshA4HulklingChecksTheComparisonAgainAsItResolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hulk := mshA4Hulkling(g, me.ID)
	mshA4CastCreature(t, g, "Entrant", 3, 3)
	// The trigger is on the stack; Hulkling grows in response to 3/4.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(hulk, game.CounterPlusOne, 2) })
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, hulk); got != 2 {
		t.Errorf("Hulkling has %d counters, want just the 2 added in response (the 3/3 is no longer greater)", got)
	}
}

// --- Heroic Feast ------------------------------------------------------------------

func mshA4HeroicFeast(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Heroic Feast", "Enchantment", mshA4HeroicFeastOracle, nil)
	passPriorityAroundTable(t, g)
}

func TestMshA4HeroicFeastMakesAFoodWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	mshA4HeroicFeast(t, g)
	if got := b16CountNamed(g, "Food"); got != 1 {
		t.Errorf("Food tokens = %d, want 1", got)
	}
}

func TestMshA4HeroicFeastPutsACounterOnEachChosenCreatureUpToTheLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mshA4HeroicFeast(t, g)
	a := pushDiesCreatureForTest(g, me.ID, "Bear A", "", "Creature — Bear", 2, 2)
	b := pushDiesCreatureForTest(g, me.ID, "Bear B", "", "Creature — Bear", 2, 2)
	c := pushDiesCreatureForTest(g, me.ID, "Bear C", "", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	pickTargetRefs(t, g, me.ID, nil, mshA4Ref(a), mshA4Ref(b))
	passPriorityAroundTable(t, g)
	if plusOnes(g, a) != 1 || plusOnes(g, b) != 1 || plusOnes(g, c) != 0 {
		t.Errorf("counters a=%d b=%d c=%d, want 1 1 0", plusOnes(g, a), plusOnes(g, b), plusOnes(g, c))
	}
}

func TestMshA4HeroicFeastRefusesMoreTargetsThanLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mshA4HeroicFeast(t, g)
	a := pushDiesCreatureForTest(g, me.ID, "Bear A", "", "Creature — Bear", 2, 2)
	b := pushDiesCreatureForTest(g, me.ID, "Bear B", "", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the life-gain trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{mshA4Ref(a), mshA4Ref(b)}); err == nil {
		t.Fatal("two targets were accepted for one life gained")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{mshA4Ref(a)}); err != nil {
		t.Fatalf("one target for one life: %v", err)
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, a) != 1 || plusOnes(g, b) != 0 {
		t.Errorf("counters a=%d b=%d, want 1 0", plusOnes(g, a), plusOnes(g, b))
	}
}

func TestMshA4HeroicFeastIgnoresAnOpponentsLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mshA4HeroicFeast(t, g)
	a := pushDiesCreatureForTest(g, me.ID, "Bear A", "", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 3) })
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || plusOnes(g, a) != 0 {
		t.Error("an opponent gaining life triggered Heroic Feast")
	}
}

// --- Hydraulic Helper ---------------------------------------------------------------

func TestMshA4HydraulicHelperManaPaysForArtifactSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	helper := b27Push(g, me.ID, "Hydraulic Helper", "Artifact Creature — Robot", mshA4HelperOracle, "{1}{U}", 2, 3, "U")
	advanceToMain(t, g)
	b28TapForMana(t, g, me.ID, helper, "")
	if got := poolColors(me); len(got) != 1 || got[0] != "U" {
		t.Fatalf("pool after tapping: %v, want one {U}", got)
	}
	if _, err := castFromHandErr(g, me, "Plain Spell", "Sorcery", "{U}", "", strict); err == nil {
		t.Fatal("the Helper's mana paid for a nonartifact spell")
	}
	if _, err := castFromHandErr(g, me, "Gadget", "Artifact", "{U}", "", strict); err != nil {
		t.Fatalf("the Helper's mana could not pay for an artifact spell: %v", err)
	}
	if !mshA4HasKeyword(g, helper, "defender") {
		t.Error("Hydraulic Helper has no defender")
	}
}

// --- Hidden Lair --------------------------------------------------------------------

func mshA4TapSecondRow(t *testing.T, g *game.Game, who, land uuid.UUID, color string) {
	t.Helper()
	if err := g.ActivateManaAbility(who, land, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	b10ResolveAllManaPicks(t, g, who, color)
}

func TestMshA4HiddenLairTapsForColourOnTheTurnItEntersWithNoBasic(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := playLandFromHand(t, g, "Hidden Lair", mshA4HiddenLairOracle)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	mshA4TapSecondRow(t, g, me.ID, land, "B")
	if got := poolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("tapped for {B} the turn it entered: pool %v", got)
	}
}

func TestMshA4HiddenLairNeedsABasicLandAfterTheTurnItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushCatalogPermanent(g, me.ID, "Hidden Lair", "Land", mshA4HiddenLairOracle, false)
	pushCatalogPermanent(g, me.ID, "Other Place", "Land", "", false)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("the coloured ability worked with no basic land and without having entered this turn")
	}
	if b16Tapped(t, g, land) {
		t.Error("a refused activation tapped the land")
	}
	pushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	mshA4TapSecondRow(t, g, me.ID, land, "U")
	if got := poolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("tapped for {U} with a basic land: pool %v", got)
	}
}

func TestMshA4HiddenLairAlwaysTapsForColourless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushCatalogPermanent(g, me.ID, "Hidden Lair", "Land", mshA4HiddenLairOracle, false)
	advanceToMain(t, g)
	b28TapForMana(t, g, me.ID, land, "")
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("colourless ability: pool %v", got)
	}
}

// --- Doc Samson ---------------------------------------------------------------------

func mshA4DocSamson(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "Doc Samson, Super Psychiatrist", "Legendary Creature — Gamma Doctor Hero", mshA4DocSamsonOracle, "{4}{G}", 3, 6, "G")
}

func mshA4PutCounters(g *game.Game, id uuid.UUID, kind string, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, kind, n) })
}

func TestMshA4DocSamsonAddsOneCounterOfEachKindOnPermanentsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mshA4DocSamson(g, me.ID)
	bear := pushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	land := pushCatalogPermanent(g, me.ID, "Odd Land", "Land", "", false)
	theirs := pushDiesCreatureForTest(g, opp.ID, "Their Bear", "", "Creature — Bear", 2, 2)

	mshA4PutCounters(g, bear, game.CounterPlusOne, 2)
	if got := plusOnes(g, bear); got != 3 {
		t.Errorf("two +1/+1 counters on my creature became %d, want 3", got)
	}
	mshA4PutCounters(g, bear, "charge", 1)
	if got := countersOn(g, bear, "charge"); got != 2 {
		t.Errorf("one charge counter became %d, want 2 (each kind gets its own extra)", got)
	}
	mshA4PutCounters(g, land, "charge", 1)
	if got := countersOn(g, land, "charge"); got != 2 {
		t.Errorf("a land is a permanent you control: one charge counter became %d, want 2", got)
	}
	mshA4PutCounters(g, theirs, game.CounterPlusOne, 2)
	if got := plusOnes(g, theirs); got != 2 {
		t.Errorf("an opponent's creature got %d counters, want exactly the 2 put", got)
	}
}

func TestMshA4DocSamsonTapsForManaOfOneColourEqualToHisPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	samson := mshA4DocSamson(g, me.ID)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, samson, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	b10ResolveAllManaPicks(t, g, me.ID, "R")
	if got := poolColors(me); len(got) != 3 || got[0] != "R" || got[1] != "R" || got[2] != "R" {
		t.Errorf("a 3-power Samson made %v, want three {R}", got)
	}
}

func TestMshA4DocSamsonsPowerCountsHisOwnReplacedCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	samson := mshA4DocSamson(g, me.ID)
	mshA4PutCounters(g, samson, game.CounterPlusOne, 1)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, samson, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	b10ResolveAllManaPicks(t, g, me.ID, "G")
	if got := poolColors(me); len(got) != 5 {
		t.Errorf("Samson with two counters (power 5) made %v, want five mana", got)
	}
}

// --- Hercules -----------------------------------------------------------------------

func TestMshA4HerculesPowerUpGrowsAndGrantsThreeKeywordsOnceAtFullPriceLater(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	herc := castPowerUpCreature(t, g, me, "Hercules, Prince of Power", "Legendary Creature — Demigod Warrior Hero", "{2}{G}", "{C}{C}{G}", mshA4HerculesOracle, 3)
	if got := chargedPowerUp(t, g, me.ID, herc); got != "{2}" {
		t.Fatalf("the turn he entered: %q, want {2}", got)
	}
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, herc, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := plusOnes(g, herc); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	for _, kw := range []string{"vigilance", "indestructible", "haste"} {
		if !mshA4HasKeyword(g, herc, kw) {
			t.Errorf("Hercules lacks %s after the power-up", kw)
		}
	}
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, herc, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a second power-up was allowed")
	}
}

func TestMshA4HerculesKeywordsEndWithTheTurnAndTheCostIsFullLater(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	herc := pushPowerUpPermanent(g, me.ID, "Hercules, Prince of Power", "Legendary Creature — Demigod Warrior Hero", "{2}{G}", mshA4HerculesOracle)
	if got := chargedPowerUp(t, g, me.ID, herc); got != "{4}{G}" {
		t.Fatalf("a Hercules that did not enter this turn: %q, want {4}{G}", got)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, herc, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	if !mshA4HasKeyword(g, herc, "indestructible") {
		t.Fatal("no indestructible this turn")
	}
	powerUpPassTo(t, g, 0)
	if mshA4HasKeyword(g, herc, "indestructible") || mshA4HasKeyword(g, herc, "haste") {
		t.Error("the keywords outlasted the turn")
	}
	if got := plusOnes(g, herc); got != 1 {
		t.Errorf("the counter is permanent: %d, want 1", got)
	}
}

// --- Iron Man, Master of Machines ----------------------------------------------------

func mshA4IronMan(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "Iron Man, Master of Machines", "Legendary Artifact Creature — Human Hero", mshA4IronManOracle, "{2}{U}{R}", 1, 4, "U", "R")
}

func mshA4Power(g *game.Game, id uuid.UUID) int {
	var p int
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			p = c.CurrentPower()
		}
	})
	return p
}

func TestMshA4IronManGetsPlusOnePowerPerOtherArtifactYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	iron := mshA4IronMan(g, me.ID)
	if got := mshA4Power(g, iron); got != 1 {
		t.Fatalf("alone: power %d, want 1 (he does not count himself)", got)
	}
	b27Push(g, me.ID, "Rock", "Artifact", "", "{2}", 0, 0)
	b27Push(g, me.ID, "Gear", "Artifact Creature — Construct", "", "{1}", 1, 1)
	b27Push(g, opp.ID, "Their Rock", "Artifact", "", "{4}", 0, 0)
	b27Push(g, me.ID, "Not Artifact", "Creature — Bear", "", "{1}", 2, 2)
	if got := mshA4Power(g, iron); got != 3 {
		t.Errorf("with two other artifacts of mine: power %d, want 3", got)
	}
	if !mshA4HasKeyword(g, iron, "flying") || !mshA4HasKeyword(g, iron, "vigilance") {
		t.Error("Iron Man lacks flying or vigilance")
	}
}

func TestMshA4IronManDrawsOnAttackOnlyIfAnArtifactEnteredThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	iron := mshA4IronMan(g, me.ID)
	advanceToMain(t, g)
	castFromHand(t, g, game.Card{Name: "Gadget", TypeLine: "Artifact"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(iron, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d after an artifact entered, want 1", got)
	}
}

func TestMshA4IronManDoesNotDrawWhenNoArtifactEnteredThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	iron := mshA4IronMan(g, me.ID)
	b27Push(g, me.ID, "Old Rock", "Artifact", "", "{2}", 0, 0)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(iron, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d with no artifact having entered this turn, want 0", got)
	}
}

// --- Iron Lad, Diverging Destiny ------------------------------------------------------

func TestMshA4IronLadRevealsTheTopCardAndDrawsOnlyAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lad := b27Push(g, me.ID, "Iron Lad, Diverging Destiny", "Legendary Artifact Creature — Human Hero", mshA4IronLadOracle, "{2}{U}", 2, 2, "U")
	advanceToMain(t, g)

	seedLibraryTop(me, "Plain Bear", "Creature — Bear", "{1}{G}")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, lad, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("a nonartifact top card drew %d, want 0", got)
	}
	if !b16Tapped(t, g, lad) {
		t.Error("the ability did not tap Iron Lad")
	}

	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(lad) })
	artifact := seedLibraryTop(me, "Sprocket", "Artifact", "{1}")
	b16Activate(t, g, me.ID, lad, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("an artifact top card drew %d, want 1", got)
	}
	if !me.Hand.Contains(artifact) {
		t.Error("the drawn card is not the revealed artifact")
	}
}

func TestMshA4IronLadShowsYouTheTopOfYourLibrary(t *testing.T) {
	spec, _ := Lookup(mshA4IronLadOracle)
	if spec.LibraryTopVisible != game.LibraryTopOwner {
		t.Errorf("LibraryTopVisible = %v, want the owner", spec.LibraryTopVisible)
	}
	for _, kw := range []string{"flying", "vigilance"} {
		found := false
		for _, k := range spec.PrintedKeywords {
			found = found || k == kw
		}
		if !found {
			t.Errorf("Iron Lad lacks %s", kw)
		}
	}
}

// --- Bruce Banner // The Incredible Hulk -----------------------------------------------

func mshA4BruceCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   mshA4BruceOracle,
		Layout:     game.LayoutModalDFC,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Bruce Banner", TypeLine: "Legendary Creature — Human Scientist Hero", ManaCost: "{U}", Colors: []string{"U"}, Power: 1, Toughness: 1},
			{Name: "The Incredible Hulk", TypeLine: "Legendary Creature — Gamma Berserker Hero", ManaCost: "{2}{R}{R}{G}{G}", Colors: []string{"R", "G"}, Power: 8, Toughness: 8},
		},
	}
	c.SetFace(0)
	return c
}

func mshA4TransformedHulk(g *game.Game, owner uuid.UUID) uuid.UUID {
	c := mshA4BruceCard(owner)
	c.SetFace(1)
	return pushBattlefieldCardWithTimestamp(g, c)
}

func TestMshA4BruceBannerDrawsXCardsForTwiceXAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bruce := pushBattlefieldCardWithTimestamp(g, mshA4BruceCard(me.ID))
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{C}{C}")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, bruce, 0, game.ActivateAbilityParams{XValue: 2})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d with X=2, want 2", got)
	}
	if !b16Tapped(t, g, bruce) {
		t.Error("the ability did not tap Bruce Banner")
	}
	if len(poolColors(me)) != 0 {
		t.Errorf("mana left in the pool after {X}{X} with X=2: %v", poolColors(me))
	}
}

func TestMshA4BruceBannerCannotDrawOutsideASorceryWindow(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bruce := pushBattlefieldCardWithTimestamp(g, mshA4BruceCard(me.ID))
	advanceTo(t, g, game.StepBeginCombat)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bruce, 0, game.ActivateAbilityParams{XValue: 1}); err == nil {
		t.Error("the draw ability was activated outside a main phase")
	}
}

func TestMshA4BruceBannerTransformsForSixAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bruce := pushBattlefieldCardWithTimestamp(g, mshA4BruceCard(me.ID))
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{R}{R}{G}{G}")
	b16Activate(t, g, me.ID, bruce, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, bruce)
	if !ok || c.Name != "The Incredible Hulk" {
		t.Fatalf("after the transform: %+v", c)
	}
	if c.CurrentPower() != 8 || c.CurrentToughness() != 8 {
		t.Errorf("The Incredible Hulk is %d/%d, want 8/8", c.CurrentPower(), c.CurrentToughness())
	}
	if !mshA4HasKeyword(g, bruce, "reach") || !mshA4HasKeyword(g, bruce, "trample") {
		t.Error("The Incredible Hulk lacks reach or trample")
	}
}

func TestMshA4IncredibleHulkEnrageGrowsHimAndOnlyUntapsWhenAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hulk := mshA4TransformedHulk(g, me.ID)
	src := pushVanillaCreature(g, g.Seats[1].ID, "Pinger", 1, 1)
	b16Tap(g, hulk)
	mshA4Damage(g, src, hulk, 2)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, hulk); got != 1 {
		t.Fatalf("Enrage: %d counters, want 1", got)
	}
	if !b16Tapped(t, g, hulk) {
		t.Error("a Hulk that is not attacking was untapped")
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step == game.StepBeginCombat && next.PhaseOrdinal > 1 {
		t.Errorf("an extra combat appeared for a Hulk that was not attacking: %+v", next)
	}
}

func TestMshA4IncredibleHulkEnrageWhileAttackingUntapsHimAndAddsACombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hulk := mshA4TransformedHulk(g, me.ID)
	src := pushVanillaCreature(g, opp.ID, "Pinger", 1, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(hulk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, hulk) {
		t.Fatal("the attacking Hulk is not tapped")
	}
	mshA4Damage(g, src, hulk, 1)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, hulk); got != 1 {
		t.Errorf("Enrage: %d counters, want 1", got)
	}
	if b16Tapped(t, g, hulk) {
		t.Error("the attacking Hulk was not untapped")
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat || next.PhaseOrdinal != 2 {
		t.Errorf("after the combat: %+v, want an additional combat", next)
	}
}

// mshA4Damage deals noncombat damage to a creature as one damage
// instruction and lets the enrage trigger, which waits for the batch, fire.
func mshA4Damage(g *game.Game, source, target uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(source, target, n) })
	g.RunStateChecksForTest()
}

func TestMshA4IncredibleHulkEnrageTriggersOnceForDamageFromTwoSourcesAtOnce(t *testing.T) {
	g := newCatalogGame(t)
	hulk := mshA4TransformedHulk(g, g.Seats[0].ID)
	a := pushVanillaCreature(g, g.Seats[1].ID, "Pinger A", 1, 1)
	b := pushVanillaCreature(g, g.Seats[1].ID, "Pinger B", 1, 1)
	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			_ = g.DealDamageToCreatureForEffect(a, hulk, 1)
			return g.DealDamageToCreatureForEffect(b, hulk, 1)
		})
	})
	g.RunStateChecksForTest()
	if n := radTriggersFrom(g, hulk); n != 1 {
		t.Fatalf("two sources at once made %d Enrage triggers, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, hulk); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
}

func TestMshA4IncredibleHulkEnrageTriggersOncePerDamageEvent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hulk := mshA4TransformedHulk(g, me.ID)
	a := pushVanillaCreature(g, g.Seats[1].ID, "Pinger A", 1, 1)
	b := pushVanillaCreature(g, g.Seats[1].ID, "Pinger B", 1, 1)
	mshA4Damage(g, a, hulk, 1)
	passPriorityAroundTable(t, g)
	mshA4Damage(g, b, hulk, 1)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, hulk); got != 2 {
		t.Errorf("two separate damage instructions gave %d counters, want 2", got)
	}
}
