package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// otj_r_cards_test.go — Outlaws of Thunder Junction (OTJ), the R cards
// (#4468). Each test plays the real catalog entry through the engine and
// asserts what the printed text says.

const (
	otjRRailwayBrawlerOracle    = "b653db44-ddae-44e8-8fa6-6af3232cd586"
	otjRRakishCrewOracle        = "ddf34404-39dd-4dcf-a706-ec78d47167e9"
	otjRRattlebackOracle        = "5538d4cf-98eb-4e20-a9c1-42c7cdcad194"
	otjRRaucousOracle           = "534e74f6-52f4-4824-9178-cce5e4bcbcc2"
	otjRRavenOracle             = "eec7d4f4-4141-49ed-8978-aefacc0a6609"
	otjRRazzleDazzlerOracle     = "a1ac6d5c-b2d0-4ca4-8af8-16fd3dd3b3b5"
	otjRReachForTheSkyOracle    = "68930fb6-d833-40cf-8917-1a1bf60edbe3"
	otjRRecklessLackeyOracle    = "73733fa5-d292-40ec-8f17-05ba5feb40ab"
	otjRRedrockSentinelOracle   = "51a6819d-c8fb-4f52-ab0c-4ae845728ed2"
	otjRRoadrunnerOracle        = "78c7c73f-847d-43bf-9e02-cf85357453f4"
	otjRRictusRobberOracle      = "91e49e8e-575f-43c8-9263-a8568697e7f4"
	otjRRiseOfTheVarmintsOracle = "d981977d-acfa-432a-b1d7-1808a89a5998"
	otjRRodeoOracle             = "82debe1c-31cb-411f-a353-98a952180666"
	otjRRooftopAssassinOracle   = "f3baef8c-f8d7-4cb2-b5cc-b889a991e4e8"
	otjRRustlerRampageOracle    = "ef0c0f28-e1ee-42a1-9878-95a0ceb0a9fa"
	otjRLawbringerOracle        = "82b6a283-9897-4ce1-8de3-e7579cc9b75b"
)

func TestOtjREveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		otjRRailwayBrawlerOracle:    "Railway Brawler",
		otjRRakishCrewOracle:        "Rakish Crew",
		otjRRattlebackOracle:        "Rattleback Apothecary",
		otjRRaucousOracle:           "Raucous Entertainer",
		otjRRavenOracle:             "Raven of Fell Omens",
		otjRRazzleDazzlerOracle:     "Razzle-Dazzler",
		otjRReachForTheSkyOracle:    "Reach for the Sky",
		otjRRecklessLackeyOracle:    "Reckless Lackey",
		otjRRedrockSentinelOracle:   "Redrock Sentinel",
		otjRRoadrunnerOracle:        "Resilient Roadrunner",
		otjRRictusRobberOracle:      "Rictus Robber",
		otjRRiseOfTheVarmintsOracle: "Rise of the Varmints",
		otjRRodeoOracle:             "Rodeo Pyromancers",
		otjRRooftopAssassinOracle:   "Rooftop Assassin",
		otjRRustlerRampageOracle:    "Rustler Rampage",
		otjRLawbringerOracle:        "Ruthless Lawbringer",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	for oracle, cost := range map[string]string{
		otjRRailwayBrawlerOracle:    "{3}{G}",
		otjRRictusRobberOracle:      "{2}{B}",
		otjRRiseOfTheVarmintsOracle: "{2}{G}",
	} {
		if sa := effectstest.PlotOf(oracle); sa == nil || sa.Cost != cost {
			t.Errorf("%s: plot %+v, want %s", oracle, sa, cost)
		}
	}
}

// otjRTreasures counts the Treasure tokens `controller` controls.
func otjRTreasures(g *game.Game, controller uuid.UUID) int {
	n, _ := effectstest.CountTokensNamed(g, controller, "Treasure")
	return n
}

// --- Railway Brawler ---------------------------------------------------------

func TestOtjRRailwayBrawlerPutsCountersEqualToTheNewcomersPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brawler := effectstest.B12Push(g, me.ID, "Railway Brawler", "Creature — Rhino Warrior", otjRRailwayBrawlerOracle, 5, 5)
	for _, kw := range []string{"reach", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, brawler, kw) {
			t.Errorf("%s is printed", kw)
		}
	}
	ogre := effectstest.CastHoldCreature(t, g, "Ogre", "Creature — Ogre", "", 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, ogre); got != 3 {
		t.Errorf("a 3/3 entered: %d counters, want 3", got)
	}
	if got := effectstest.PlusOneCounters(g, brawler); got != 0 {
		t.Errorf("the Brawler put %d counters on itself", got)
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if got := effectstest.PlusOneCounters(g, theirs); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
}

