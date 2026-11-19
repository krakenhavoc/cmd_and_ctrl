package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b16_cards_test.go — Innistrad Remastered sorceries, GAP batch 16
// (#5948). Each test plays the real catalog entry through the engine.

const (
	gapMurderousCompulsionOracle = "8fd8b4ef-69c8-43aa-9cbe-c46dba845ba8"
	gapDeadlyAllureOracle        = "f9cdaecd-9621-4a9a-888e-748ecfe551fd"
	gapSpiderSpawningOracle      = "da113cc9-bb2b-4af4-9c43-32c165b87363"
	gapTraverseOracle            = "c9dd2bf5-32e0-47f7-9a49-501c262d745b"
	gapTravelPreparationsOracle  = "8432729e-5fd4-4661-8b79-e41b1d03eb7b"
	gapAlchemistsGreetingOracle  = "9aaf0df7-5e2d-4f49-a8cb-66523be15ad6"
	gapMysticRetrievalOracle     = "b7cfbfa1-dd38-47ff-9138-fa6b247afc11"
	gapRiseFromTheTidesOracle    = "390b862a-2a85-44bd-832c-f5bbd3eb4ea0"
	gapAngelicPurgeOracle        = "b262a057-9ace-48b4-b418-1bb751ccb300"
	gapGatherTheTownsfolkOracle  = "ff45677d-977e-4ef5-859c-52393ed71a6c"
)

func TestGapB16EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gapMurderousCompulsionOracle: "Murderous Compulsion",
		gapDeadlyAllureOracle:        "Deadly Allure",
		gapSpiderSpawningOracle:      "Spider Spawning",
		gapTraverseOracle:            "Traverse the Ulvenwald",
		gapTravelPreparationsOracle:  "Travel Preparations",
		gapAlchemistsGreetingOracle:  "Alchemist's Greeting",
		gapMysticRetrievalOracle:     "Mystic Retrieval",
		gapRiseFromTheTidesOracle:    "Rise from the Tides",
		gapAngelicPurgeOracle:        "Angelic Purge",
		gapGatherTheTownsfolkOracle:  "Gather the Townsfolk",
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
	for oracle, madness := range map[string]string{
		gapMurderousCompulsionOracle: "{1}{B}",
		gapAlchemistsGreetingOracle:  "{1}{R}",
	} {
		if spec, _ := Lookup(oracle); spec.Madness != madness {
			t.Errorf("%s madness %q, want %q", spec.Name, spec.Madness, madness)
		}
	}
	for _, oracle := range []string{gapDeadlyAllureOracle, gapSpiderSpawningOracle, gapTravelPreparationsOracle, gapMysticRetrievalOracle} {
		if spec, _ := Lookup(oracle); len(spec.AlternativeCosts) != 1 {
			t.Errorf("%s has no flashback", spec.Name)
		}
	}
}

func TestGapB16MurderousCompulsionDestroysOnlyATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	tapped := b12Creature(g, opp.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	untapped := b12Creature(g, opp.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tapped) })
	if err := castCatalogSpellErrStrict(t, g, "Murderous Compulsion", "Sorcery", gapMurderousCompulsionOracle, cardRefs(untapped)); err == nil {
		t.Fatal("an untapped creature was accepted as the target")
	}
	castCatalogSpell(t, g, "Murderous Compulsion", "Sorcery", gapMurderousCompulsionOracle, cardRefs(tapped))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature was destroyed")
	}
}

func TestGapB16DeadlyAllureGivesDeathtouchAndMustBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	wall := reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	castCatalogSpell(t, g, "Deadly Allure", "Sorcery", gapDeadlyAllureOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, bear, "deathtouch", "haste")
	reqToBlockers(t, g, bear)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Bear must be blocked if able (Deadly Allure)." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, bear, wall); err != nil {
		t.Fatal(err)
	}
}

