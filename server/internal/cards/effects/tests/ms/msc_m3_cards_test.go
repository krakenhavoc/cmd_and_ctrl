package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m3_cards_test.go — Marvel Super Heroes Commander, M–Z batch 3 (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mscM3MJOracle         = "06f647a7-ee53-42bd-8c73-53a0c6ede34a"
	mscM3MedusaOracle     = "91accef6-378f-4777-ad58-fa1af7893d8a"
	mscM3MachineManOracle = "853587be-9709-43a4-81c9-5b6dabe523cf"
	mscM3MockingbirdOrcl  = "d34affa9-3629-49e1-b1e4-4742aff6eafc"
	mscM3HulkOracle       = "b8a627a1-d6d7-45d1-aeea-ff7b834cbb29"
	mscM3RadioactiveOracl = "6e901cab-fba9-4581-9f09-90473bbd8e30"
	mscM3SongbirdOracle   = "46f78ae1-fa51-4ef9-9ffa-8a3ed71a39e4"
	mscM3WolfsbaneOracle  = "23d3ff11-bf19-4315-b53b-3e4f20c6a89b"
	mscM3MeowerOracle     = "a0037ff8-64db-45dd-bba1-db5c49095083"
	mscM3NighthawkOracle  = "eb8a373a-ba15-4b3f-959a-b7c708fea616"
	mscM3WallOffOracle    = "fbde72f6-a8b2-4864-ad6d-b91aa03d290e"
)

func TestMscM3EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mscM3MJOracle:         "MJ, Rising Star",
		mscM3MedusaOracle:     "Medusa, Inhuman Queen",
		mscM3MachineManOracle: "Machine Man, Model X-51",
		mscM3MockingbirdOrcl:  "Mockingbird, Bobbi Morse",
		mscM3HulkOracle:       "Professor Hulk",
		mscM3RadioactiveOracl: "Radioactive Man",
		mscM3SongbirdOracle:   "Songbird, Sonic Screamer",
		mscM3WolfsbaneOracle:  "Wolfsbane, Highland Hero",
		mscM3MeowerOracle:     "Masked Meower",
		mscM3NighthawkOracle:  "Nighthawk, Dark Defender",
		mscM3WallOffOracle:    "Wall Off",
	}
	for oracle, name := range want {
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

func mscM3Live(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c
}

func mscM3Gain(g *game.Game, player uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, player, n) })
}

// --- MJ, Rising Star ---

func TestMscM3MJGrowsOncePerLifeGainOfHerControllerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mj := effectstest.B27Push(g, me.ID, "MJ, Rising Star", "Legendary Creature — Human Performer", mscM3MJOracle, "{2}{W}", 2, 3, "W")
	effectstest.AssertKeywords(t, g, mj, "vigilance")
	mscM3Gain(g, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, mj); got != 0 {
		t.Fatalf("an opponent's life gain gave MJ %d counters, want 0", got)
	}
	mscM3Gain(g, me.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, mj); got != 1 {
		t.Errorf("gaining 5 life in one event gave %d counters, want 1", got)
	}
	mscM3Gain(g, me.ID, -2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, mj); got != 1 {
		t.Errorf("losing life changed the counters to %d, want 1", got)
	}
}

// --- Medusa, Inhuman Queen ---

func TestMscM3MedusaGrowsOnEveryPlayersNoncreatureSpellButNotCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	medusa := effectstest.B27Push(g, me.ID, "Medusa, Inhuman Queen", "Legendary Creature — Inhuman Noble Hero", mscM3MedusaOracle, "{2}{G}", 2, 2, "G")
	effectstest.AssertKeywords(t, g, medusa, "reach", "vigilance")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFree(t, g, me, "Bear Cub", "Creature — Bear")
	if got := effectstest.PlusOnes(g, medusa); got != 0 {
		t.Fatalf("a creature spell gave Medusa %d counters, want 0", got)
	}
	effectstest.CastFree(t, g, me, "Artifact Bear", "Artifact Creature — Bear")
	if got := effectstest.PlusOnes(g, medusa); got != 0 {
		t.Fatalf("an artifact creature spell gave Medusa %d counters, want 0", got)
	}
	effectstest.CastFree(t, g, me, "Shock", "Instant")
	if got := effectstest.PlusOnes(g, medusa); got != 1 {
		t.Fatalf("her controller's instant gave %d counters, want 1", got)
	}
	effectstest.CastFree(t, g, opp, "Opposing Instant", "Instant")
	if got := effectstest.PlusOnes(g, medusa); got != 2 {
		t.Errorf("an opponent's instant left %d counters, want 2", got)
	}
}

// --- Machine Man, Model X-51 ---