// --- Rakish Crew -------------------------------------------------------------

func TestOtjRRakishCrewMakesAMercenaryAndDrainsWhenAnOutlawDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	rogue := effectstest.B12Creature(g, me.ID, "My Rogue", "Creature — Human Rogue", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirRogue := effectstest.B12Creature(g, opp.ID, "Their Rogue", "Creature — Human Rogue", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Rakish Crew", "Enchantment", otjRRakishCrewOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Mercenary"); n != 1 {
		t.Fatalf("%d Mercenary tokens, want 1", n)
	}

	life, oppLife, otherLife := me.Life, opp.Life, other.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || me.Life != life || other.Life != otherLife {
		t.Fatalf("a Bear is no outlaw: life %d → %d", life, me.Life)
	}

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(rogue))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rogue) {
		t.Fatal("the Rogue did not die")
	}
	if me.Life != life+1 || opp.Life != oppLife-1 || other.Life != otherLife-1 {
		t.Errorf("your Rogue died: you %d → %d, opponents %d → %d and %d → %d, want +1 and -1 each",
			life, me.Life, oppLife, opp.Life, otherLife, other.Life)
	}

	life, otherLife = me.Life, other.Life
	otjB1BoltOpponent(t, g, opp.ID) // a crime, nothing more
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(theirRogue))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirRogue) || me.Life != life || other.Life != otherLife {
		t.Errorf("an opponent's outlaw died: you %d → %d, want no drain", life, me.Life)
	}
}

// --- Rattleback Apothecary ---------------------------------------------------

func TestOtjRRattlebackApothecaryGivesTheChosenKeywordAfterACrime(t *testing.T) {
	for _, tc := range []struct {
		option int
		want   string
		not    string
	}{
		{0, "menace", "lifelink"},
		{1, "lifelink", "menace"},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		apothecary := effectstest.B12Push(g, me.ID, "Rattleback Apothecary", "Creature — Gorgon Warlock", otjRRattlebackOracle, 3, 2)
		if !effectstest.HasEffectiveKeyword(t, g, apothecary, "deathtouch") {
			t.Error("deathtouch is printed")
		}
		bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		giant := effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		effectstest.AdvanceToMain(t, g)

		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(giant))
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.LatestOptionPickFor(g, me.ID) != nil {
			t.Fatal("targeting your own creature is no crime")
		}

		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opp.ID))
		p := effectstest.LatestPickTarget(g, me.ID)
		if p == nil {
			t.Fatal("the crime asked for no target")
		}
		if !effectstest.HasID(p.PickTargetCards, bear) || effectstest.HasID(p.PickTargetCards, theirs) {
			t.Fatalf("target creature you control: %v", p.PickTargetCards)
		}
		effectstest.PickCard(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.LatestOptionPickFor(g, me.ID)
		if c == nil || len(c.PickOptions) != 2 {
			t.Fatalf("keyword choice %+v, want menace or lifelink", c)
		}
		if err := g.ResolveOptionPick(c.ID, me.ID, tc.option); err != nil {
			t.Fatalf("ResolveOptionPick: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, bear, tc.want) || effectstest.HasEffectiveKeyword(t, g, bear, tc.not) {
			t.Errorf("option %d: abilities %v, want %s only", tc.option, effectstest.EffectiveAbilities(t, g, bear), tc.want)
		}
		effectstest.AdvanceToMainOf(t, g, 1)
		if effectstest.HasEffectiveKeyword(t, g, bear, tc.want) {
			t.Errorf("%s lasts until end of turn only", tc.want)
		}
	}
}

// --- Raucous Entertainer -----------------------------------------------------

