package effects

import (
	"testing"

	"github.com/google/uuid"

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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mj := b27Push(g, me.ID, "MJ, Rising Star", "Legendary Creature — Human Performer", mscM3MJOracle, "{2}{W}", 2, 3, "W")
	assertKeywords(t, g, mj, "vigilance")
	mscM3Gain(g, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, mj); got != 0 {
		t.Fatalf("an opponent's life gain gave MJ %d counters, want 0", got)
	}
	mscM3Gain(g, me.ID, 5)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, mj); got != 1 {
		t.Errorf("gaining 5 life in one event gave %d counters, want 1", got)
	}
	mscM3Gain(g, me.ID, -2)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, mj); got != 1 {
		t.Errorf("losing life changed the counters to %d, want 1", got)
	}
}

// --- Medusa, Inhuman Queen ---

func TestMscM3MedusaGrowsOnEveryPlayersNoncreatureSpellButNotCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	medusa := b27Push(g, me.ID, "Medusa, Inhuman Queen", "Legendary Creature — Inhuman Noble Hero", mscM3MedusaOracle, "{2}{G}", 2, 2, "G")
	assertKeywords(t, g, medusa, "reach", "vigilance")
	advanceToMain(t, g)
	castFree(t, g, me, "Bear Cub", "Creature — Bear")
	if got := plusOnes(g, medusa); got != 0 {
		t.Fatalf("a creature spell gave Medusa %d counters, want 0", got)
	}
	castFree(t, g, me, "Artifact Bear", "Artifact Creature — Bear")
	if got := plusOnes(g, medusa); got != 0 {
		t.Fatalf("an artifact creature spell gave Medusa %d counters, want 0", got)
	}
	castFree(t, g, me, "Shock", "Instant")
	if got := plusOnes(g, medusa); got != 1 {
		t.Fatalf("her controller's instant gave %d counters, want 1", got)
	}
	castFree(t, g, opp, "Opposing Instant", "Instant")
	if got := plusOnes(g, medusa); got != 2 {
		t.Errorf("an opponent's instant left %d counters, want 2", got)
	}
}

// --- Machine Man, Model X-51 ---

func TestMscM3MachineManGrowsAndFliesOnYourNoncreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mm := b27Push(g, me.ID, "Machine Man, Model X-51", "Legendary Artifact Creature — Robot Hero", mscM3MachineManOracle, "{4}", 2, 3)
	advanceToMain(t, g)
	castFree(t, g, opp, "Their Instant", "Instant")
	castFree(t, g, me, "Bear Cub", "Creature — Bear")
	c := mscM3Live(t, g, mm)
	if plusOnes(g, mm) != 0 || game.HasKeyword(&c, "flying") {
		t.Fatalf("an opponent's spell or a creature spell triggered him: %d counters, keywords %v", plusOnes(g, mm), c.Keywords)
	}
	castFree(t, g, me, "Shock", "Instant")
	c = mscM3Live(t, g, mm)
	if plusOnes(g, mm) != 1 || !game.HasKeyword(&c, "flying") {
		t.Errorf("after my noncreature spell: %d counters, keywords %v; want 1 and flying", plusOnes(g, mm), c.Keywords)
	}
}

// --- Mockingbird, Bobbi Morse ---

func TestMscM3MockingbirdGrowsOnlyWhenSheSurvivesTheDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := b27Push(g, me.ID, "Mockingbird, Bobbi Morse", "Legendary Creature — Human Hero", mscM3MockingbirdOrcl, "{3}{W}", 3, 3, "W")
	foe := b12Creature(g, opp.ID, "Pinger", "Creature — Test", 1, 9)
	b27Damage(g, foe, bird, 2)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bird); got != 1 {
		t.Fatalf("surviving 2 damage gave %d counters, want 1", got)
	}
	// Now 4/4 with 2 damage marked: 2 more is lethal, so no counter and she dies.
	b27Damage(g, foe, bird, 2)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) {
		t.Fatal("Mockingbird should have died to lethal damage")
	}
	if got := plusOnes(g, bird); got != 0 {
		t.Errorf("a dead Mockingbird has %d counters, want 0", got)
	}
}

// --- Professor Hulk ---

func TestMscM3ProfessorHulkDrawsAsManyAsTheCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hulk := b27Push(g, me.ID, "Professor Hulk", "Legendary Creature — Gamma Scientist Hero", mscM3HulkOracle, "{4}{U}{U}", 6, 6, "U")
	assertKeywords(t, g, hulk, "trample")
	before := me.Hand.Size()
	attackWith(t, g, opp.ID, hulk)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 6 {
		t.Errorf("drew %d cards for 6 combat damage, want 6", got)
	}
}