func TestMscM3MachineManGrowsAndFliesOnYourNoncreatureSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mm := effectstest.B27Push(g, me.ID, "Machine Man, Model X-51", "Legendary Artifact Creature — Robot Hero", mscM3MachineManOracle, "{4}", 2, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFree(t, g, opp, "Their Instant", "Instant")
	effectstest.CastFree(t, g, me, "Bear Cub", "Creature — Bear")
	c := mscM3Live(t, g, mm)
	if effectstest.PlusOnes(g, mm) != 0 || game.HasKeyword(&c, "flying") {
		t.Fatalf("an opponent's spell or a creature spell triggered him: %d counters, keywords %v", effectstest.PlusOnes(g, mm), c.Keywords)
	}
	effectstest.CastFree(t, g, me, "Shock", "Instant")
	c = mscM3Live(t, g, mm)
	if effectstest.PlusOnes(g, mm) != 1 || !game.HasKeyword(&c, "flying") {
		t.Errorf("after my noncreature spell: %d counters, keywords %v; want 1 and flying", effectstest.PlusOnes(g, mm), c.Keywords)
	}
}

// --- Mockingbird, Bobbi Morse ---

func TestMscM3MockingbirdGrowsOnlyWhenSheSurvivesTheDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := effectstest.B27Push(g, me.ID, "Mockingbird, Bobbi Morse", "Legendary Creature — Human Hero", mscM3MockingbirdOrcl, "{3}{W}", 3, 3, "W")
	foe := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Test", 1, 9)
	effectstest.B27Damage(g, foe, bird, 2)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, bird); got != 1 {
		t.Fatalf("surviving 2 damage gave %d counters, want 1", got)
	}
	// Now 4/4 with 2 damage marked: 2 more is lethal, so no counter and she dies.
	effectstest.B27Damage(g, foe, bird, 2)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) {
		t.Fatal("Mockingbird should have died to lethal damage")
	}
	if got := effectstest.PlusOnes(g, bird); got != 0 {
		t.Errorf("a dead Mockingbird has %d counters, want 0", got)
	}
}

// --- Professor Hulk ---

func TestMscM3ProfessorHulkDrawsAsManyAsTheCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hulk := effectstest.B27Push(g, me.ID, "Professor Hulk", "Legendary Creature — Gamma Scientist Hero", mscM3HulkOracle, "{4}{U}{U}", 6, 6, "U")
	effectstest.AssertKeywords(t, g, hulk, "trample")
	before := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, hulk)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 6 {
		t.Errorf("drew %d cards for 6 combat damage, want 6", got)
	}
}

func TestMscM3ProfessorHulkDrawsNothingWhenBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hulk := effectstest.B27Push(g, me.ID, "Professor Hulk", "Legendary Creature — Gamma Scientist Hero", mscM3HulkOracle, "{4}{U}{U}", 6, 6, "U")
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(hulk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, hulk); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	before := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("drew %d cards when the 6 damage went to a blocker, want 0", got)
	}
}

// --- Radioactive Man ---

func TestMscM3RadioactiveManHalvesTheLifeOfThePlayerHeHits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rad := effectstest.B27Push(g, me.ID, "Radioactive Man", "Legendary Creature — Human Scientist Villain", mscM3RadioactiveOracl, "{4}{B}", 3, 5, "B")
	effectstest.AssertKeywords(t, g, rad, "deathtouch")
	other := g.Seats[2].Life
	effectstest.AttackWith(t, g, opp.ID, rad)
	effectstest.PassPriorityAroundTable(t, g)
	// 40 - 3 combat damage = 37; half rounded up (19) is lost.
	if opp.Life != 18 {
		t.Errorf("opponent life %d, want 18 (37 minus ceil(37/2))", opp.Life)
	}
	if g.Seats[2].Life != other {
		t.Error("a player he did not hit lost life")
	}
}

// --- Songbird, Sonic Screamer ---

func TestMscM3SongbirdDiscardsACardToFly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bird := effectstest.B27Push(g, me.ID, "Songbird, Sonic Screamer", "Legendary Creature — Human Villain", mscM3SongbirdOracle, "{1}{B}", 2, 1, "B")
	effectstest.AssertKeywords(t, g, bird, "lifelink")
	spare := effectstest.PushToHand(g, me.ID, "Spare Card", "Instant")
	if err := g.ActivateCatalogAbility(me.ID, bird, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the ability was activated without paying the discard")
	}
	if c := mscM3Live(t, g, bird); game.HasKeyword(&c, "flying") {
		t.Fatal("a refused activation still granted flying")
	}
	if err := g.ActivateCatalogAbility(me.ID, bird, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{spare}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(spare) {
		t.Error("the discarded card is still in hand")
	}
	c := mscM3Live(t, g, bird)
	if !game.HasKeyword(&c, "flying") {
		t.Errorf("keywords %v, want flying", c.Keywords)
	}
}