func TestOtjRRaucousEntertainerCountersOnlyCreaturesThatEnteredThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ent := effectstest.B12Push(g, me.ID, "Raucous Entertainer", "Creature — Plant Bard", otjRRaucousOracle, 2, 2)
	old := effectstest.B12Creature(g, me.ID, "Old Bear", "Creature — Bear", 2, 2)
	fresh := effectstest.CastHoldCreature(t, g, "Fresh Bear", "Creature — Bear", "", 2, 2)
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, ent, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, fresh); got != 1 {
		t.Errorf("the creature that entered this turn has %d counters, want 1", got)
	}
	if got := effectstest.PlusOneCounters(g, old); got != 0 {
		t.Errorf("a creature from an earlier turn got %d counters", got)
	}
	if !effectstest.B16Tapped(t, g, ent) {
		t.Error("{T} is part of the cost")
	}
}

// --- Raven of Fell Omens -----------------------------------------------------

func TestOtjRRavenOfFellOmensDrainsOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	raven := effectstest.B12Push(g, me.ID, "Raven of Fell Omens", "Creature — Bird", otjRRavenOracle, 1, 2)
	if !effectstest.HasEffectiveKeyword(t, g, raven, "flying") {
		t.Error("flying is printed")
	}
	effectstest.AdvanceToMain(t, g)
	life, oppLife, otherLife := me.Life, opp.Life, other.Life
	otjB1BoltOpponent(t, g, opp.ID)
	if me.Life != life+1 || opp.Life != oppLife-4 || other.Life != otherLife-1 {
		t.Fatalf("first crime: you %d → %d, bolted %d → %d, other %d → %d", life, me.Life, oppLife, opp.Life, otherLife, other.Life)
	}
	otjB1BoltOpponent(t, g, opp.ID)
	if me.Life != life+1 || other.Life != otherLife-1 {
		t.Errorf("a second crime this turn drained again: you %d, other %d", me.Life, other.Life)
	}
}

// --- Razzle-Dazzler ----------------------------------------------------------

func TestOtjRRazzleDazzlerGrowsAndSlipsBlockersOnTheSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dazzler := effectstest.B12Push(g, me.ID, "Razzle-Dazzler", "Creature — Human Wizard", otjRRazzleDazzlerOracle, 1, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.B24CastNoncreature(t, g, "Opt", "Instant", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, dazzler); got != 0 {
		t.Fatalf("the first spell put %d counters", got)
	}
	effectstest.B24CastNoncreature(t, g, "Opt", "Instant", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, dazzler); got != 1 {
		t.Fatalf("the second spell put %d counters, want 1", got)
	}
	effectstest.B24CastNoncreature(t, g, "Opt", "Instant", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, dazzler); got != 1 {
		t.Errorf("the third spell put another counter: %d", got)
	}
	effectstest.BrAttack(t, g, dazzler)
	if err := g.DeclareBlocker(wall, dazzler); err == nil {
		t.Error("the Dazzler was blocked")
	}
}

// --- Reach for the Sky -------------------------------------------------------

func TestOtjRReachForTheSkyPumpsAndDrawsWhenItGoesToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	aura := effectstest.CastCatalogSpell(t, g, "Reach for the Sky", "Enchantment — Aura", otjRReachForTheSkyOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 5 || tt != 4 {
		t.Errorf("the enchanted Bear is %d/%d, want 5/4", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "reach") {
		t.Error("the enchanted Bear lacks reach")
	}
	hand := me.Hand.Size()
	for i := 0; i < 2; i++ {
		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(bear))
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(aura) || !me.Graveyard.Contains(aura) {
		t.Fatal("the Bear died and the Aura went to the graveyard")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// --- Reckless Lackey ---------------------------------------------------------

func TestOtjRRecklessLackeySacrificesForACardAndATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "One", "Two")
	lackey := effectstest.B12Push(g, me.ID, "Reckless Lackey", "Creature — Goblin Pirate", otjRRecklessLackeyOracle, 1, 2)
	for _, kw := range []string{"first strike", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, lackey, kw) {
			t.Errorf("%s is printed", kw)
		}
	}
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.B06AddMana(me, "R", "C", "C")
	effectstest.B16Activate(t, g, me.ID, lackey, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(lackey) {
		t.Fatal("the sacrifice is a cost, paid at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if got := otjRTreasures(g, me.ID); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
}

// --- Redrock Sentinel --------------------------------------------------------

func TestOtjRRedrockSentinelEatsALandForACardAndATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "One", "Two")
	sentinel := effectstest.PushCatalogPermanent(g, me.ID, "Redrock Sentinel", "Artifact Creature — Golem", otjRRedrockSentinelOracle, false)
	forest := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, sentinel, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a creature was accepted for 'sacrifice a land'")
	}
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, sentinel, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{forest}})
	if g.Battlefield.Contains(forest) || !me.Graveyard.Contains(forest) {
		t.Fatal("the land is sacrificed as a cost, at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, sentinel) {
		t.Error("{T} is part of the cost")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if got := otjRTreasures(g, me.ID); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, sentinel, "defender") {
		t.Error("defender is printed")
	}
}