func TestGapB16DeadlyAllureFlashesBackForG(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	id := seedGraveyardCard(t, g, "Deadly Allure", "Sorcery", gapDeadlyAllureOracle)
	advanceToMain(t, g)
	b06AddMana(me, "B")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Targets: cardRefs(bear), Strict: true}); err == nil {
		t.Fatal("flashback was paid with {B}: it costs {G}")
	}
	me.ManaPool.EmptyPool()
	b06AddMana(me, "G")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back sorcery was not exiled")
	}
}

func TestGapB16SpiderSpawningMakesOneSpiderPerCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	pushCatalogGraveyardCard(me, "Dead Elf", "Creature — Elf", "", 1, 1)
	pushCatalogGraveyardCard(me, "Dead Bolt", "Instant", "", 0, 0)
	castCatalogSpell(t, g, "Spider Spawning", "Sorcery", gapSpiderSpawningOracle, nil)
	passPriorityAroundTable(t, g)
	spiders := battlefieldIDsNamed(g, "Spider")
	if len(spiders) != 2 {
		t.Fatalf("%d Spiders, want 2 (one per creature card)", len(spiders))
	}
	assertKeywords(t, g, spiders[0], "reach")
	if effectivePower(t, g, spiders[0]) != 1 || effectiveToughness(t, g, spiders[0]) != 2 {
		t.Error("a Spider is not 1/2")
	}
}

func TestGapB16SpiderSpawningWithNoCreatureCardsMakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Spider Spawning", "Sorcery", gapSpiderSpawningOracle, nil)
	passPriorityAroundTable(t, g)
	if n := len(battlefieldIDsNamed(g, "Spider")); n != 0 {
		t.Errorf("%d Spiders from an empty graveyard", n)
	}
}

func TestGapB16TraverseTheUlvenwaldFetchesABasicLandThenACreatureOrLandWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact"},
	)
	castCatalogSpell(t, g, "Traverse the Ulvenwald", "Sorcery", gapTraverseOracle, nil)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	if !b02bHandHasNamed(me, "Forest") {
		t.Fatal("one basic land in the library: it goes straight to hand")
	}
	if b02bHandHasNamed(me, "Bear") || b02bHandHasNamed(me, "Sol Ring") {
		t.Fatal("a nonland card was fetched without delirium")
	}

	for _, c := range []game.Card{
		{Name: "Dead Bear", TypeLine: "Creature — Bear"},
		{Name: "Dead Rock", TypeLine: "Artifact"},
		{Name: "Dead Bolt", TypeLine: "Instant"},
		{Name: "Dead Land", TypeLine: "Land"},
	} {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
		me.Graveyard.PushTop(c)
	}
	castCatalogSpell(t, g, "Traverse the Ulvenwald", "Sorcery", gapTraverseOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("with delirium the searcher chooses between a creature and a land")
	}
	if searchOptionNamed(g, c, "Sol Ring") != uuid.Nil {
		t.Error("an artifact is offered: it is a creature or land search")
	}
	answerSearchNamed(t, g, me.ID, "Bear")
	if !b02bHandHasNamed(me, "Bear") {
		t.Error("the chosen creature card did not reach the hand")
	}
}

func TestGapB16TravelPreparationsPutsACounterOnEachOfUpToTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	c := b12Creature(g, me.ID, "Third", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Travel Preparations", "Sorcery", gapTravelPreparationsOracle, cardRefs(a, b))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := plusOneCounters(g, id); got != 1 {
			t.Errorf("a target has %d +1/+1 counters, want 1", got)
		}
	}
	if got := plusOneCounters(g, c); got != 0 {
		t.Errorf("a non-target has %d +1/+1 counters", got)
	}
}

func TestGapB16TravelPreparationsFlashesBackForW(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	id := seedGraveyardCard(t, g, "Travel Preparations", "Sorcery", gapTravelPreparationsOracle)
	advanceToMain(t, g)
	b06AddMana(me, "W", "W")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Targets: cardRefs(a), Strict: true}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, a); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back sorcery was not exiled")
	}
}

func TestGapB16AlchemistsGreetingDealsFourToACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 4)
	big := b12Creature(g, opp.ID, "Big", "Creature — Beast", 5, 5)
	castCatalogSpell(t, g, "Alchemist's Greeting", "Sorcery", gapAlchemistsGreetingOracle, cardRefs(small))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a 2/4 survived 4 damage")
	}
	castCatalogSpell(t, g, "Alchemist's Greeting", "Sorcery", gapAlchemistsGreetingOracle, cardRefs(big))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big) {
		t.Error("a 5/5 died to 4 damage")
	}
	if err := castCatalogSpellErrStrict(t, g, "Alchemist's Greeting", "Sorcery", gapAlchemistsGreetingOracle, []game.TargetRef{playerTarget(opp.ID)}); err == nil {
		t.Error("a player was accepted: it targets a creature")
	}
}

func TestGapB16MysticRetrievalReturnsAnInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bolt := pushCatalogGraveyardCard(me, "Dead Bolt", "Instant", "", 0, 0)
	bear := pushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	if err := castCatalogSpellErrStrict(t, g, "Mystic Retrieval", "Sorcery", gapMysticRetrievalOracle, cardRefs(bear)); err == nil {
		t.Fatal("a creature card was accepted")
	}
	castCatalogSpell(t, g, "Mystic Retrieval", "Sorcery", gapMysticRetrievalOracle, cardRefs(bolt))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the instant card did not return to hand")
	}
}

func TestGapB16RiseFromTheTidesMakesTappedZombiesPerInstantAndSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogGraveyardCard(me, "Dead Bolt", "Instant", "", 0, 0)
	pushCatalogGraveyardCard(me, "Dead Wrath", "Sorcery", "", 0, 0)
	pushCatalogGraveyardCard(me, "Dead Arcane", "Instant — Arcane", "", 0, 0)
	pushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Rise from the Tides", "Sorcery", gapRiseFromTheTidesOracle, nil)
	passPriorityAroundTable(t, g)
	zombies := battlefieldIDsNamed(g, "Zombie")
	if len(zombies) != 3 {
		t.Fatalf("%d Zombies, want 3", len(zombies))
	}
	for _, z := range zombies {
		if !tappedForTest(t, g, z) {
			t.Error("a Zombie entered untapped")
		}
		if effectivePower(t, g, z) != 2 || effectiveToughness(t, g, z) != 2 {
			t.Error("a Zombie is not 2/2")
		}
	}
}

func TestGapB16AngelicPurgeSacrificesAPermanentAndExilesTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := b12Permanent(g, me.ID, "Fodder", "Artifact")
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Angelic Purge", TypeLine: "Sorcery", OracleID: gapAngelicPurgeOracle,
		ManaCost: "{2}{W}", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	b06AddMana(me, "W", "W", "W")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(victim), Strict: true}); err == nil {
		t.Fatal("cast without sacrificing a permanent")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(land), SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err == nil {
		t.Fatal("a land was accepted as the target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(victim), SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed permanent is still out")
	}
	if g.Battlefield.Contains(victim) || !g.Exile.Contains(victim) {
		t.Error("the target was not exiled")
	}
}

func TestGapB16GatherTheTownsfolkMakesTwoOrFiveByLife(t *testing.T) {
	for _, tc := range []struct{ life, want int }{{20, 2}, {6, 2}, {5, 5}, {1, 5}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		g.WithWriteLock(func() { me.Life = tc.life })
		castCatalogSpell(t, g, "Gather the Townsfolk", "Sorcery", gapGatherTheTownsfolkOracle, nil)
		passPriorityAroundTable(t, g)
		humans := battlefieldIDsNamed(g, "Human")
		if len(humans) != tc.want {
			t.Errorf("life %d: %d Human tokens, want %d", tc.life, len(humans), tc.want)
			continue
		}
		if effectivePower(t, g, humans[0]) != 1 || controllerOf(t, g, humans[0]) != me.ID {
			t.Errorf("life %d: want 1/1 Humans under your control", tc.life)
		}
	}
}