func TestMscM3ProfessorHulkDrawsNothingWhenBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hulk := b27Push(g, me.ID, "Professor Hulk", "Legendary Creature — Gamma Scientist Hero", mscM3HulkOracle, "{4}{U}{U}", 6, 6, "U")
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(hulk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, hulk); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	before := me.Hand.Size()
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("drew %d cards when the 6 damage went to a blocker, want 0", got)
	}
}

// --- Radioactive Man ---

func TestMscM3RadioactiveManHalvesTheLifeOfThePlayerHeHits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rad := b27Push(g, me.ID, "Radioactive Man", "Legendary Creature — Human Scientist Villain", mscM3RadioactiveOracl, "{4}{B}", 3, 5, "B")
	assertKeywords(t, g, rad, "deathtouch")
	other := g.Seats[2].Life
	attackWith(t, g, opp.ID, rad)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	bird := b27Push(g, me.ID, "Songbird, Sonic Screamer", "Legendary Creature — Human Villain", mscM3SongbirdOracle, "{1}{B}", 2, 1, "B")
	assertKeywords(t, g, bird, "lifelink")
	spare := pushToHand(g, me.ID, "Spare Card", "Instant")
	if err := g.ActivateCatalogAbility(me.ID, bird, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the ability was activated without paying the discard")
	}
	if c := mscM3Live(t, g, bird); game.HasKeyword(&c, "flying") {
		t.Fatal("a refused activation still granted flying")
	}
	if err := g.ActivateCatalogAbility(me.ID, bird, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{spare}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	wolf := b27Push(g, me.ID, "Wolfsbane, Highland Hero", "Legendary Creature — Mutant Werewolf Hero", mscM3WolfsbaneOracle, "{1}{G}", 2, 2, "G")
	assertKeywords(t, g, wolf, "trample")
	floatMana(t, g, me, "{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	p, tough := effectivePTOf(t, g, wolf)
	if p != 4 || tough != 4 {
		t.Fatalf("Wolfsbane is %d/%d, want 4/4", p, tough)
	}
	floatMana(t, g, me, "{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a second activation in the same turn was allowed")
	}
}

// --- Masked Meower ---

func TestMscM3MaskedMeowerDiscardsAndSacrificesItselfToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cat := b27Push(g, me.ID, "Masked Meower", "Creature — Spider Cat Hero", mscM3MeowerOracle, "{R}", 1, 1, "R")
	assertKeywords(t, g, cat, "haste")
	if err := g.ActivateCatalogAbility(me.ID, cat, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated without a discard")
	}
	if !g.Battlefield.Contains(cat) {
		t.Fatal("a refused activation sacrificed the Meower")
	}
	spare := pushToHand(g, me.ID, "Spare Card", "Instant")
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, cat, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{spare}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(cat) {
		t.Error("the Meower was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	hawk := castHoldCreature(t, g, "Nighthawk, Dark Defender", "Legendary Creature — Human Hero", mscM3NighthawkOracle, 2, 1)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, bear); p != 3 {
		t.Fatalf("after Nighthawk entered the bear has power %d, want 3", p)
	}
	assertKeywords(t, g, hawk, "flying")

	// Another Hero entering triggers him again; a non-Hero does not.
	castHoldCreature(t, g, "Plain Bear", "Creature — Bear", "", 2, 2)
	if ch := latestPickTarget(g, me.ID); ch != nil {
		t.Fatal("a non-Hero entering raised Nighthawk's trigger")
	}
	castHoldCreature(t, g, "Other Hero", "Creature — Human Hero", "", 1, 1)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("after a second Hero the bear is %d/%d, want 4/4", p, tough)
	}
}

func TestMscM3NighthawkDoesNotTriggerForAnOpponentsHero(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Nighthawk, Dark Defender", "Legendary Creature — Human Hero", mscM3NighthawkOracle, "{1}{W}", 2, 1, "W")
	theirs := b12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 1, 1)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: theirs})
	})
	if ch := latestPickTarget(g, me.ID); ch != nil {
		t.Error("an opposing Hero entering raised Nighthawk's trigger")
	}
}

// --- Wall Off ---

func TestMscM3WallOffCostsOneLessPerOpposingCreatureAndMakesAWall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12Creature(g, g.Seats[2].ID, "Third Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	// Two opposing creatures reduce {3}{W} by {2}; my own bear does not count.
	if got := selfPricedMV(t, g, me, mscM3WallOffOracle, "Instant", "{3}{W}"); got != 2 {
		t.Fatalf("Wall Off with two opposing creatures and one of mine has mana value %d, want 2 ({1}{W})", got)
	}
	floatMana(t, g, me, "{W}")
	floatMana(t, g, me, "{C}")
	spell := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Wall Off", TypeLine: "Instant", ManaCost: "{3}{W}",
			OracleID: mscM3WallOffOracle, Colors: []string{"W"}, Owner: me.ID, Controller: me.ID})
	})
	life := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast for {1}{W}: %v", err)
	}
	passPriorityAroundTable(t, g)
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