// --- Resilient Roadrunner ----------------------------------------------------

func TestOtjRResilientRoadrunnerIsProtectedFromCoyotes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	runner := effectstest.B12Push(g, me.ID, "Resilient Roadrunner", "Creature — Bird", otjRRoadrunnerOracle, 2, 2)
	coyote := effectstest.B12Creature(g, opp.ID, "Coyote", "Creature — Coyote", 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, runner, "haste") {
		t.Error("haste is printed")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.BrAttack(t, g, runner)
	if err := g.DeclareBlocker(coyote, runner); err == nil {
		t.Error("a Coyote blocked it")
	}
	if err := g.DeclareBlocker(bear, runner); err != nil {
		t.Errorf("a Bear may block it: %v", err)
	}
}

func TestOtjRResilientRoadrunnerCanBeBlockedOnlyByHasteAfterPaying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	runner := effectstest.B12Push(g, me.ID, "Resilient Roadrunner", "Creature — Bird", otjRRoadrunnerOracle, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hasty := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Hasty Goblin", TypeLine: "Creature — Goblin", Power: 1, Toughness: 1, Keywords: []string{"haste"}})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, runner, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BrAttack(t, g, runner)
	if err := g.DeclareBlocker(bear, runner); err == nil {
		t.Error("a creature without haste blocked it")
	}
	if err := g.DeclareBlocker(hasty, runner); err != nil {
		t.Errorf("a creature with haste may block it: %v", err)
	}
}

// --- Rictus Robber -----------------------------------------------------------

func TestOtjRRictusRobberMakesAZombieOnlyIfACreatureDied(t *testing.T) {
	for _, tc := range []struct {
		name string
		died bool
		want int
	}{
		{"no creature died", false, 0},
		{"a creature died", true, 1},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		effectstest.AdvanceToMain(t, g)
		if tc.died {
			effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(theirs))
			effectstest.PassPriorityAroundTable(t, g)
		}
		effectstest.CastHoldCreature(t, g, "Rictus Robber", "Creature — Zombie Rogue", otjRRictusRobberOracle, 4, 3)
		effectstest.PassPriorityAroundTable(t, g)
		var zombies []uuid.UUID
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Zombie Rogue" && c.IsToken() && c.Controller == me.ID {
				zombies = append(zombies, c.InstanceID)
			}
		}
		if len(zombies) != tc.want {
			t.Fatalf("%s: %d Zombie Rogues, want %d", tc.name, len(zombies), tc.want)
		}
		if tc.want == 1 {
			if p, tt := effectstest.PtOf(t, g, zombies[0]); p != 2 || tt != 2 {
				t.Errorf("the Zombie Rogue is %d/%d, want 2/2", p, tt)
			}
		}
	}
}

// --- Rise of the Varmints ----------------------------------------------------

func TestOtjRRiseOfTheVarmintsCountsCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Ogre", TypeLine: "Creature — Ogre", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Shock", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Dead Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Rise of the Varmints", "Sorcery", otjRRiseOfTheVarmintsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var varmints []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Varmint" && c.IsToken() && c.Controller == me.ID {
			varmints = append(varmints, c.InstanceID)
		}
	}
	if len(varmints) != 2 {
		t.Fatalf("%d Varmints, want 2 (two creature cards in your graveyard)", len(varmints))
	}
	if p, tt := effectstest.PtOf(t, g, varmints[0]); p != 2 || tt != 1 {
		t.Errorf("the Varmint is %d/%d, want 2/1", p, tt)
	}
	if c := effectstest.LayeredCard(t, g, varmints[0]); len(c.Colors) != 1 || c.Colors[0] != "G" {
		t.Errorf("the Varmint's colours are %v, want green", c.Colors)
	}
}

// --- Rodeo Pyromancers -------------------------------------------------------