// --- Wolfsbane, Highland Hero ---

func TestMscM3WolfsbanePumpsOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wolf := effectstest.B27Push(g, me.ID, "Wolfsbane, Highland Hero", "Legendary Creature — Mutant Werewolf Hero", mscM3WolfsbaneOracle, "{1}{G}", 2, 2, "G")
	effectstest.AssertKeywords(t, g, wolf, "trample")
	effectstest.FloatMana(t, g, me, "{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p, tough := effectstest.EffectivePTOf(t, g, wolf)
	if p != 4 || tough != 4 {
		t.Fatalf("Wolfsbane is %d/%d, want 4/4", p, tough)
	}
	effectstest.FloatMana(t, g, me, "{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a second activation in the same turn was allowed")
	}
}

// --- Masked Meower ---

func TestMscM3MaskedMeowerDiscardsAndSacrificesItselfToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cat := effectstest.B27Push(g, me.ID, "Masked Meower", "Creature — Spider Cat Hero", mscM3MeowerOracle, "{R}", 1, 1, "R")
	effectstest.AssertKeywords(t, g, cat, "haste")
	if err := g.ActivateCatalogAbility(me.ID, cat, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without a discard")
	}
	if !g.Battlefield.Contains(cat) {
		t.Fatal("a refused activation sacrificed the Meower")
	}
	spare := effectstest.PushToHand(g, me.ID, "Spare Card", "Instant")
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, cat, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{spare}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(cat) {
		t.Error("the Meower was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The spare was in hand before (+0 net after discarding it), then one drawn.
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("hand changed by %d, want 0 (discarded the spare, drew one)", got)
	}
	if me.Hand.Contains(spare) {
		t.Error("the discarded card is still in hand")
	}
}

// --- Nighthawk, Dark Defender ---

func TestMscM3NighthawkPumpsATargetWhenHeOrAnotherHeroEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	hawk := effectstest.CastHoldCreature(t, g, "Nighthawk, Dark Defender", "Legendary Creature — Human Hero", mscM3NighthawkOracle, 2, 1)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 3 {
		t.Fatalf("after Nighthawk entered the bear has power %d, want 3", p)
	}
	effectstest.AssertKeywords(t, g, hawk, "flying")

	// Another Hero entering triggers him again; a non-Hero does not.
	effectstest.CastHoldCreature(t, g, "Plain Bear", "Creature — Bear", "", 2, 2)
	if ch := effectstest.LatestPickTarget(g, me.ID); ch != nil {
		t.Fatal("a non-Hero entering raised Nighthawk's trigger")
	}
	effectstest.CastHoldCreature(t, g, "Other Hero", "Creature — Human Hero", "", 1, 1)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("after a second Hero the bear is %d/%d, want 4/4", p, tough)
	}
}

func TestMscM3NighthawkDoesNotTriggerForAnOpponentsHero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Nighthawk, Dark Defender", "Legendary Creature — Human Hero", mscM3NighthawkOracle, "{1}{W}", 2, 1, "W")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 1, 1)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: theirs})
	})
	if ch := effectstest.LatestPickTarget(g, me.ID); ch != nil {
		t.Error("an opposing Hero entering raised Nighthawk's trigger")
	}
}

// --- Wall Off ---

func TestMscM3WallOffCostsOneLessPerOpposingCreatureAndMakesAWall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, g.Seats[2].ID, "Third Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	// Two opposing creatures reduce {3}{W} by {2}; my own bear does not count.
	if got := effectstest.SelfPricedMV(t, g, me, mscM3WallOffOracle, "Instant", "{3}{W}"); got != 2 {
		t.Fatalf("Wall Off with two opposing creatures and one of mine has mana value %d, want 2 ({1}{W})", got)
	}
	effectstest.FloatMana(t, g, me, "{W}")
	effectstest.FloatMana(t, g, me, "{C}")
	spell := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Wall Off", TypeLine: "Instant", ManaCost: "{3}{W}",
			OracleID: mscM3WallOffOracle, Colors: []string{"W"}, Owner: me.ID, Controller: me.ID})
	})
	life := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast for {1}{W}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left floating", me.ManaPool.Total())
	}
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4", life, me.Life)
	}
	tokens := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Wall" {
			tokens++
			if c.Power != 0 || c.Toughness != 4 || !game.HasKeyword(&c, "defender") {
				t.Errorf("the Wall is %d/%d with keywords %v, want 0/4 defender", c.Power, c.Toughness, c.Keywords)
			}
		}
	}
	if tokens != 1 {
		t.Errorf("%d Wall tokens, want 1", tokens)
	}
}