func TestOtjRRodeoPyromancersAddsTwoRedOnTheFirstSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Rodeo Pyromancers", "Creature — Human Mercenary", otjRRodeoOracle, 3, 4)
	effectstest.B24CastNoncreature(t, g, "Opt", "Instant", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PoolColors(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Fatalf("the first spell: pool %v, want R R", got)
	}
	effectstest.B24CastNoncreature(t, g, "Opt", "Instant", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PoolColors(me); len(got) != 2 {
		t.Errorf("the second spell added mana: pool %v", got)
	}
}

// --- Rooftop Assassin --------------------------------------------------------

func TestOtjRRooftopAssassinDestroysADamagedOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hurt := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	fresh := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "My Giant", "Creature — Giant", 4, 4)
	effectstest.AdvanceToMain(t, g)
	for _, id := range []uuid.UUID{hurt, mine} {
		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(id))
		effectstest.PassPriorityAroundTable(t, g)
	}
	assassin := effectstest.CastHoldCreature(t, g, "Rooftop Assassin", "Creature — Vampire Assassin", otjRRooftopAssassinOracle, 2, 2)
	for _, kw := range []string{"flash", "flying", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, assassin, kw) {
			t.Errorf("%s is printed", kw)
		}
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if !effectstest.HasID(p.PickTargetCards, hurt) || effectstest.HasID(p.PickTargetCards, fresh) || effectstest.HasID(p.PickTargetCards, mine) {
		t.Fatalf("a damaged creature an opponent controls: %v", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, hurt)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hurt) {
		t.Error("the damaged creature was not destroyed")
	}
}

// --- Rustler Rampage ---------------------------------------------------------

func TestOtjRRustlerRampageUntapsAPlayersCreaturesAndGrantsDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ogre := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{bear, ogre, land} {
			effectstest.BattlefieldCardPtr(t, g, id).Tapped = true
		}
	})
	effectstest.CastModal(t, g, "Rustler Rampage", "Instant", otjRRustlerRampageOracle,
		[]int{0, 1},
		[]game.TargetRef{
			effectstest.ModeRef(game.TargetPlayer, me.ID, 0, 0),
			effectstest.ModeRef(game.TargetCard, ogre, 1, 0),
		})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, bear) || effectstest.B16Tapped(t, g, ogre) {
		t.Error("the target player's creatures were not untapped")
	}
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("a land was untapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, ogre, "double strike") || effectstest.HasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("double strike went to the wrong creature")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, ogre, "double strike") {
		t.Error("double strike lasts until end of turn only")
	}
}

func TestOtjRRustlerRampageDoubleStrikeAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ogre := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, bear).Tapped = true })
	effectstest.CastModal(t, g, "Rustler Rampage", "Instant", otjRRustlerRampageOracle,
		[]int{1}, []game.TargetRef{effectstest.ModeRef(game.TargetCard, ogre, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("the untap bullet was not chosen")
	}
	if !effectstest.HasEffectiveKeyword(t, g, ogre, "double strike") {
		t.Error("the Ogre lacks double strike")
	}
}

// --- Ruthless Lawbringer -----------------------------------------------------

func TestOtjRRuthlessLawbringerSacrificesAnotherCreatureToDestroyANonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	effectstest.AdvanceToMain(t, g)
	law := effectstest.CastHoldCreature(t, g, "Ruthless Lawbringer", "Creature — Vampire Assassin", otjRLawbringerOracle, 3, 2)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.ChooseCards {
		offered[id] = true
	}
	if !offered[bear] || offered[law] {
		t.Errorf("offered %v: want another creature, never the Lawbringer", c.ChooseCards)
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, bear)
	effectstest.B04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear was not sacrificed before the reflexive trigger")
	}
	effectstest.DftCGPickRefused(t, g, me.ID, land)
	effectstest.PickCard(t, g, me.ID, ench)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment was not destroyed")
	}
	if !g.Battlefield.Contains(land) || !g.Battlefield.Contains(law) {
		t.Error("something else left the battlefield")
	}
}

func TestOtjRRuthlessLawbringerMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastHoldCreature(t, g, "Ruthless Lawbringer", "Creature — Vampire Assassin", otjRLawbringerOracle, 3, 2)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(ench) {
		t.Error("declining changed the battlefield")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("no sacrifice, no reflexive trigger")
	}
}
